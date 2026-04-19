package channel

import (
	"context"

	"github.com/ctrlsam/rigour/internal/messaging"
)

// TypedChannel is an in-process Producer+Consumer backed by a buffered Go channel.
// It implements messaging.Producer[T] and messaging.Consumer[T], making it a
// drop-in replacement for distributed transports (e.g. Kafka) in single-node mode.
type TypedChannel[T any] struct {
	ch chan messaging.TypedMessage[T]
}

func New[T any](bufferSize int) *TypedChannel[T] {
	return &TypedChannel[T]{ch: make(chan messaging.TypedMessage[T], bufferSize)}
}

func (c *TypedChannel[T]) Publish(_ context.Context, key []byte, value T) error {
	c.ch <- messaging.TypedMessage[T]{Key: key, Value: value}
	return nil
}

func (c *TypedChannel[T]) Fetch(ctx context.Context) (*messaging.TypedMessage[T], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-c.ch:
		return &msg, nil
	}
}

func (c *TypedChannel[T]) Close() error {
	close(c.ch)
	return nil
}
