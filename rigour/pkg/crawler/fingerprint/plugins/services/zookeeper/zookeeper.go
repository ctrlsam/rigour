package zookeeper

import (
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const ZOOKEEPER = "zookeeper"

type ZOOKEEPERPlugin struct{}

func init() {
	plugins.RegisterPlugin(&ZOOKEEPERPlugin{})
}

func (p *ZOOKEEPERPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	// ZooKeeper four-letter command: ruok → imok
	response, err := utils.SendRecv(conn, []byte("ruok"), timeout)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(string(response), "imok") {
		return nil, nil
	}
	payload := plugins.ServiceZooKeeper{}
	return plugins.CreateServiceFrom(target, payload, false, "", plugins.TCP), nil
}

func (p *ZOOKEEPERPlugin) PortPriority(port uint16) bool {
	return port == 2181
}

func (p *ZOOKEEPERPlugin) Name() string           { return ZOOKEEPER }
func (p *ZOOKEEPERPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *ZOOKEEPERPlugin) Priority() int          { return 300 }
