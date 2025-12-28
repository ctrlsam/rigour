package kafka

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ctrlsam/rigour/internal/messaging"

	kafka "github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers string // Comma-separated list of broker addresses
	Topic   string
	GroupID string
}

type Consumer struct {
	reader *kafka.Reader
}

func (c ConsumerConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("kafka: brokers is empty")
	}
	if strings.TrimSpace(c.Topic) == "" {
		return errors.New("kafka: topic is empty")
	}
	if strings.TrimSpace(c.GroupID) == "" {
		return errors.New("kafka: group id is empty")
	}
	return nil
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	brokers := strings.Split(cfg.Brokers, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 1 * time.Second, // automatic commits every second
	})

	return &Consumer{reader: reader}, nil
}

func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

// TypedConsumer[T] wraps a Consumer and automatically unmarshals messages to type T.
type TypedConsumer[T any] struct {
	consumer *Consumer
	codec    messaging.Codec[T]
}

var _ messaging.Consumer[any] = (*TypedConsumer[any])(nil)

// NewTypedConsumer creates a new consumer that parses messages to type T.
func NewTypedConsumer[T any](cfg ConsumerConfig) (*TypedConsumer[T], error) {
	consumer, err := NewConsumer(cfg)
	if err != nil {
		return nil, err
	}
	return &TypedConsumer[T]{
		consumer: consumer,
		codec:    messaging.JSONCodec[T]{},
	}, nil
}

// Fetch fetches and parses a message to type T.
func (tc *TypedConsumer[T]) Fetch(ctx context.Context) (*messaging.TypedMessage[T], error) {
	if tc == nil || tc.consumer == nil {
		return nil, errors.New("kafka: typed consumer is nil")
	}
	if tc.consumer.reader == nil {
		return nil, errors.New("kafka: consumer is nil")
	}

	m, err := tc.consumer.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	var value T
	if err := tc.codec.Unmarshal(m.Value, &value); err != nil {
		return nil, err
	}

	return &messaging.TypedMessage[T]{Key: m.Key, Value: value, Raw: m.Value}, nil
}

// Close closes the underlying consumer.
func (tc *TypedConsumer[T]) Close() error {
	if tc == nil || tc.consumer == nil {
		return nil
	}
	return tc.consumer.Close()
}
