package etcd

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const ETCD = "etcd"

type ETCDPlugin struct{}

type etcdVersion struct {
	EtcdServer  string `json:"etcdserver"`
	EtcdCluster string `json:"etcdcluster"`
}

func init() {
	plugins.RegisterPlugin(&ETCDPlugin{})
}

func (p *ETCDPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	ip := target.Address.Addr().String()
	req := fmt.Sprintf("GET /version HTTP/1.0\r\nHost: %s\r\n\r\n", ip)
	response, err := utils.SendRecv(conn, []byte(req), timeout)
	if err != nil {
		return nil, err
	}
	s := string(response)
	if !strings.Contains(s, "200") {
		return nil, nil
	}
	// Locate the response body after the header/body separator
	idx := strings.Index(s, "\r\n\r\n")
	if idx < 0 {
		return nil, nil
	}
	body := strings.TrimSpace(s[idx+4:])
	var ver etcdVersion
	if err := json.Unmarshal([]byte(body), &ver); err != nil || ver.EtcdServer == "" {
		return nil, nil
	}
	payload := plugins.ServiceEtcd{
		ServerVersion:  ver.EtcdServer,
		ClusterVersion: ver.EtcdCluster,
	}
	return plugins.CreateServiceFrom(target, payload, false, ver.EtcdServer, plugins.TCP), nil
}

func (p *ETCDPlugin) PortPriority(port uint16) bool {
	return port == 2379
}

func (p *ETCDPlugin) Name() string           { return ETCD }
func (p *ETCDPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *ETCDPlugin) Priority() int          { return 400 }
