package mongodb

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
	utils "github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins/pluginutils"
)

const MONGODB = "mongodb"

type MONGODBPlugin struct{}

func init() {
	plugins.RegisterPlugin(&MONGODBPlugin{})
}

// isMasterQuery is a minimal MongoDB wire protocol OP_QUERY (opCode 2004)
// targeting admin.$cmd with the document {isMaster: 1}.
//
// Wire format: 16-byte header | flags(4) | collection(cstring) | skip(4) | return(4) | BSON
var isMasterQuery = []byte{
	// Header
	0x3a, 0x00, 0x00, 0x00, // messageLength: 58
	0x01, 0x00, 0x00, 0x00, // requestID: 1
	0x00, 0x00, 0x00, 0x00, // responseTo: 0
	0xd4, 0x07, 0x00, 0x00, // opCode: 2004 (OP_QUERY)
	// Body
	0x00, 0x00, 0x00, 0x00, // flags: 0
	// fullCollectionName "admin.$cmd\x00"
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x24, 0x63, 0x6d, 0x64, 0x00,
	0x00, 0x00, 0x00, 0x00, // numberToSkip: 0
	0xff, 0xff, 0xff, 0xff, // numberToReturn: -1
	// BSON {isMaster: 1} — 19 bytes
	0x13, 0x00, 0x00, 0x00, // document length: 19
	0x10,                                                               // element type: int32
	0x69, 0x73, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x00, // "isMaster\x00"
	0x01, 0x00, 0x00, 0x00, // value: 1
	0x00, // end of document
}

func (p *MONGODBPlugin) Run(conn net.Conn, timeout time.Duration, target plugins.Target) (*plugins.Service, error) {
	response, err := utils.SendRecv(conn, isMasterQuery, timeout)
	if err != nil {
		return nil, err
	}
	// OP_REPLY response has opCode == 1 at bytes 12–15
	if len(response) < 16 {
		return nil, nil
	}
	if binary.LittleEndian.Uint32(response[12:16]) != 1 {
		return nil, nil
	}
	// Attempt to parse maxWireVersion from the BSON document embedded in the reply.
	// OP_REPLY fixed header beyond message header: responseFlags(4) + cursorID(8) + startingFrom(4) + numberReturned(4) = 20 bytes
	payload := plugins.ServiceMongoDB{}
	if len(response) >= 36 {
		bsonStart := 36
		payload.MaxWireVersion = parseBSONInt32Field(response[bsonStart:], "maxWireVersion")
	}
	return plugins.CreateServiceFrom(target, payload, false, "", plugins.TCP), nil
}

// parseBSONInt32Field scans a BSON document for a top-level int32 field by name.
// Returns 0 if not found or the document is malformed.
func parseBSONInt32Field(doc []byte, key string) int {
	if len(doc) < 5 {
		return 0
	}
	i := 4 // skip document length
	for i < len(doc)-1 {
		elemType := doc[i]
		i++
		// Read null-terminated key
		j := i
		for j < len(doc) && doc[j] != 0x00 {
			j++
		}
		if j >= len(doc) {
			break
		}
		elemKey := string(doc[i:j])
		i = j + 1 // skip null terminator
		switch elemType {
		case 0x10: // int32
			if len(doc) < i+4 {
				return 0
			}
			val := int(binary.LittleEndian.Uint32(doc[i : i+4]))
			if elemKey == key {
				return val
			}
			i += 4
		case 0x12: // int64
			i += 8
		case 0x01: // double
			i += 8
		case 0x08: // bool
			i++
		case 0x09: // UTC datetime
			i += 8
		case 0x02, 0x0d, 0x0e: // string, JS code, symbol
			if len(doc) < i+4 {
				return 0
			}
			strLen := int(binary.LittleEndian.Uint32(doc[i : i+4]))
			i += 4 + strLen
		case 0x03, 0x04: // embedded document or array
			if len(doc) < i+4 {
				return 0
			}
			docLen := int(binary.LittleEndian.Uint32(doc[i : i+4]))
			i += docLen
		case 0x07: // ObjectID
			i += 12
		case 0x11: // timestamp
			i += 8
		case 0x00: // end of document
			return 0
		default:
			return 0
		}
	}
	return 0
}

func (p *MONGODBPlugin) PortPriority(port uint16) bool {
	return port == 27017
}

func (p *MONGODBPlugin) Name() string           { return MONGODB }
func (p *MONGODBPlugin) Type() plugins.Protocol { return plugins.TCP }
func (p *MONGODBPlugin) Priority() int          { return 150 }
