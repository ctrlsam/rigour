package fingerprint

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
)

var dialer = &net.Dialer{
	Timeout: 2 * time.Second,
}

var sortedTCPPlugins = make([]plugins.Plugin, 0)
var sortedTCPTLSPlugins = make([]plugins.Plugin, 0)
var sortedUDPPlugins = make([]plugins.Plugin, 0)
var tlsConfig = tls.Config{}

func init() {
	setupPlugins()
	cipherSuites := make([]uint16, 0)

	for _, suite := range tls.CipherSuites() {
		cipherSuites = append(cipherSuites, suite.ID)
	}

	for _, suite := range tls.InsecureCipherSuites() {
		cipherSuites = append(cipherSuites, suite.ID)
	}
	tlsConfig.InsecureSkipVerify = true
	tlsConfig.CipherSuites = cipherSuites
	tlsConfig.MinVersion = tls.VersionTLS10
}

func setupPlugins() {
	if len(sortedTCPPlugins) > 0 {
		return // already sorted
	}

	sortedTCPPlugins = append(sortedTCPPlugins, plugins.Plugins[plugins.TCP]...)
	sortedTCPTLSPlugins = append(sortedTCPTLSPlugins, plugins.Plugins[plugins.TCPTLS]...)
	sortedUDPPlugins = append(sortedUDPPlugins, plugins.Plugins[plugins.UDP]...)

	sort.Slice(sortedTCPPlugins, func(i, j int) bool {
		return sortedTCPPlugins[i].Priority() < sortedTCPPlugins[j].Priority()
	})
	sort.Slice(sortedUDPPlugins, func(i, j int) bool {
		return sortedUDPPlugins[i].Priority() < sortedUDPPlugins[j].Priority()
	})
	sort.Slice(sortedTCPTLSPlugins, func(i, j int) bool {
		return sortedTCPTLSPlugins[i].Priority() < sortedTCPTLSPlugins[j].Priority()
	})
}

// ScanTarget attempts to identify the service that is running on a given
// port. The fingerprinter supports two modes of operation referred to as the
// fast lane and slow lane. The fast lane aims to be as fast as possible and
// only attempts to fingerprint services by mapping them to their default port.
// The slow lane isn't as focused on performance and instead tries to be as
// accurate as possible.
func (c *FingerprintConfig) FingerprintTarget(target plugins.Target) (*plugins.Service, error) {
	ip := target.Address.Addr().String()
	port := target.Address.Port()
	var pluginsChecked int

	logMatch := func(result *plugins.Service) {
		if c.Verbose {
			log.Printf("[fingerprint] %s:%d matched %s after %d probes\n",
				ip, port, result.Protocol, pluginsChecked)
		}
	}

	// Fast lane: plugins whose default port matches
	for _, plugin := range sortedTCPPlugins {
		if !c.pluginAllowed(plugin.Name()) {
			continue
		}
		if plugin.PortPriority(port) {
			conn, err := DialTCP(ip, port)
			if err != nil {
				return nil, nil
			}
			pluginsChecked++
			result, err := simplePluginRunner(conn, target, c, plugin)
			if err != nil && c.Verbose {
				log.Printf("error: %v scanning %v\n", err, target.Address.String())
			}
			if result != nil && err == nil {
				logMatch(result)
				return result, nil
			}
		}
	}

	tlsConn, tlsErr := DialTLS(target)
	isTLS := tlsErr == nil
	if isTLS {
		for _, plugin := range sortedTCPTLSPlugins {
			if !c.pluginAllowed(plugin.Name()) {
				continue
			}
			if plugin.PortPriority(port) {
				pluginsChecked++
				result, err := simplePluginRunner(tlsConn, target, c, plugin)
				if err != nil && c.Verbose {
					log.Printf("error: %v scanning %v\n", err, target.Address.String())
				}
				if result != nil && err == nil {
					logMatch(result)
					return result, nil
				}
				tlsConn, err = DialTLS(target)
				if err != nil {
					break
				}
			}
		}
	}

	// Fast mode: only probe plugins where port matches the plugin's default port
	if c.FastMode {
		if c.Verbose {
			log.Printf("[fingerprint] %s:%d unidentified after %d probes (fast mode)\n", ip, port, pluginsChecked)
		}
		return nil, nil
	}

	// Slow lane: try every allowed plugin that wasn't already tried in the fast lane
	// (fast lane already handled plugins where PortPriority(port)==true).
	if isTLS {
		for _, plugin := range sortedTCPTLSPlugins {
			if !c.pluginAllowed(plugin.Name()) || plugin.PortPriority(port) {
				continue
			}
			tlsConn, err := DialTLS(target)
			if err != nil {
				break
			}
			pluginsChecked++
			result, err := simplePluginRunner(tlsConn, target, c, plugin)
			if err != nil && c.Verbose {
				log.Printf("error: %v scanning %v\n", err, target.Address.String())
			}
			if result != nil && err == nil {
				logMatch(result)
				return result, nil
			}
		}
	} else {
		// Attempt a passive banner read on a short-lived connection. Many server-first
		// protocols (SSH, FTP, SMTP, Redis, MySQL, VNC, …) send a greeting immediately,
		// which lets us route directly to the right plugin without trying all of them.
		bannerTimeout := min(300*time.Millisecond, c.DefaultTimeout/3)
		var banner []byte
		if bannerConn, err := DialTCP(ip, port); err == nil {
			banner = readBanner(bannerConn, bannerTimeout)
			bannerConn.Close()
		}

		if len(banner) > 0 {
			for _, plugin := range sortedTCPPlugins {
				if !c.pluginAllowed(plugin.Name()) || plugin.PortPriority(port) {
					continue
				}
				bm, ok := plugin.(plugins.BannerMatcher)
				if !ok || !bm.MatchBanner(banner) {
					continue
				}
				conn, err := DialTCP(ip, port)
				if err != nil {
					return nil, nil
				}
				pluginsChecked++
				result, err := simplePluginRunner(conn, target, c, plugin)
				if err != nil && c.Verbose {
					log.Printf("error: %v scanning %v\n", err, target.Address.String())
				}
				if result != nil && err == nil {
					logMatch(result)
					return result, nil
				}
				// Banner matched a plugin but run failed — fall through to full probe.
				break
			}
		}

		// Active-probe fallback: plugins that must send data first (client-first protocols),
		// or banner-matchers that were already tried above.
		for _, plugin := range sortedTCPPlugins {
			if !c.pluginAllowed(plugin.Name()) || plugin.PortPriority(port) {
				continue
			}
			// Skip banner-matched plugins — already handled above.
			if _, ok := plugin.(plugins.BannerMatcher); ok && len(banner) > 0 {
				continue
			}
			conn, err := DialTCP(ip, port)
			if err != nil {
				return nil, nil
			}
			pluginsChecked++
			result, err := simplePluginRunner(conn, target, c, plugin)
			if err != nil && c.Verbose {
				log.Printf("error: %v scanning %v\n", err, target.Address.String())
			}
			if result != nil && err == nil {
				logMatch(result)
				return result, nil
			}
		}
	}

	if c.Verbose {
		log.Printf("[fingerprint] %s:%d unidentified after %d probes\n", ip, port, pluginsChecked)
	}
	return nil, nil
}

// UDP Scan of the target
func (c *FingerprintConfig) UDPScanTarget(target plugins.Target) (*plugins.Service, error) {
	// first check the default port mappings for TCP / TLS
	for _, plugin := range sortedUDPPlugins {
		if !c.pluginAllowed(plugin.Name()) {
			continue
		}
		ip := target.Address.Addr().String()
		port := target.Address.Port()
		if plugin.PortPriority(port) {
			conn, err := DialUDP(ip, port)
			if err != nil {
				return nil, fmt.Errorf("unable to connect, err = %w", err)
			}
			result, err := simplePluginRunner(conn, target, c, plugin)
			if err != nil && c.Verbose {
				log.Printf("error: %v scanning %v\n", err, target.Address.String())
			}
			if result != nil && err == nil {
				return result, nil
			}
		}
	}

	// if we're fast mode, return (because fast mode only checks the default port service mapping)
	if c.FastMode {
		return nil, nil
	}

	for _, plugin := range sortedUDPPlugins {
		if !c.pluginAllowed(plugin.Name()) {
			continue
		}
		conn, err := DialUDP(target.Address.Addr().String(), target.Address.Port())
		if err != nil {
			return nil, nil
		}
		result, err := simplePluginRunner(conn, target, c, plugin)
		if result != nil && err == nil {
			return result, nil
		}
	}
	return nil, nil
}

// This will attempt to close the provided Conn after running the plugin.
func simplePluginRunner(
	conn net.Conn,
	target plugins.Target,
	config *FingerprintConfig,
	plugin plugins.Plugin,
) (*plugins.Service, error) {
	// Log probe start.
	if config.Verbose {
		log.Printf("%v %v-> scanning %v\n",
			target.Address.String(),
			target.Host,
			plugins.CreatePluginID(plugin),
		)
	}

	result, err := plugin.Run(conn, config.DefaultTimeout, target)

	// Log probe completion.
	if config.Verbose {
		log.Printf(
			"%v %v-> completed %v\n",
			target.Address.String(),
			target.Host,
			plugins.CreatePluginID(plugin),
		)
	}
	return result, err
}

// SortedTCPPlugins returns the TCP plugin list (sorted by priority) for inspection.
func SortedTCPPlugins() []plugins.Plugin { return sortedTCPPlugins }

// SortedTLSPlugins returns the TCP+TLS plugin list for inspection.
func SortedTLSPlugins() []plugins.Plugin { return sortedTCPTLSPlugins }

// SortedUDPPlugins returns the UDP plugin list for inspection.
func SortedUDPPlugins() []plugins.Plugin { return sortedUDPPlugins }

func DialTLS(target plugins.Target) (net.Conn, error) {
	config := &tlsConfig
	if target.Host != "" {
		// make a new config clone to add the custom host for each new tls connection
		c := config.Clone()
		c.ServerName = target.Host
		config = c
	}
	return tls.DialWithDialer(dialer, "tcp", target.Address.String(), config)
}

func DialTCP(ip string, port uint16) (net.Conn, error) {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	return dialer.Dial("tcp", addr)
}

// readBanner attempts a short passive read on conn. Returns whatever bytes the
// server sends within timeout. The caller is responsible for closing conn.
func readBanner(conn net.Conn, timeout time.Duration) []byte {
	conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 512)
	n, _ := conn.Read(buf)
	return buf[:n]
}

func DialUDP(ip string, port uint16) (net.Conn, error) {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	return dialer.Dial("udp", addr)
}
