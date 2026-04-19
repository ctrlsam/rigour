package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ctrlsam/rigour/internal/metrics"
	"github.com/ctrlsam/rigour/pkg/crawler/discovery"
	"github.com/ctrlsam/rigour/pkg/crawler/discovery/naabu"
	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint"
	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	"github.com/ctrlsam/rigour/pkg/types"
)

// ScanTargetWithDiscoveryStream runs discovery and fingerprinting and invokes onEvent
// as soon as a service is identified.
//
// Fingerprinting runs concurrently with a configurable worker pool to avoid blocking
// port discovery. The onEvent callback must be concurrency-safe.
func ScanTargetWithDiscoveryStream(
	ctx context.Context,
	ipRange string,
	cfg discovery.DiscoveryConfig,
	scanCfg fingerprint.FingerprintConfig,
	onEvent func(types.Service),
) error {

	if strings.TrimSpace(ipRange) == "" {
		return fmt.Errorf("target is empty")
	}
	if onEvent == nil {
		return fmt.Errorf("onEvent callback is nil")
	}

	numWorkers := scanCfg.Workers
	if numWorkers <= 0 {
		numWorkers = fingerprint.DefaultWorkers
	}
	queueCap := scanCfg.QueueSize
	if queueCap <= 0 {
		queueCap = fingerprint.DefaultQueueSize
	}

	// Total addressable IPs in the target range, for progress display.
	var cidrSize int64
	if prefix, err := netip.ParsePrefix(ipRange); err == nil {
		bits := prefix.Addr().BitLen() - prefix.Bits()
		if bits >= 0 && bits < 63 {
			cidrSize = int64(1) << bits
		}
	} else if _, err := netip.ParseAddr(ipRange); err == nil {
		cidrSize = 1
	}

	pluginCounts := countAllowedPlugins(&scanCfg)
	mode := "slow"
	if scanCfg.FastMode {
		mode = "fast"
	}
	slog.Info("scan start",
		"cidr", ipRange,
		"workers", numWorkers,
		"queue_cap", queueCap,
		"tcp_plugins", pluginCounts.tcp,
		"tls_plugins", pluginCounts.tls,
		"mode", mode,
	)

	portQueue := make(chan discovery.Result, queueCap)

	var (
		wg              sync.WaitGroup
		activeWorkers   atomic.Int32
		totalDiscovered atomic.Int64 // ports reported by naabu
		totalIdentified atomic.Int64
		totalUnknown    atomic.Int64
		totalInline     atomic.Int64
		totalUniqueIPs  atomic.Int64 // unique hosts with at least one open port

		// rate tracking — swapped atomically each tick
		prevDiscovered    atomic.Int64
		prevFingerprinted atomic.Int64

		seenIPs sync.Map
	)
	prevTickTime := time.Now()

	// Periodic stats — exits on context cancel or explicit close.
	statsDone := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				now := time.Now()
				elapsed := now.Sub(prevTickTime).Seconds()
				prevTickTime = now

				curDisc := totalDiscovered.Load()
				curFp := totalIdentified.Load() + totalUnknown.Load()
				portsPerSec := float64(curDisc-prevDiscovered.Swap(curDisc)) / elapsed
				svcsPerSec := float64(curFp-prevFingerprinted.Swap(curFp)) / elapsed

				slog.Info("scan progress",
					"cidr", ipRange,
					"ips", fmt.Sprintf("%d/%d", totalUniqueIPs.Load(), cidrSize),
					"fp", fmt.Sprintf("%d/%d", curFp, curDisc),
					"queue", fmt.Sprintf("%d/%d", len(portQueue), queueCap),
					"workers", fmt.Sprintf("%d/%d", activeWorkers.Load(), numWorkers),
					"identified", totalIdentified.Load(),
					"unknown", totalUnknown.Load(),
					"inline", totalInline.Load(),
					"ports_per_sec", fmt.Sprintf("%.1f", portsPerSec),
					"svcs_per_sec", fmt.Sprintf("%.1f", svcsPerSec),
				)
			case <-ctx.Done():
				return
			case <-statsDone:
				return
			}
		}
	}()

	fingerprintAndEmit := func(r discovery.Result) {
		activeWorkers.Add(1)
		metrics.FingerprintActiveWorkers.Inc()
		defer func() {
			activeWorkers.Add(-1)
			metrics.FingerprintActiveWorkers.Dec()
		}()

		addr := netip.AddrPortFrom(netip.MustParseAddr(r.Host), uint16(r.Port))
		t := plugins.Target{Address: addr}

		svc, _ := scanCfg.FingerprintTarget(t)
		if svc != nil {
			out := types.FromPluginService(svc, time.Now())
			slog.Info("service identified",
				"host", r.Host,
				"port", r.Port,
				"transport", out.Transport,
				"protocol", out.Protocol,
			)
			totalIdentified.Add(1)
			metrics.ServicesIdentified.WithLabelValues(out.Protocol).Inc()
			onEvent(*out)
		} else {
			slog.Info("service unknown", "host", r.Host, "port", r.Port)
			totalUnknown.Add(1)
			metrics.ServicesUnknown.Inc()
			onEvent(types.Service{
				IP:        r.Host,
				Port:      r.Port,
				Protocol:  "unknown",
				Transport: strings.ToLower(r.Protocol),
				LastScan:  time.Now(),
			})
		}
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { recover() }()
			for r := range portQueue {
				metrics.FingerprintQueueDepth.Dec()
				fingerprintAndEmit(r)
			}
		}()
	}

	onOpen := func(r discovery.Result) {
		// discard ipv6 results
		if strings.Contains(r.Host, ":") {
			return
		}

		totalDiscovered.Add(1)
		if _, loaded := seenIPs.LoadOrStore(r.Host, struct{}{}); !loaded {
			totalUniqueIPs.Add(1)
		}
		slog.Debug("port open", "host", r.Host, "port", r.Port)
		metrics.PortsDiscovered.Inc()

		select {
		case portQueue <- r:
			metrics.FingerprintQueueDepth.Inc()
		default:
			slog.Warn("fingerprint queue full, running inline",
				"queue", len(portQueue),
				"cap", queueCap,
				"host", r.Host,
				"port", r.Port,
			)
			totalInline.Add(1)
			metrics.FingerprintInlineTotal.Inc()
			fingerprintAndEmit(r)
		}
	}

	err := naabu.Run(ctx, ipRange, discovery.DiscoveryConfig{
		ScanType: cfg.ScanType,
		Ports:    cfg.Ports,
		TopPorts: cfg.TopPorts,
		Retries:  cfg.Retries,
		Rate:     cfg.Rate,
	}, onOpen)

	close(portQueue)
	close(statsDone)
	wg.Wait()

	slog.Info("scan complete",
		"cidr", ipRange,
		"identified", totalIdentified.Load(),
		"unknown", totalUnknown.Load(),
		"total", totalIdentified.Load()+totalUnknown.Load(),
	)
	return err
}

type pluginCount struct {
	tcp int
	tls int
	udp int
}

func countAllowedPlugins(cfg *fingerprint.FingerprintConfig) pluginCount {
	counts := pluginCount{}
	for _, p := range fingerprint.SortedTCPPlugins() {
		if cfg.PluginAllowed(p.Name()) {
			counts.tcp++
		}
	}
	for _, p := range fingerprint.SortedTLSPlugins() {
		if cfg.PluginAllowed(p.Name()) {
			counts.tls++
		}
	}
	for _, p := range fingerprint.SortedUDPPlugins() {
		if cfg.PluginAllowed(p.Name()) {
			counts.udp++
		}
	}
	return counts
}
