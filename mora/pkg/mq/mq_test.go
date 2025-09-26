package mq

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	expected := Config{
		Driver:  "memory",
		DSN:     "",
		Options: make(map[string]string),
	}

	if cfg.Driver != expected.Driver {
		t.Errorf("DefaultConfig().Driver = %v, want %v", cfg.Driver, expected.Driver)
	}

	if cfg.DSN != expected.DSN {
		t.Errorf("DefaultConfig().DSN = %v, want %v", cfg.DSN, expected.DSN)
	}

	if cfg.Options == nil {
		t.Error("DefaultConfig().Options should not be nil")
	}
}

func TestNew(t *testing.T) {
	t.Run("memory driver", func(t *testing.T) {
		cfg := Config{Driver: "memory"}
		client, err := New(cfg)

		if err != nil {
			t.Errorf("New() error = %v", err)
		}

		if client == nil {
			t.Error("New() returned nil client")
		}

		// Clean up
		client.Close()
	})

	t.Run("unsupported driver", func(t *testing.T) {
		cfg := Config{Driver: "unsupported"}
		client, err := New(cfg)

		if err == nil {
			t.Error("New() should return error for unsupported driver")
		}

		if client != nil {
			t.Error("New() should return nil client for unsupported driver")
			client.Close()
		}
	})

	t.Run("redis driver with invalid DSN", func(t *testing.T) {
		cfg := Config{
			Driver: "redis",
			DSN:    "invalid://dsn",
		}
		client, err := New(cfg)

		if err == nil {
			t.Error("New() should return error for invalid Redis DSN")
		}

		if client != nil {
			t.Error("New() should return nil client for invalid DSN")
		}
	})
}

func TestMemoryMQ_BasicOperations(t *testing.T) {
	mq := NewMemoryMQ()
	defer mq.Close()

	ctx := context.Background()
	topic := "test-topic"
	payload := []byte("test message")

	t.Run("publish and consume", func(t *testing.T) {
		var receivedMsg *Message
		var wg sync.WaitGroup
		wg.Add(1)

		handler := func(ctx context.Context, msg *Message) error {
			receivedMsg = msg
			wg.Done()
			return nil
		}

		// Start consumer in background
		consumerCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go mq.Subscribe(consumerCtx, topic, handler)

		// Give consumer time to start
		time.Sleep(100 * time.Millisecond)

		// Publish message
		err := mq.Publish(ctx, topic, payload)
		if err != nil {
			t.Errorf("Publish() error = %v", err)
		}

		// Wait for message to be received
		wg.Wait()

		// Verify received message
		if receivedMsg == nil {
			t.Error("No message received")
		} else {
			if string(receivedMsg.Payload) != string(payload) {
				t.Errorf("Received payload = %s, want %s", receivedMsg.Payload, payload)
			}
			if receivedMsg.Topic != topic {
				t.Errorf("Received topic = %s, want %s", receivedMsg.Topic, topic)
			}
		}
	})

	t.Run("publish with options", func(t *testing.T) {
		headers := map[string]interface{}{
			"user_id": "123",
			"type":    "test",
		}

		var receivedMsg *Message
		var wg sync.WaitGroup
		wg.Add(1)

		handler := func(ctx context.Context, msg *Message) error {
			receivedMsg = msg
			wg.Done()
			return nil
		}

		consumerCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go mq.Subscribe(consumerCtx, topic+"_opts", handler)
		time.Sleep(100 * time.Millisecond)

		err := mq.Publish(ctx, topic+"_opts", payload,
			WithHeaders(headers),
			WithMaxRetry(5),
		)
		if err != nil {
			t.Errorf("Publish() with options error = %v", err)
		}

		wg.Wait()

		if receivedMsg == nil {
			t.Error("No message received")
		} else {
			if receivedMsg.MaxRetry != 5 {
				t.Errorf("MaxRetry = %d, want 5", receivedMsg.MaxRetry)
			}
			if receivedMsg.Headers["user_id"] != "123" {
				t.Errorf("Header user_id = %v, want 123", receivedMsg.Headers["user_id"])
			}
		}
	})
}

func TestMemoryMQ_DelayedMessages(t *testing.T) {
	mq := NewMemoryMQ()
	defer mq.Close()

	ctx := context.Background()
	topic := "delayed-topic"
	payload := []byte("delayed message")

	var receivedMsg *Message
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, msg *Message) error {
		receivedMsg = msg
		wg.Done()
		return nil
	}

	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go mq.Subscribe(consumerCtx, topic, handler)
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	delay := 200 * time.Millisecond

	err := mq.PublishWithDelay(ctx, topic, payload, delay)
	if err != nil {
		t.Errorf("PublishWithDelay() error = %v", err)
	}

	wg.Wait()
	elapsed := time.Since(start)

	if elapsed < delay {
		t.Errorf("Message received too early: elapsed %v < delay %v", elapsed, delay)
	}

	if receivedMsg == nil {
		t.Error("No delayed message received")
	} else if string(receivedMsg.Payload) != string(payload) {
		t.Errorf("Delayed message payload = %s, want %s", receivedMsg.Payload, payload)
	}
}

func TestMemoryMQ_ConcurrentWorkers(t *testing.T) {
	mq := NewMemoryMQ()
	defer mq.Close()

	ctx := context.Background()
	topic := "concurrent-topic"
	messageCount := 10

	var receivedCount int32
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(messageCount)

	handler := func(ctx context.Context, msg *Message) error {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
		return nil
	}

	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start consumer with 3 concurrent workers
	go mq.Subscribe(consumerCtx, topic, handler, WithConcurrentWorkers(3))
	time.Sleep(100 * time.Millisecond)

	// Publish multiple messages
	for i := 0; i < messageCount; i++ {
		payload := []byte("message " + string(rune('0'+i)))
		err := mq.Publish(ctx, topic, payload)
		if err != nil {
			t.Errorf("Publish() message %d error = %v", i, err)
		}
	}

	wg.Wait()

	mu.Lock()
	count := receivedCount
	mu.Unlock()

	if int(count) != messageCount {
		t.Errorf("Received %d messages, want %d", count, messageCount)
	}
}

func TestMemoryMQ_RetryLogic(t *testing.T) {
	mq := NewMemoryMQ()
	defer mq.Close()

	ctx := context.Background()
	topic := "retry-topic"
	payload := []byte("retry message")

	var attempts int
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, msg *Message) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("simulated failure")
		}
		wg.Done()
		return nil
	}

	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go mq.Subscribe(consumerCtx, topic, handler,
		WithConsumeMaxRetry(3),
		WithConsumeRetryDelay(10*time.Millisecond),
	)
	time.Sleep(100 * time.Millisecond)

	err := mq.Publish(ctx, topic, payload)
	if err != nil {
		t.Errorf("Publish() error = %v", err)
	}

	wg.Wait()

	if attempts != 3 {
		t.Errorf("Handler called %d times, want 3", attempts)
	}
}

func TestMemoryMQ_Close(t *testing.T) {
	mq := NewMemoryMQ()

	// Add some consumers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mq.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	// Close MQ
	err := mq.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Try to publish after close
	err = mq.Publish(ctx, "test", []byte("test"))
	if err != ErrMQClosed {
		t.Errorf("Publish() after close should return ErrMQClosed, got %v", err)
	}

	// Try to subscribe after close
	err = mq.Subscribe(ctx, "test", func(ctx context.Context, msg *Message) error {
		return nil
	})
	if err != ErrMQClosed {
		t.Errorf("Subscribe() after close should return ErrMQClosed, got %v", err)
	}

	// Close again should not error
	err = mq.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestMessage(t *testing.T) {
	t.Run("message creation", func(t *testing.T) {
		msg := &Message{
			ID:      "test-id",
			Topic:   "test-topic",
			Payload: []byte("test payload"),
			Headers: map[string]interface{}{
				"key": "value",
			},
			Retry:     1,
			MaxRetry:  3,
			CreatedAt: time.Now(),
		}

		if msg.ID != "test-id" {
			t.Errorf("Message ID = %s, want test-id", msg.ID)
		}

		if msg.Topic != "test-topic" {
			t.Errorf("Message Topic = %s, want test-topic", msg.Topic)
		}

		if string(msg.Payload) != "test payload" {
			t.Errorf("Message Payload = %s, want test payload", msg.Payload)
		}

		if msg.Headers["key"] != "value" {
			t.Errorf("Message Headers[key] = %v, want value", msg.Headers["key"])
		}
	})
}

func TestPublishOptions(t *testing.T) {
	options := &PublishOptions{}

	// Test WithHeaders
	headers := map[string]interface{}{"test": "value"}
	WithHeaders(headers)(options)
	if options.Headers["test"] != "value" {
		t.Error("WithHeaders did not set headers correctly")
	}

	// Test WithMaxRetry
	WithMaxRetry(5)(options)
	if options.MaxRetry != 5 {
		t.Error("WithMaxRetry did not set max retry correctly")
	}

	// Test WithRetryDelay
	delay := time.Minute
	WithRetryDelay(delay)(options)
	if options.RetryDelay != delay {
		t.Error("WithRetryDelay did not set retry delay correctly")
	}
}

func TestConsumeOptions(t *testing.T) {
	options := &ConsumeOptions{}

	// Test WithConcurrentWorkers
	WithConcurrentWorkers(10)(options)
	if options.ConcurrentWorkers != 10 {
		t.Error("WithConcurrentWorkers did not set concurrent workers correctly")
	}

	// Test WithConsumeMaxRetry
	WithConsumeMaxRetry(7)(options)
	if options.MaxRetry != 7 {
		t.Error("WithConsumeMaxRetry did not set max retry correctly")
	}

	// Test WithConsumeRetryDelay
	delay := time.Hour
	WithConsumeRetryDelay(delay)(options)
	if options.RetryDelay != delay {
		t.Error("WithConsumeRetryDelay did not set retry delay correctly")
	}

	// Test WithDeadLetterQueue
	dlq := "dead-letter-queue"
	WithDeadLetterQueue(dlq)(options)
	if options.DeadLetterQueue != dlq {
		t.Error("WithDeadLetterQueue did not set dead letter queue correctly")
	}
}

func TestGenerateMessageID(t *testing.T) {
	id1 := generateMessageID()
	id2 := generateMessageID()

	if id1 == "" {
		t.Error("generateMessageID() returned empty string")
	}

	if id2 == "" {
		t.Error("generateMessageID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateMessageID() returned same ID twice")
	}

	// Should start with "msg_"
	if len(id1) < 4 || id1[:4] != "msg_" {
		t.Errorf("generateMessageID() should start with 'msg_', got %s", id1)
	}
}