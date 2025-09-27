package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMQ implements message queue using Redis
type RedisMQ struct {
	client *redis.Client
	closed bool
}

// NewRedisMQ creates a new Redis-based message queue
func NewRedisMQ(cfg Config) (*RedisMQ, error) {
	opts, err := redis.ParseURL(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis DSN: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close() // Close the client before returning nil
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisMQ{
		client: client,
	}, nil
}

// Publish publishes a message to a topic using Redis list
func (rmq *RedisMQ) Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error {
	return rmq.PublishWithDelay(ctx, topic, payload, 0, opts...)
}

// PublishWithDelay publishes a message with delay using Redis sorted set
func (rmq *RedisMQ) PublishWithDelay(ctx context.Context, topic string, payload []byte, delay time.Duration, opts ...PublishOption) error {
	if rmq.closed {
		return ErrMQClosed
	}

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

	// Serialize message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if delay > 0 {
		// Use sorted set for delayed messages
		score := float64(time.Now().Add(delay).Unix())
		delayedKey := fmt.Sprintf("delayed:%s", topic)
		return rmq.client.ZAdd(ctx, delayedKey, redis.Z{
			Score:  score,
			Member: msgBytes,
		}).Err()
	}

	// Use list for immediate messages
	listKey := fmt.Sprintf("queue:%s", topic)
	return rmq.client.RPush(ctx, listKey, msgBytes).Err()
}

// Subscribe subscribes to a topic and processes messages
func (rmq *RedisMQ) Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...ConsumeOption) error {
	if rmq.closed {
		return ErrMQClosed
	}

	// Apply options
	options := &ConsumeOptions{
		ConcurrentWorkers: 1,
		MaxRetry:          3,
		RetryDelay:        time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Start workers
	for i := 0; i < options.ConcurrentWorkers; i++ {
		go rmq.worker(ctx, topic, handler, options)
	}

	// Start delayed message processor
	go rmq.delayedMessageProcessor(ctx, topic)

	// Wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// worker processes messages from Redis queue
func (rmq *RedisMQ) worker(ctx context.Context, topic string, handler MessageHandler, options *ConsumeOptions) {
	listKey := fmt.Sprintf("queue:%s", topic)
	processingKey := fmt.Sprintf("processing:%s", topic)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Move message from queue to processing list atomically
		result, err := rmq.client.BRPopLPush(ctx, listKey, processingKey, time.Second).Result()
		if err != nil {
			if err == redis.Nil {
				// No message available, continue polling
				continue
			}
			// Other error, wait before retry
			time.Sleep(time.Second)
			continue
		}

		// Deserialize message
		var msg Message
		if err := json.Unmarshal([]byte(result), &msg); err != nil {
			// Remove malformed message from processing list
			rmq.client.LRem(ctx, processingKey, 1, result)
			continue
		}

		// Process message
		err = rmq.processMessage(ctx, &msg, handler, options)
		if err != nil {
			// Handle failed message
			if msg.Retry >= options.MaxRetry {
				if options.DeadLetterQueue != "" {
					rmq.sendToDeadLetterQueue(ctx, options.DeadLetterQueue, &msg)
				}
				// Remove from processing list
				rmq.client.LRem(ctx, processingKey, 1, result)
			} else {
				// Retry: move back to queue
				rmq.client.LRem(ctx, processingKey, 1, result)
				time.Sleep(options.RetryDelay)
				msgBytes, _ := json.Marshal(msg)
				rmq.client.RPush(ctx, listKey, msgBytes)
			}
		} else {
			// Success: remove from processing list
			rmq.client.LRem(ctx, processingKey, 1, result)
		}
	}
}

// delayedMessageProcessor moves delayed messages to main queue when ready
func (rmq *RedisMQ) delayedMessageProcessor(ctx context.Context, topic string) {
	delayedKey := fmt.Sprintf("delayed:%s", topic)
	listKey := fmt.Sprintf("queue:%s", topic)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := float64(time.Now().Unix())

			// Get messages that are ready to be processed
			msgs, err := rmq.client.ZRangeByScore(ctx, delayedKey, &redis.ZRangeBy{
				Min: "0",
				Max: fmt.Sprintf("%f", now),
			}).Result()

			if err != nil || len(msgs) == 0 {
				continue
			}

			// Move messages from delayed set to main queue
			pipe := rmq.client.Pipeline()
			for _, msgStr := range msgs {
				pipe.ZRem(ctx, delayedKey, msgStr)
				pipe.RPush(ctx, listKey, msgStr)
			}
			pipe.Exec(ctx)
		}
	}
}

// processMessage processes a single message with retry logic
func (rmq *RedisMQ) processMessage(ctx context.Context, msg *Message, handler MessageHandler, options *ConsumeOptions) error {
	msg.Retry++
	err := handler(ctx, msg)
	if err != nil && msg.Retry < options.MaxRetry {
		return fmt.Errorf("message processing failed (retry %d/%d): %w", msg.Retry, options.MaxRetry, err)
	}
	return err
}

// sendToDeadLetterQueue sends failed message to dead letter queue
func (rmq *RedisMQ) sendToDeadLetterQueue(ctx context.Context, dlqTopic string, msg *Message) error {
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

	// Serialize and send to DLQ
	dlqMsgBytes, err := json.Marshal(dlqMsg)
	if err != nil {
		return err
	}

	dlqListKey := fmt.Sprintf("queue:%s", dlqTopic)
	return rmq.client.RPush(ctx, dlqListKey, dlqMsgBytes).Err()
}

// Close closes the Redis MQ client
func (rmq *RedisMQ) Close() error {
	if rmq.closed {
		return nil
	}
	rmq.closed = true
	return rmq.client.Close()
}

// GetClient returns the underlying Redis client
func (rmq *RedisMQ) GetClient() *redis.Client {
	return rmq.client
}

// Stats returns Redis MQ statistics
func (rmq *RedisMQ) Stats(ctx context.Context, topic string) (map[string]int64, error) {
	listKey := fmt.Sprintf("queue:%s", topic)
	processingKey := fmt.Sprintf("processing:%s", topic)
	delayedKey := fmt.Sprintf("delayed:%s", topic)

	pipe := rmq.client.Pipeline()
	queueLen := pipe.LLen(ctx, listKey)
	processingLen := pipe.LLen(ctx, processingKey)
	delayedLen := pipe.ZCard(ctx, delayedKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"queue":      queueLen.Val(),
		"processing": processingLen.Val(),
		"delayed":    delayedLen.Val(),
	}, nil
}
