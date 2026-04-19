package docker

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const DOCKER = "docker"

type DOCKERPlugin struct{}

func init() {
	plugins.RegisterPlugin(&DOCKERPlugin{})
}

func (p *DOCKERPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	ip := target.Address.Addr().String()
	req := fmt.Sprintf("GET /_ping HTTP/1.0\r\nHost: %s\r\n\r\n", ip)
	response, err := utils.SendRecv(conn, []byte(req), timeout)
	if err != nil {
		return nil, err
	}
	s := string(response)
	if !strings.Contains(s, "200") {
		return nil, nil
	}
	// Require at least one Docker-specific header to avoid false positives
	hasDockerHeader := false
	payload := plugins.ServiceDockerAPI{}
	for _, line := range strings.Split(s, "\r\n") {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "server: docker/") {
			hasDockerHeader = true
			// "Server: Docker/24.0.5 (linux)" → version = "24.0.5"
			rest := line[len("Server: Docker/"):]
			if idx := strings.Index(rest, " "); idx > 0 {
				payload.Version = rest[:idx]
			} else {
				payload.Version = strings.TrimSpace(rest)
			}
		}
		if strings.HasPrefix(lower, "docker-experimental:") || strings.HasPrefix(lower, "api-version:") {
			hasDockerHeader = true
		}
		if strings.HasPrefix(lower, "ostype: ") {
			payload.OSType = strings.TrimSpace(line[len("ostype: "):])
		}
	}
	if !hasDockerHeader {
		return nil, nil
	}
	return plugins.CreateServiceFrom(target, payload, false, payload.Version, plugins.TCP), nil
}

func (p *DOCKERPlugin) PortPriority(port uint16) bool {
	return port == 2375
}

func (p *DOCKERPlugin) Name() string           { return DOCKER }
func (p *DOCKERPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *DOCKERPlugin) Priority() int          { return 450 }
