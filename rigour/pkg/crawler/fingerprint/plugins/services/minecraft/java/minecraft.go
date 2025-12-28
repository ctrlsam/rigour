package minecraftjava

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

type MinecraftPlugin struct{}

const MinecraftJava = "minecraft-java"
const DefaultPort = uint16(25565)

func init() {
	plugins.RegisterPlugin(&MinecraftPlugin{})
}

func (p *MinecraftPlugin) PortPriority(port uint16) bool {
	return port == DefaultPort
}

func (p *MinecraftPlugin) Name() string {
	return MinecraftJava
}

func (p *MinecraftPlugin) Type() plugins.Protocol {
	return plugins.TCP
}

func (p *MinecraftPlugin) Priority() int {
	return 1
}

type mcStatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
	} `json:"players"`
	Description any    `json:"description"`
	Favicon     string `json:"favicon"`
	SecureChat  bool   `json:"enforcesSecureChat"`
}

func writeVarInt(w io.Writer, v int32) error {
	uv := uint32(v)
	for {
		if (uv & ^uint32(0x7F)) == 0 {
			_, err := w.Write([]byte{byte(uv)})
			return err
		}
		b := byte(uv&0x7F) | 0x80
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
		uv >>= 7
	}
}

func readVarInt(r io.ByteReader) (int32, error) {
	var numRead int
	var result int32
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		value := int32(b & 0x7F)
		result |= value << (7 * numRead)
		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint too big")
		}
		if (b & 0x80) == 0 {
			break
		}
	}
	return result, nil
}

func writeString(w io.Writer, s string) error {
	if err := writeVarInt(w, int32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func encodeHandshake(host string, port uint16) ([]byte, error) {
	// Handshake packet: id=0x00
	// protocol version: use 754 (1.16.5) as a broadly accepted value for status.
	// server address, server port, next state=1 (status)
	body := &bytes.Buffer{}
	if err := writeVarInt(body, 0); err != nil { // packet id
		return nil, err
	}
	if err := writeVarInt(body, 754); err != nil {
		return nil, err
	}
	if err := writeString(body, host); err != nil {
		return nil, err
	}
	if err := binary.Write(body, binary.BigEndian, port); err != nil {
		return nil, err
	}
	if err := writeVarInt(body, 1); err != nil { // next state: status
		return nil, err
	}

	pkt := &bytes.Buffer{}
	if err := writeVarInt(pkt, int32(body.Len())); err != nil {
		return nil, err
	}
	pkt.Write(body.Bytes())
	return pkt.Bytes(), nil
}

func encodeStatusRequest() ([]byte, error) {
	body := &bytes.Buffer{}
	if err := writeVarInt(body, 0); err != nil { // packet id
		return nil, err
	}
	pkt := &bytes.Buffer{}
	if err := writeVarInt(pkt, int32(body.Len())); err != nil {
		return nil, err
	}
	pkt.Write(body.Bytes())
	return pkt.Bytes(), nil
}

func decodeStatusResponse(frame []byte) (mcStatusResponse, error) {
	// Frame is: packet length varint + packet data.
	// We'll parse using a bytes.Reader with ByteReader.
	r := bytes.NewReader(frame)
	length, err := readVarInt(r)
	if err != nil {
		return mcStatusResponse{}, err
	}
	if length <= 0 {
		return mcStatusResponse{}, fmt.Errorf("invalid packet length")
	}

	// Read exactly length bytes as packet payload
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return mcStatusResponse{}, err
	}
	pr := bytes.NewReader(payload)
	packetID, err := readVarInt(pr)
	if err != nil {
		return mcStatusResponse{}, err
	}
	if packetID != 0 {
		return mcStatusResponse{}, fmt.Errorf("unexpected packet id %d", packetID)
	}

	jsonLen, err := readVarInt(pr)
	if err != nil {
		return mcStatusResponse{}, err
	}
	if jsonLen < 0 || int(jsonLen) > pr.Len() {
		return mcStatusResponse{}, fmt.Errorf("invalid json length")
	}

	jsonBytes := make([]byte, jsonLen)
	if _, err := io.ReadFull(pr, jsonBytes); err != nil {
		return mcStatusResponse{}, err
	}

	var resp mcStatusResponse
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		return mcStatusResponse{}, err
	}
	return resp, nil
}

func (p *MinecraftPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	host := target.Host
	if host == "" {
		host = target.Address.Addr().String()
	}

	handshake, err := encodeHandshake(host, uint16(target.Address.Port()))
	if err != nil {
		return nil, err
	}
	if err := utils.Send(conn, handshake, timeout); err != nil {
		return nil, err
	}

	statusReq, err := encodeStatusRequest()
	if err != nil {
		return nil, err
	}
	respFrame, err := utils.SendRecv(conn, statusReq, timeout)
	if err != nil {
		return nil, err
	}
	if len(respFrame) == 0 {
		return nil, nil
	}

	resp, err := decodeStatusResponse(respFrame)
	if err != nil {
		return nil, nil
	}

	payload := plugins.ServiceMinecraftJava{
		VersionName:     resp.Version.Name,
		ProtocolVersion: resp.Version.Protocol,
		PlayersOnline:   resp.Players.Online,
		PlayersMax:      resp.Players.Max,
		Description:     resp.Description,
		Favicon:         resp.Favicon,
		EnforcesSecure:  resp.SecureChat,
	}

	return plugins.CreateServiceFrom(target, payload, false, resp.Version.Name, plugins.TCP), nil
}
