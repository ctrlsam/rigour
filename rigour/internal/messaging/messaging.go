package messaging

import "context"

// Producer publishes messages to a topic/stream.
//
// Implementations are expected to be safe for concurrent use.
// Key is optional but recommended for stable partitioning.
type Producer interface {
	PublishBytes(ctx context.Context, key, value []byte) error
	Close() error
}

// TypedProducer[T] publishes typed messages to a topic/stream.
// The implementation should handle serialization of type T.
//
// Implementations are expected to be safe for concurrent use.
// Key is optional but recommended for stable partitioning.
type TypedProducer[T any] interface {
	PublishMessage(ctx context.Context, key []byte, value T) error
	Close() error
}

// Consumer subscribes to a topic/stream and yields raw messages.
//
// Implementations should honor ctx cancellation and return context.Canceled when appropriate.
type Consumer interface {
	FetchMessage(ctx context.Context) (*Message, error)
	Close() error
}

// Message is a provider-agnostic message envelope.
type Message struct {
	Key   []byte
	Value []byte
}
