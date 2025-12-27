package kafka

import (
	"context"
	"errors"
	"strings"

	"github.com/ctrlsam/rigour/internal/messaging"

	kafka "github.com/segmentio/kafka-go"
)

type ProducerConfig struct {
	Brokers string // Comma-separated list of broker addresses
	Topic   string
}

type Producer struct {
	writer *kafka.Writer
}

var _ messaging.Producer = (*Producer)(nil)

func (c ProducerConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("kafka: brokers is empty")
	}
	if strings.TrimSpace(c.Topic) == "" {
		return errors.New("kafka: topic is empty")
	}
	return nil
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	brokers := strings.Split(cfg.Brokers, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchSize:              10,
		BatchTimeout:           100,
		Async:                  true,
	}

	producer := &Producer{writer: writer}

	return producer, nil
}

func (producer *Producer) Close() error {
	if producer == nil || producer.writer == nil {
		return nil
	}
	return producer.writer.Close()
}

func (producer *Producer) PublishBytes(ctx context.Context, key []byte, value []byte) error {
	if producer == nil || producer.writer == nil {
		return errors.New("kafka: producer is nil")
	}
	return producer.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)
}

// TypedProducer[T] wraps a Producer and automatically marshals messages of type T.
type TypedProducer[T any] struct {
	producer *Producer
	codec    messaging.Codec[T]
}

// NewTypedProducer creates a new producer that serializes messages of type T.
func NewTypedProducer[T any](cfg ProducerConfig) (*TypedProducer[T], error) {
	producer, err := NewProducer(cfg)
	if err != nil {
		return nil, err
	}
	return &TypedProducer[T]{
		producer: producer,
		codec:    messaging.JSONCodec[T]{},
	}, nil
}

// PublishMessage publishes and serializes a message of type T.
func (tp *TypedProducer[T]) PublishMessage(ctx context.Context, key []byte, value T) error {
	if tp == nil || tp.producer == nil {
		return errors.New("kafka: typed producer is nil")
	}

	serialized, err := tp.codec.Marshal(value)
	if err != nil {
		return err
	}

	return tp.producer.PublishBytes(ctx, key, serialized)
}

// Close closes the underlying producer.
func (tp *TypedProducer[T]) Close() error {
	if tp == nil || tp.producer == nil {
		return nil
	}
	return tp.producer.Close()
}
