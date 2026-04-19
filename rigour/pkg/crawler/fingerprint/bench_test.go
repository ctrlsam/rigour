package fingerprint_test

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint"
	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"

	// Register plugins under test.
	_ "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/services/http"
	_ "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/services/redis"
	_ "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/services/ssh"
)

// serveMock starts a local TCP listener on an ephemeral port that writes banner
// to every accepted connection, then closes it. Listener is cleaned up via t.Cleanup.
func serveMock(t testing.TB, banner []byte) netip.AddrPort {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write(banner)
			}(c)
		}
	}()
	ap, _ := netip.ParseAddrPort(ln.Addr().String())
	return ap
}

// slowCfg runs the full slow-lane (all plugins tried) — the worst-case path.
var slowCfg = fingerprint.FingerprintConfig{
	DefaultTimeout: 500 * time.Millisecond,
	FastMode:       false,
}

// fastCfg runs the fast-lane only (port-priority matching).
var fastCfg = fingerprint.FingerprintConfig{
	DefaultTimeout: 500 * time.Millisecond,
	FastMode:       true,
}

// BenchmarkFingerprintUnknown measures the worst-case cost: garbage banner, no plugin
// matches, every plugin in the slow lane is exhausted.
func BenchmarkFingerprintUnknown(b *testing.B) {
	addr := serveMock(b, []byte("XXUNKNOWN_PROTO_GARBAGE\r\n"))
	target := plugins.Target{Address: addr}
	b.ResetTimer()
	for b.Loop() {
		slowCfg.FingerprintTarget(target)
	}
}

// BenchmarkFingerprintUnknownFastMode measures fast-lane cost when no plugin port-priority
// matches the ephemeral port — represents a typical port >49152.
func BenchmarkFingerprintUnknownFastMode(b *testing.B) {
	addr := serveMock(b, []byte("XXUNKNOWN_PROTO_GARBAGE\r\n"))
	target := plugins.Target{Address: addr}
	b.ResetTimer()
	for b.Loop() {
		fastCfg.FingerprintTarget(target)
	}
}

// BenchmarkFingerprintHTTPSlow measures the full slow-lane cost when the server speaks
// HTTP but is on an ephemeral port (no fast-lane shortcut).
func BenchmarkFingerprintHTTPSlow(b *testing.B) {
	addr := serveMock(b, []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
	target := plugins.Target{Address: addr}
	b.ResetTimer()
	for b.Loop() {
		slowCfg.FingerprintTarget(target)
	}
}

// BenchmarkFingerprintRedisSlow measures slow-lane Redis identification on an ephemeral port.
// Redis plugin sends PING and expects +PONG.
func BenchmarkFingerprintRedisSlow(b *testing.B) {
	// Simulate a Redis server: write +PONG\r\n on any connection.
	addr := serveMock(b, []byte("+PONG\r\n"))
	target := plugins.Target{Address: addr}
	b.ResetTimer()
	for b.Loop() {
		slowCfg.FingerprintTarget(target)
	}
}
