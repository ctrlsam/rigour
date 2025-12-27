package messaging

import (
	"context"
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

// TypedConsumerImpl[T] wraps a raw Consumer and decodes messages into T.
// This is a concrete implementation of the typed consumer pattern.
type TypedConsumerImpl[T any] struct {
	Consumer Consumer
	Codec    Codec[T]
}

type TypedMessage[T any] struct {
	Key   []byte
	Value T
	Raw   []byte
}

// FetchMessage fetches and parses a message to type T.
func (tc TypedConsumerImpl[T]) FetchMessage(ctx context.Context) (*TypedMessage[T], error) {
	m, err := tc.Consumer.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	var v T
	if err := tc.Codec.Unmarshal(m.Value, &v); err != nil {
		return nil, err
	}

	return &TypedMessage[T]{Key: m.Key, Value: v, Raw: m.Value}, nil
}

// Close closes the underlying consumer.
func (tc TypedConsumerImpl[T]) Close() error {
	if tc.Consumer == nil {
		return nil
	}
	return tc.Consumer.Close()
}

// Fetch blocks until a message is available or ctx is canceled.
// Deprecated: Use FetchMessage instead.
func (tc TypedConsumerImpl[T]) Fetch(ctx context.Context) (*TypedMessage[T], error) {
	return tc.FetchMessage(ctx)
}
