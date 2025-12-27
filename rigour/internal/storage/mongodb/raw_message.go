package mongodb

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// bsonRawMessage is like json.RawMessage, but it can be unmarshaled directly from
// BSON values produced by MongoDB.
//
// Why this exists:
// The MongoDB Go driver cannot decode an embedded BSON document/array directly
// into json.RawMessage. It *can* decode strings/byte slices, but in our schema
// `metadata` is commonly stored as an embedded document.
//
// This type converts whatever BSON value is present into canonical JSON bytes.
// Supported BSON types: document, array, string, binary, null/undefined.
type bsonRawMessage json.RawMessage

func (m *bsonRawMessage) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if m == nil {
		return fmt.Errorf("bsonRawMessage: UnmarshalBSONValue on nil receiver")
	}

	switch t {
	case bsontype.Null, bsontype.Undefined:
		*m = nil
		return nil

	case bsontype.String:
		var s string
		if err := bson.UnmarshalValue(t, data, &s); err != nil {
			return err
		}
		// Best-effort: if it's already JSON, keep it; otherwise encode as a JSON string.
		b := []byte(s)
		if json.Valid(b) {
			*m = bsonRawMessage(b)
			return nil
		}
		encoded, err := json.Marshal(s)
		if err != nil {
			return err
		}
		*m = bsonRawMessage(encoded)
		return nil

	case bsontype.Binary:
		var b []byte
		if err := bson.UnmarshalValue(t, data, &b); err != nil {
			return err
		}
		if len(bytes.TrimSpace(b)) == 0 {
			*m = nil
			return nil
		}
		if json.Valid(b) {
			*m = bsonRawMessage(b)
			return nil
		}
		// Unknown binary payload; encode as base64 string via json.Marshal([]byte)
		encoded, err := json.Marshal(b)
		if err != nil {
			return err
		}
		*m = bsonRawMessage(encoded)
		return nil

	case bsontype.EmbeddedDocument:
		// Prefer *standard* JSON rather than MongoDB Extended JSON.
		// Also avoid decoding into bson.D which would serialize as [{"Key":...,"Value":...}].
		var v bson.M
		if err := bson.UnmarshalValue(t, data, &v); err != nil {
			return err
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		*m = bsonRawMessage(encoded)
		return nil

	case bsontype.Array:
		var v bson.A
		if err := bson.UnmarshalValue(t, data, &v); err != nil {
			return err
		}
		// Normalize: if it's a single-element array, unwrap it
		// This handles cases where scalar values are accidentally stored as [value]
		if len(v) == 1 {
			encoded, err := json.Marshal(v[0])
			if err != nil {
				return err
			}
			*m = bsonRawMessage(encoded)
			return nil
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		*m = bsonRawMessage(encoded)
		return nil

	default:
		// Fallback: try to round-trip through interface{} and JSON.
		var v interface{}
		if err := bson.UnmarshalValue(t, data, &v); err != nil {
			return err
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		*m = bsonRawMessage(encoded)
		return nil
	}
}

func (m bsonRawMessage) MarshalJSON() ([]byte, error) {
	if len(m) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(m).MarshalJSON()
}
