package mq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryMQ implements message queue using in-memory storage
type MemoryMQ struct {
	topics    map[string]chan *Message
	consumers map[string][]chan *Message
	mutex     sync.RWMutex
	closed    bool
}

// NewMemoryMQ creates a new in-memory message queue
func NewMemoryMQ() *MemoryMQ {
	return &MemoryMQ{
		topics:    make(map[string]chan *Message),
		consumers: make(map[string][]chan *Message),
	}
}

// Publish publishes a message to a topic
func (mq *MemoryMQ) Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error {
	return mq.PublishWithDelay(ctx, topic, payload, 0, opts...)
}

// PublishWithDelay publishes a message with delay
func (mq *MemoryMQ) PublishWithDelay(ctx context.Context, topic string, payload []byte, delay time.Duration, opts ...PublishOption) error {
	mq.mutex.RLock()
	if mq.closed {
		mq.mutex.RUnlock()
		return ErrMQClosed
	}
	mq.mutex.RUnlock()

	// Apply options
	options := &PublishOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Create message
	msg := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Payload:   payload,
		Headers:   options.Headers,
		MaxRetry:  options.MaxRetry,
		CreatedAt: time.Now(),
	}

	if delay > 0 {
		delayUntil := time.Now().Add(delay)
		msg.DelayUntil = &delayUntil
	}

	// If delay is specified, wait before publishing
	if delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	mq.mutex.RLock()
	consumers := mq.consumers[topic]
	mq.mutex.RUnlock()

	// Send to all consumers of this topic
	for _, consumer := range consumers {
		select {
		case consumer <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Consumer buffer full, continue to next consumer
		}
	}

	return nil
}

// Subscribe subscribes to a topic and processes messages with handler
func (mq *MemoryMQ) Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...ConsumeOption) error {
	mq.mutex.Lock()
	if mq.closed {
		mq.mutex.Unlock()
		return ErrMQClosed
	}

	// Apply options
	options := &ConsumeOptions{
		ConcurrentWorkers: 1,
		MaxRetry:         3,
		RetryDelay:       time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Create consumer channel
	consumerChan := make(chan *Message, 100) // Buffer for 100 messages
	if mq.consumers[topic] == nil {
		mq.consumers[topic] = make([]chan *Message, 0)
	}
	mq.consumers[topic] = append(mq.consumers[topic], consumerChan)
	mq.mutex.Unlock()

	// Start workers
	for i := 0; i < options.ConcurrentWorkers; i++ {
		go mq.worker(ctx, consumerChan, handler, options)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// worker processes messages from consumer channel
func (mq *MemoryMQ) worker(ctx context.Context, consumerChan chan *Message, handler MessageHandler, options *ConsumeOptions) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-consumerChan:
			if msg == nil {
				return
			}

			// Check if message should be delayed
			if msg.DelayUntil != nil && time.Now().Before(*msg.DelayUntil) {
				// Re-queue the message after delay
				go func(m *Message) {
					select {
					case <-time.After(time.Until(*m.DelayUntil)):
						select {
						case consumerChan <- m:
						case <-ctx.Done():
						}
					case <-ctx.Done():
					}
				}(msg)
				continue
			}

			// Process message with retries
			err := mq.processMessage(ctx, msg, handler, options)
			if err != nil {
				// Handle failed message based on options
				if options.DeadLetterQueue != "" && msg.Retry >= options.MaxRetry {
					// Send to dead letter queue
					mq.sendToDeadLetterQueue(ctx, options.DeadLetterQueue, msg)
				}
			}
		}
	}
}

// processMessage processes a single message with retry logic
func (mq *MemoryMQ) processMessage(ctx context.Context, msg *Message, handler MessageHandler, options *ConsumeOptions) error {
	for msg.Retry <= options.MaxRetry {
		err := handler(ctx, msg)
		if err == nil {
			return nil // Success
		}

		msg.Retry++
		if msg.Retry <= options.MaxRetry {
			// Wait before retry
			select {
			case <-time.After(options.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return ErrMaxRetriesExceeded
}

// sendToDeadLetterQueue sends failed message to dead letter queue
func (mq *MemoryMQ) sendToDeadLetterQueue(ctx context.Context, dlqTopic string, msg *Message) {
	// Create DLQ message
	dlqMsg := &Message{
		ID:        generateMessageID(),
		Topic:     dlqTopic,
		Payload:   msg.Payload,
		Headers:   msg.Headers,
		CreatedAt: time.Now(),
	}

	// Add original message info to headers
	if dlqMsg.Headers == nil {
		dlqMsg.Headers = make(map[string]interface{})
	}
	dlqMsg.Headers["original_topic"] = msg.Topic
	dlqMsg.Headers["original_id"] = msg.ID
	dlqMsg.Headers["failed_retries"] = msg.Retry

	mq.mutex.RLock()
	consumers := mq.consumers[dlqTopic]
	mq.mutex.RUnlock()

	// Send to DLQ consumers
	for _, consumer := range consumers {
		select {
		case consumer <- dlqMsg:
		case <-ctx.Done():
			return
		default:
			// DLQ consumer buffer full, skip
		}
	}
}

// Close closes the memory MQ
func (mq *MemoryMQ) Close() error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if mq.closed {
		return nil
	}

	mq.closed = true

	// Close all consumer channels
	for topic, consumers := range mq.consumers {
		for _, consumer := range consumers {
			close(consumer)
		}
		delete(mq.consumers, topic)
	}

	// Close all topic channels
	for topic, ch := range mq.topics {
		close(ch)
		delete(mq.topics, topic)
	}

	return nil
}

var (
	// ErrMQClosed is returned when MQ is closed
	ErrMQClosed = fmt.Errorf("message queue is closed")
	// ErrMaxRetriesExceeded is returned when max retries are exceeded
	ErrMaxRetriesExceeded = fmt.Errorf("maximum retries exceeded")
)