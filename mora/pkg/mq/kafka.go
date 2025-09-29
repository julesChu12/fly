package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// KafkaMQ implements the Client interface using Apache Kafka
type KafkaMQ struct {
	config         Config
	producer       sarama.SyncProducer
	consumerGroup  sarama.ConsumerGroup
	consumers      map[string]*KafkaConsumer
	mu             sync.RWMutex
	closeOnce      sync.Once
	closed         chan struct{}
}

// KafkaConsumer handles message consumption for a specific topic
type KafkaConsumer struct {
	handler        MessageHandler
	consumerGroup  sarama.ConsumerGroup
	topic          string
	options        ConsumeOptions
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewKafkaMQ creates a new Kafka-based message queue client
func NewKafkaMQ(cfg Config) (Client, error) {
	config := sarama.NewConfig()

	// Producer configuration
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all replicas
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionGZIP
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1

	// Consumer configuration
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.MaxProcessingTime = 30 * time.Second
	config.Consumer.Return.Errors = true

	// Version configuration
	version, err := sarama.ParseKafkaVersion("2.8.0")
	if err != nil {
		return nil, fmt.Errorf("failed to parse kafka version: %w", err)
	}
	config.Version = version

	// Parse broker addresses
	brokers := []string{"localhost:9092"} // default
	if cfg.DSN != "" {
		brokers = []string{cfg.DSN}
	}

	// Create producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, "mora-mq-group", config)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create kafka consumer group: %w", err)
	}

	return &KafkaMQ{
		config:        cfg,
		producer:      producer,
		consumerGroup: consumerGroup,
		consumers:     make(map[string]*KafkaConsumer),
		closed:        make(chan struct{}),
	}, nil
}

// Publish publishes a message to a topic
func (k *KafkaMQ) Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error {
	options := &PublishOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Create message
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
		Headers: make([]sarama.RecordHeader, 0),
	}

	// Add message ID
	messageID := generateMessageID()
	message.Headers = append(message.Headers, sarama.RecordHeader{
		Key:   []byte("message_id"),
		Value: []byte(messageID),
	})

	// Add custom headers
	if options.Headers != nil {
		for key, value := range options.Headers {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				continue
			}
			message.Headers = append(message.Headers, sarama.RecordHeader{
				Key:   []byte(key),
				Value: valueBytes,
			})
		}
	}

	// Add retry information
	if options.MaxRetry > 0 {
		message.Headers = append(message.Headers, sarama.RecordHeader{
			Key:   []byte("max_retry"),
			Value: []byte(fmt.Sprintf("%d", options.MaxRetry)),
		})
	}

	// Add timestamp
	message.Timestamp = time.Now()

	// Send message
	_, _, err := k.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	log.Printf("Message published to topic %s with ID %s", topic, messageID)
	return nil
}

// PublishWithDelay publishes a message with delay (not natively supported by Kafka)
// This implementation uses headers to mark delayed messages and relies on consumer logic
func (k *KafkaMQ) PublishWithDelay(ctx context.Context, topic string, payload []byte, delay time.Duration, opts ...PublishOption) error {
	options := &PublishOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Add delay information to headers
	if options.Headers == nil {
		options.Headers = make(map[string]interface{})
	}
	options.Headers["delay_until"] = time.Now().Add(delay).Unix()
	options.Headers["is_delayed"] = true

	// Use the regular publish method with delay headers
	newOpts := append(opts, WithHeaders(options.Headers))
	return k.Publish(ctx, topic, payload, newOpts...)
}

// Subscribe subscribes to a topic and processes messages with handler
func (k *KafkaMQ) Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...ConsumeOption) error {
	options := &ConsumeOptions{
		ConcurrentWorkers: 1,
		MaxRetry:          3,
		RetryDelay:        5 * time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if already subscribed to this topic
	if _, exists := k.consumers[topic]; exists {
		return fmt.Errorf("already subscribed to topic: %s", topic)
	}

	// Create consumer context
	consumerCtx, cancel := context.WithCancel(ctx)

	consumer := &KafkaConsumer{
		handler:       handler,
		consumerGroup: k.consumerGroup,
		topic:         topic,
		options:       *options,
		ctx:           consumerCtx,
		cancel:        cancel,
	}

	k.consumers[topic] = consumer

	// Start consuming in goroutine
	go k.startConsumer(consumer)

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// startConsumer starts the consumer for a specific topic
func (k *KafkaMQ) startConsumer(consumer *KafkaConsumer) {
	defer consumer.cancel()

	for {
		select {
		case <-consumer.ctx.Done():
			log.Printf("Consumer for topic %s stopped", consumer.topic)
			return
		case <-k.closed:
			log.Printf("Kafka client closed, stopping consumer for topic %s", consumer.topic)
			return
		default:
			// Start consuming
			handler := &consumerGroupHandler{
				consumer: consumer,
			}

			err := consumer.consumerGroup.Consume(consumer.ctx, []string{consumer.topic}, handler)
			if err != nil {
				log.Printf("Error consuming from topic %s: %v", consumer.topic, err)
				time.Sleep(5 * time.Second) // Wait before retry
			}
		}
	}
}

// Close closes the Kafka client
func (k *KafkaMQ) Close() error {
	k.closeOnce.Do(func() {
		close(k.closed)

		// Stop all consumers
		k.mu.Lock()
		for topic, consumer := range k.consumers {
			consumer.cancel()
			log.Printf("Stopped consumer for topic: %s", topic)
		}
		k.consumers = make(map[string]*KafkaConsumer)
		k.mu.Unlock()

		// Close producer and consumer group
		if err := k.producer.Close(); err != nil {
			log.Printf("Error closing kafka producer: %v", err)
		}

		if err := k.consumerGroup.Close(); err != nil {
			log.Printf("Error closing kafka consumer group: %v", err)
		}

		log.Println("Kafka client closed")
	})

	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *KafkaConsumer
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Create worker pool
	workers := h.consumer.options.ConcurrentWorkers
	if workers <= 0 {
		workers = 1
	}

	// Create worker channels
	messageChan := make(chan *sarama.ConsumerMessage, workers*2)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.worker(session, messageChan)
		}()
	}

	// Feed messages to workers
	go func() {
		defer close(messageChan)
		for message := range claim.Messages() {
			select {
			case messageChan <- message:
			case <-h.consumer.ctx.Done():
				return
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
	return nil
}

// worker processes messages from the message channel
func (h *consumerGroupHandler) worker(session sarama.ConsumerGroupSession, messageChan <-chan *sarama.ConsumerMessage) {
	for kafkaMsg := range messageChan {
		// Convert Kafka message to our Message format
		msg := h.convertKafkaMessage(kafkaMsg)

		// Check if message is delayed
		if h.isDelayedMessage(msg) {
			// Skip delayed messages that haven't reached their time
			continue
		}

		// Process message with retry logic
		var err error
		maxRetries := h.consumer.options.MaxRetry

		for retry := 0; retry <= maxRetries; retry++ {
			err = h.consumer.handler(h.consumer.ctx, msg)
			if err == nil {
				// Success - mark message as processed
				session.MarkMessage(kafkaMsg, "")
				break
			}

			if retry < maxRetries {
				log.Printf("Message processing failed (attempt %d/%d): %v", retry+1, maxRetries+1, err)
				time.Sleep(h.consumer.options.RetryDelay)
			}
		}

		if err != nil {
			log.Printf("Message processing failed after %d retries: %v", maxRetries+1, err)
			// Could implement dead letter queue here
		}
	}
}

// convertKafkaMessage converts Kafka message to our Message format
func (h *consumerGroupHandler) convertKafkaMessage(kafkaMsg *sarama.ConsumerMessage) *Message {
	headers := make(map[string]interface{})
	messageID := uuid.New().String()

	// Extract headers
	for _, header := range kafkaMsg.Headers {
		key := string(header.Key)
		value := string(header.Value)

		if key == "message_id" {
			messageID = value
		} else {
			// Try to unmarshal JSON values
			var jsonValue interface{}
			if err := json.Unmarshal(header.Value, &jsonValue); err == nil {
				headers[key] = jsonValue
			} else {
				headers[key] = value
			}
		}
	}

	return &Message{
		ID:        messageID,
		Topic:     kafkaMsg.Topic,
		Payload:   kafkaMsg.Value,
		Headers:   headers,
		CreatedAt: kafkaMsg.Timestamp,
	}
}

// isDelayedMessage checks if a message should be delayed
func (h *consumerGroupHandler) isDelayedMessage(msg *Message) bool {
	if delayUntil, exists := msg.Headers["delay_until"]; exists {
		if timestamp, ok := delayUntil.(float64); ok {
			delayTime := time.Unix(int64(timestamp), 0)
			if time.Now().Before(delayTime) {
				return true
			}
		}
	}
	return false
}