package netbios

import (
	"crypto/rand"
	"net"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins/pluginutils"
)

const NETBIOS = "netbios-ns"

type Plugin struct{}

func init() {
	plugins.RegisterPlugin(&Plugin{})
}

func (p *Plugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	transactionID := make([]byte, 2)
	_, err := rand.Read(transactionID)
	if err != nil {
		return nil, &utils.RandomizeError{Message: "Transaction ID"}
	}
	InitialConnectionPackage := append(transactionID, []byte{ //nolint:gocritic
		// Transaction ID
		0x00, 0x10, // Flag: Broadcast
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Queries
		0x20, 0x43, 0x4b, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x00,
		0x00, 0x21,
		0x00, 0x01,
	}...)

	response, err := utils.SendRecv(conn, InitialConnectionPackage, timeout)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, nil
	}

	stringBegin := strings.Index(string(response), "\x00\x00\x00\x00\x00") + 7
	stringEnd := strings.Index(string(response), "\x20\x20\x20")
	if stringBegin == -1 || stringEnd == -1 || stringEnd < stringBegin ||
		stringBegin >= len(response) || stringEnd >= len(response) {
		return nil, nil
	}
	payload := plugins.ServiceNetbios{
		NetBIOSName: string(response[stringBegin:stringEnd]),
	}
	return plugins.CreateServiceFrom(target, payload, false, "", plugins.UDP), nil
}

func (p *Plugin) PortPriority(i uint16) bool {
	return i == 137
}

func (p *Plugin) Name() string {
	return NETBIOS
}

func (p *Plugin) Type() plugins.Protocol {
	return plugins.UDP
}

func (p *Plugin) Priority() int {
	return 700
}
