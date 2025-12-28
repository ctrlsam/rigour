package messaging

import (
	"encoding/json"
)

// Codec encodes/decodes values to/from bytes.
//
// This keeps serialization decisions out of transport providers.
type Codec[T any] interface {
	Marshal(value T) ([]byte, error)
	Unmarshal(data []byte, out *T) error
}

// JSONCodec is a convenience codec for JSON serialization.
type JSONCodec[T any] struct{}

func (JSONCodec[T]) Marshal(value T) ([]byte, error) { return json.Marshal(value) }
func (JSONCodec[T]) Unmarshal(data []byte, out *T) error {
	return json.Unmarshal(data, out)
}

type TypedMessage[T any] struct {
	Key   []byte
	Value T
	Raw   []byte
}
