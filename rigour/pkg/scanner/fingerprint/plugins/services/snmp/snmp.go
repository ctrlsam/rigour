package snmp

import (
	"bytes"
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins/pluginutils"
)

const SNMP = "SNMP"

type SNMPPlugin struct{}

func init() {
	plugins.RegisterPlugin(&SNMPPlugin{})
}

func (f *SNMPPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	RequestID := []byte{0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}
	InitialConnectionPackage := []byte{
		0x30, 0x29, // package length
		0x02, 0x01, 0x00, // Version: 1
		0x04, 0x06, // Community
		0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, // method: "public"
		0xa0, // PDU type: GET
		0x1c,
		0x02, 0x04, 0xff, 0xff, 0xff, 0xff, // Request ID: -1
		0x02, 0x01, 0x00, // Error status: no error
		0x02, 0x01, 0x00, // Error index
		0x30, 0x0e, 0x30, 0x0c, 0x06, 0x08, 0x2b, 0x06, // Object ID
		0x01, 0x02, 0x01, 0x01, 0x01, 0x00, 0x05, 0x00,
	}
	InfoOffset := 33

	response, err := utils.SendRecv(conn, InitialConnectionPackage, timeout)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, nil
	}

	idx := strings.Index(string(response), "public")
	if idx == -1 {
		return nil, nil
	}
	stringBegin := idx + InfoOffset
	if bytes.Contains(response, RequestID) {
		if stringBegin < len(response) {
			return plugins.CreateServiceFrom(target, plugins.ServiceSNMP{}, false,
				string(response[stringBegin:]), plugins.UDP), nil
		}
		return plugins.CreateServiceFrom(target, plugins.ServiceSNMP{}, false, "", plugins.UDP), nil
	}
	return nil, nil
}

func (f *SNMPPlugin) Name() string {
	return SNMP
}

func (f *SNMPPlugin) PortPriority(i uint16) bool {
	return i == 161
}

func (f *SNMPPlugin) Type() plugins.Protocol {
	return plugins.UDP
}

func (f *SNMPPlugin) Priority() int {
	return 81
}
