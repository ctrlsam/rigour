package messaging

import "context"

// Producer publishes typed messages to a topic/stream.
//
// Implementations are expected to be safe for concurrent use.
// Key is optional but recommended for stable partitioning.
type Producer[T any] interface {
	Publish(ctx context.Context, key []byte, value T) error
	Close() error
}

// Consumer subscribes to a topic/stream and yields typed messages.
//
// Implementations should honor ctx cancellation and return context.Canceled when appropriate.
type Consumer[T any] interface {
	Fetch(ctx context.Context) (*TypedMessage[T], error)
	Close() error
}
