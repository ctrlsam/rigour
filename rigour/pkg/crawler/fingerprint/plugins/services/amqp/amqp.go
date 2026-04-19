package amqp

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const AMQP = "amqp"

type AMQPPlugin struct{}

func init() {
	plugins.RegisterPlugin(&AMQPPlugin{})
}

// AMQP 0-9-1 protocol header sent by the client to initiate the handshake.
// The server responds with a Connection.Start method frame (or a version
// rejection header if it doesn't support 0-9-1).
var amqpHeader = []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}

func (p *AMQPPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	response, err := utils.SendRecv(conn, amqpHeader, timeout)
	if err != nil {
		return nil, err
	}
	if len(response) < 7 {
		return nil, nil
	}

	var version string
	if bytes.HasPrefix(response, []byte("AMQP")) && len(response) >= 8 {
		// Server rejected our version and replied with its own
		version = fmt.Sprintf("%d-%d-%d-%d", response[4], response[5], response[6], response[7])
	} else if response[0] == 0x01 && response[1] == 0x00 && response[2] == 0x00 {
		// Connection.Start method frame (frame-type=1, channel=0)
		version = "0-9-1"
	} else {
		return nil, nil
	}

	payload := plugins.ServiceAMQP{ProtocolVersion: version}
	return plugins.CreateServiceFrom(target, payload, false, version, plugins.TCP), nil
}

func (p *AMQPPlugin) PortPriority(port uint16) bool {
	return port == 5672
}

func (p *AMQPPlugin) Name() string        { return AMQP }
func (p *AMQPPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *AMQPPlugin) Priority() int       { return 100 }
