package ftp

import (
	"net"
	"regexp"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

var ftpResponse = regexp.MustCompile(`^\d{3}[- ](.*)\r`)

const FTP = "ftp"

type FTPPlugin struct{}

func init() {
	plugins.RegisterPlugin(&FTPPlugin{})
}

func (p *FTPPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	response, err := utils.Recv(conn, timeout)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, nil
	}

	matches := ftpResponse.FindStringSubmatch(string(response))
	if matches == nil {
		return nil, nil
	}

	payload := plugins.ServiceFTP{
		Banner: string(response),
	}

	return plugins.CreateServiceFrom(target, payload, false, "", plugins.TCP), nil
}

func (p *FTPPlugin) PortPriority(i uint16) bool {
	return i == 21
}

func (p *FTPPlugin) Name() string {
	return FTP
}

func (p *FTPPlugin) Type() plugins.Protocol {
	return plugins.TCP
}

func (p *FTPPlugin) Priority() int {
	return 10
}
