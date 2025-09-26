package mq

import (
	"context"
	"fmt"
	"time"
)

// Message represents a message in the queue
type Message struct {
	ID      string                 `json:"id"`
	Topic   string                 `json:"topic"`
	Payload []byte                 `json:"payload"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Retry   int                    `json:"retry"`
	MaxRetry int                   `json:"max_retry"`
	CreatedAt time.Time            `json:"created_at"`
	DelayUntil *time.Time          `json:"delay_until,omitempty"`
}

// Publisher defines the interface for message publishers
type Publisher interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error
	// PublishWithDelay publishes a message with delay
	PublishWithDelay(ctx context.Context, topic string, payload []byte, delay time.Duration, opts ...PublishOption) error
	// Close closes the publisher
	Close() error
}

// Consumer defines the interface for message consumers
type Consumer interface {
	// Subscribe subscribes to a topic and processes messages with handler
	Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...ConsumeOption) error
	// Close closes the consumer
	Close() error
}

// MessageHandler is the function type for handling messages
type MessageHandler func(ctx context.Context, msg *Message) error

// PublishOption defines options for publishing messages
type PublishOption func(*PublishOptions)

// ConsumeOption defines options for consuming messages
type ConsumeOption func(*ConsumeOptions)

// PublishOptions holds options for publishing
type PublishOptions struct {
	Headers    map[string]interface{}
	MaxRetry   int
	RetryDelay time.Duration
}

// ConsumeOptions holds options for consuming
type ConsumeOptions struct {
	ConcurrentWorkers int
	MaxRetry         int
	RetryDelay       time.Duration
	DeadLetterQueue  string
}

// WithHeaders sets headers for publishing
func WithHeaders(headers map[string]interface{}) PublishOption {
	return func(opts *PublishOptions) {
		opts.Headers = headers
	}
}

// WithMaxRetry sets maximum retry count for publishing
func WithMaxRetry(maxRetry int) PublishOption {
	return func(opts *PublishOptions) {
		opts.MaxRetry = maxRetry
	}
}

// WithRetryDelay sets retry delay for publishing
func WithRetryDelay(delay time.Duration) PublishOption {
	return func(opts *PublishOptions) {
		opts.RetryDelay = delay
	}
}

// WithConcurrentWorkers sets concurrent workers for consuming
func WithConcurrentWorkers(workers int) ConsumeOption {
	return func(opts *ConsumeOptions) {
		opts.ConcurrentWorkers = workers
	}
}

// WithConsumeMaxRetry sets maximum retry count for consuming
func WithConsumeMaxRetry(maxRetry int) ConsumeOption {
	return func(opts *ConsumeOptions) {
		opts.MaxRetry = maxRetry
	}
}

// WithConsumeRetryDelay sets retry delay for consuming
func WithConsumeRetryDelay(delay time.Duration) ConsumeOption {
	return func(opts *ConsumeOptions) {
		opts.RetryDelay = delay
	}
}

// WithDeadLetterQueue sets dead letter queue for consuming
func WithDeadLetterQueue(dlq string) ConsumeOption {
	return func(opts *ConsumeOptions) {
		opts.DeadLetterQueue = dlq
	}
}

// Config holds the configuration for message queue
type Config struct {
	Driver   string            `json:"driver" yaml:"driver"`     // memory, redis
	DSN      string            `json:"dsn" yaml:"dsn"`           // connection string
	Options  map[string]string `json:"options" yaml:"options"`   // additional options
}

// DefaultConfig returns default MQ configuration
func DefaultConfig() Config {
	return Config{
		Driver:  "memory",
		DSN:     "",
		Options: make(map[string]string),
	}
}

// Client represents a message queue client that implements both Publisher and Consumer
type Client interface {
	Publisher
	Consumer
}

// New creates a new message queue client based on the driver
func New(cfg Config) (Client, error) {
	switch cfg.Driver {
	case "memory":
		return NewMemoryMQ(), nil
	case "redis":
		client, err := NewRedisMQ(cfg)
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported MQ driver: %s", cfg.Driver)
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}