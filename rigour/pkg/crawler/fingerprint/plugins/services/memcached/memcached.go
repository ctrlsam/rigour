package memcached

import (
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const MEMCACHED = "memcached"

type MEMCACHEDPlugin struct{}

func init() {
	plugins.RegisterPlugin(&MEMCACHEDPlugin{})
}

func (p *MEMCACHEDPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	response, err := utils.SendRecv(conn, []byte("version\r\n"), timeout)
	if err != nil {
		return nil, err
	}
	s := string(response)
	if !strings.HasPrefix(s, "VERSION ") {
		return nil, nil
	}
	version := strings.TrimSpace(strings.TrimPrefix(s, "VERSION "))
	payload := plugins.ServiceMemcached{Version: version}
	return plugins.CreateServiceFrom(target, payload, false, version, plugins.TCP), nil
}

func (p *MEMCACHEDPlugin) PortPriority(port uint16) bool {
	return port == 11211
}

func (p *MEMCACHEDPlugin) Name() string           { return MEMCACHED }
func (p *MEMCACHEDPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *MEMCACHEDPlugin) Priority() int          { return 200 }
