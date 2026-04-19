package sip

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const SIP = "sip"

type SIPPlugin struct{}

func init() {
	plugins.RegisterPlugin(&SIPPlugin{})
}

func (p *SIPPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	ip := target.Address.Addr().String()
	port := target.Address.Port()

	// Minimal RFC 3261 OPTIONS request
	req := fmt.Sprintf(
		"OPTIONS sip:rigour@%s:%d SIP/2.0\r\n"+
			"Via: SIP/2.0/TCP %s:%d;branch=z9hG4bKrigour\r\n"+
			"Max-Forwards: 70\r\n"+
			"From: \"Rigour\" <sip:rigour@%s>;tag=rigour1\r\n"+
			"To: <sip:rigour@%s:%d>\r\n"+
			"Call-ID: rigour@%s\r\n"+
			"CSeq: 1 OPTIONS\r\n"+
			"Content-Length: 0\r\n\r\n",
		ip, port, ip, port, ip, ip, port, ip,
	)
	response, err := utils.SendRecv(conn, []byte(req), timeout)
	if err != nil {
		return nil, err
	}
	s := string(response)
	if !strings.HasPrefix(s, "SIP/2.0") {
		return nil, nil
	}
	payload := plugins.ServiceSIP{}
	for _, line := range strings.Split(s, "\r\n") {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "user-agent: ") {
			payload.UserAgent = strings.TrimSpace(line[len("user-agent: "):])
		}
		if strings.HasPrefix(lower, "allow: ") {
			payload.Allow = strings.TrimSpace(line[len("allow: "):])
		}
	}
	return plugins.CreateServiceFrom(target, payload, false, "", plugins.TCP), nil
}

func (p *SIPPlugin) PortPriority(port uint16) bool {
	return port == 5060
}

func (p *SIPPlugin) Name() string           { return SIP }
func (p *SIPPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *SIPPlugin) Priority() int          { return 350 }
