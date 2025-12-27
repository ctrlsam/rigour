package ntp

import (
	"net"
	"time"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins/pluginutils"
)

const NTP = "ntp"

type Plugin struct{}

var ModeServer uint8 = 4

func init() {
	plugins.RegisterPlugin(&Plugin{})
}

func (p *Plugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	// reference: https://datatracker.ietf.org/doc/html/rfc5905#section-7.3
	InitialConnectionPackage := []byte{
		0xe3, 0x00, 0x0a, 0xf8, // LI/VN/Mode | Stratum | Poll | Precision
		0x00, 0x00, 0x00, 0x00, // Root Delay
		0x00, 0x00, 0x00, 0x00, // Root Dispersion
		0x00, 0x00, 0x00, 0x00, // Reference Identifier
		0x00, 0x00, 0x00, 0x00, // Reference Timestamp
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Origin Timestamp
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Receive Timestamp
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // Transmit Timestamp
		0x00, 0x00, 0x00, 0x00,
	}

	response, err := utils.SendRecv(conn, InitialConnectionPackage, timeout)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, nil
	}

	// check if response is valid NTP packet
	if response[0]&0x07 == ModeServer && len(response) == len(InitialConnectionPackage) {
		return plugins.CreateServiceFrom(target, plugins.ServiceNTP{}, false, "", plugins.UDP), nil
	}
	return nil, nil
}

func (p *Plugin) PortPriority(i uint16) bool {
	return i == 123
}

func (p *Plugin) Name() string {
	return NTP
}

func (p *Plugin) Type() plugins.Protocol {
	return plugins.UDP
}

func (p *Plugin) Priority() int {
	return 800
}
