package irc

import (
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const IRC = "irc"

type IRCPlugin struct{}

func init() {
	plugins.RegisterPlugin(&IRCPlugin{})
}

func (p *IRCPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	response, err := utils.Recv(conn, timeout)
	if err != nil {
		return nil, err
	}
	s := string(response)
	// IRC server messages begin with ':server COMMAND ...'
	if !strings.HasPrefix(s, ":") {
		return nil, nil
	}
	// Extract server name: first token after the leading ':'
	serverName := ""
	if parts := strings.SplitN(s[1:], " ", 2); len(parts) > 0 {
		serverName = parts[0]
	}
	payload := plugins.ServiceIRC{
		ServerName: serverName,
		Banner:     strings.TrimSpace(s),
	}
	return plugins.CreateServiceFrom(target, payload, false, "", plugins.TCP), nil
}

// MatchBanner identifies IRC from its server-first banner.
func (p *IRCPlugin) MatchBanner(banner []byte) bool {
	return len(banner) > 0 && banner[0] == ':'
}

func (p *IRCPlugin) PortPriority(port uint16) bool {
	return port == 6667 || port == 6697
}

func (p *IRCPlugin) Name() string           { return IRC }
func (p *IRCPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *IRCPlugin) Priority() int          { return 50 }
