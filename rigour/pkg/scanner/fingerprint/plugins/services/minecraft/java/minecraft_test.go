package minecraftjava

import (
	"bytes"
	"encoding/json"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/ctrlsam/rigour/pkg/scanner/fingerprint/plugins"
)

type byteReader struct{ *bytes.Reader }

func (b byteReader) ReadByte() (byte, error) { return b.Reader.ReadByte() }

func mustEncodeHandshake(t *testing.T, host string, port uint16) []byte {
	t.Helper()
	b, err := encodeHandshake(host, port)
	if err != nil {
		t.Fatalf("encodeHandshake: %v", err)
	}
	return b
}

func mustEncodeStatusReq(t *testing.T) []byte {
	t.Helper()
	b, err := encodeStatusRequest()
	if err != nil {
		t.Fatalf("encodeStatusRequest: %v", err)
	}
	return b
}

func mustFrameStatusJSON(t *testing.T, status any) []byte {
	t.Helper()
	j, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	// Packet payload: packetID=0 + jsonString
	payload := &bytes.Buffer{}
	if err := writeVarInt(payload, 0); err != nil {
		t.Fatalf("writeVarInt(packetID): %v", err)
	}
	if err := writeVarInt(payload, int32(len(j))); err != nil {
		t.Fatalf("writeVarInt(jsonLen): %v", err)
	}
	payload.Write(j)

	// Frame: length varint + payload
	frame := &bytes.Buffer{}
	if err := writeVarInt(frame, int32(payload.Len())); err != nil {
		t.Fatalf("writeVarInt(frameLen): %v", err)
	}
	frame.Write(payload.Bytes())
	return frame.Bytes()
}

func TestVarIntRoundTrip(t *testing.T) {
	vals := []int32{0, 1, 2, 127, 128, 255, 2097151}
	for _, v := range vals {
		buf := &bytes.Buffer{}
		if err := writeVarInt(buf, v); err != nil {
			t.Fatalf("writeVarInt(%d): %v", v, err)
		}
		got, err := readVarInt(byteReader{bytes.NewReader(buf.Bytes())})
		if err != nil {
			t.Fatalf("readVarInt(%d): %v", v, err)
		}
		if got != v {
			t.Fatalf("varint mismatch: want %d got %d", v, got)
		}
	}
}

func TestMinecraftJavaRunDetectsServer(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// lightweight fake MC status server on the other end of the pipe
	go func() {
		defer server.Close()
		buf := make([]byte, 4096)
		// read handshake
		_, _ = server.Read(buf)
		// read status request
		_, _ = server.Read(buf)

		status := map[string]any{
			"version":            map[string]any{"name": "1.20.4", "protocol": 765},
			"players":            map[string]any{"max": 20, "online": 3},
			"description":        map[string]any{"text": "hello"},
			"favicon":            "data:image/png;base64,AAAA",
			"enforcesSecureChat": true,
		}
		frame := mustFrameStatusJSON(t, status)
		_, _ = server.Write(frame)
	}()

	p := &MinecraftPlugin{}
	target := plugins.Target{
		Host:    "example.org",
		Address: netip.MustParseAddrPort("127.0.0.1:25565"),
	}

	service, err := p.Run(client, 2*time.Second, target)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if service == nil {
		t.Fatalf("expected service, got nil")
	}
	if service.Protocol != plugins.ProtoMinecraftJava {
		t.Fatalf("expected protocol %q got %q", plugins.ProtoMinecraftJava, service.Protocol)
	}
	if service.Port != 25565 {
		t.Fatalf("expected port 25565 got %d", service.Port)
	}

	meta := service.Metadata().(plugins.ServiceMinecraftJava)
	if meta.VersionName != "1.20.4" {
		t.Fatalf("expected versionName 1.20.4 got %q", meta.VersionName)
	}
	if meta.PlayersOnline != 3 || meta.PlayersMax != 20 {
		t.Fatalf("expected players 3/20 got %d/%d", meta.PlayersOnline, meta.PlayersMax)
	}
}

func TestMinecraftJavaRunReturnsNilOnNonMC(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		defer server.Close()
		// read whatever client writes and then respond with junk
		buf := make([]byte, 4096)
		_, _ = server.Read(buf)
		_, _ = server.Read(buf)
		_, _ = server.Write([]byte("not minecraft"))
	}()

	p := &MinecraftPlugin{}
	target := plugins.Target{Address: netip.MustParseAddrPort("127.0.0.1:25565")}

	svc, err := p.Run(client, 2*time.Second, target)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if svc != nil {
		t.Fatalf("expected nil service for non-mc response")
	}
}

func TestEncodersBuildPackets(t *testing.T) {
	h := mustEncodeHandshake(t, "localhost", 25565)
	if len(h) == 0 {
		t.Fatalf("handshake packet empty")
	}
	s := mustEncodeStatusReq(t)
	if len(s) == 0 {
		t.Fatalf("status request packet empty")
	}
}
