package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/julesChu12/fly/mora/pkg/mq"
)

func main() {
	// Example: Using Kafka MQ
	kafkaExample()

	// Example: Using Redis MQ
	redisExample()

	// Example: Using Memory MQ
	memoryExample()
}

func kafkaExample() {
	fmt.Println("=== Kafka MQ Example ===")

	// Create Kafka client
	cfg := mq.Config{
		Driver: "kafka",
		DSN:    "localhost:9092", // Kafka broker address
		Options: map[string]string{
			"group_id": "mora-example-group",
		},
	}

	client, err := mq.New(cfg)
	if err != nil {
		log.Printf("Failed to create Kafka client: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Subscribe to messages
	go func() {
		err := client.Subscribe(ctx, "user-events", func(ctx context.Context, msg *mq.Message) error {
			log.Printf("Kafka: Received message ID=%s, Topic=%s, Payload=%s",
				msg.ID, msg.Topic, string(msg.Payload))
			return nil
		}, mq.WithConcurrentWorkers(3))

		if err != nil {
			log.Printf("Subscribe error: %v", err)
		}
	}()

	// Publish some messages
	for i := 0; i < 5; i++ {
		payload := fmt.Sprintf(`{"event": "user_created", "user_id": %d}`, i+1)

		err := client.Publish(ctx, "user-events", []byte(payload),
			mq.WithHeaders(map[string]interface{}{
				"source":    "user-service",
				"timestamp": time.Now().Unix(),
			}),
			mq.WithMaxRetry(3),
		)

		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		} else {
			log.Printf("Kafka: Published message %d", i+1)
		}

		time.Sleep(1 * time.Second)
	}

	// Publish delayed message
	delayedPayload := `{"event": "user_reminder", "message": "Please complete your profile"}`
	err = client.PublishWithDelay(ctx, "user-events", []byte(delayedPayload),
		10*time.Second)
	if err != nil {
		log.Printf("Failed to publish delayed message: %v", err)
	} else {
		log.Println("Kafka: Published delayed message")
	}

	// Wait to see messages
	time.Sleep(15 * time.Second)
}

func redisExample() {
	fmt.Println("\n=== Redis MQ Example ===")

	cfg := mq.Config{
		Driver: "redis",
		DSN:    "localhost:6379",
	}

	client, err := mq.New(cfg)
	if err != nil {
		log.Printf("Failed to create Redis client: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Subscribe
	go func() {
		err := client.Subscribe(ctx, "notifications", func(ctx context.Context, msg *mq.Message) error {
			log.Printf("Redis: Received message ID=%s, Payload=%s",
				msg.ID, string(msg.Payload))
			return nil
		})

		if err != nil {
			log.Printf("Subscribe error: %v", err)
		}
	}()

	// Publish
	payload := `{"type": "email", "recipient": "user@example.com"}`
	err = client.Publish(ctx, "notifications", []byte(payload))
	if err != nil {
		log.Printf("Failed to publish: %v", err)
	} else {
		log.Println("Redis: Published message")
	}

	time.Sleep(2 * time.Second)
}

func memoryExample() {
	fmt.Println("\n=== Memory MQ Example ===")

	cfg := mq.Config{
		Driver: "memory",
	}

	client, err := mq.New(cfg)
	if err != nil {
		log.Printf("Failed to create Memory client: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Subscribe
	go func() {
		err := client.Subscribe(ctx, "tasks", func(ctx context.Context, msg *mq.Message) error {
			log.Printf("Memory: Processing task ID=%s, Payload=%s",
				msg.ID, string(msg.Payload))
			return nil
		})

		if err != nil {
			log.Printf("Subscribe error: %v", err)
		}
	}()

	// Publish
	for i := 0; i < 3; i++ {
		payload := fmt.Sprintf(`{"task": "process_order", "order_id": %d}`, i+1)
		err = client.Publish(ctx, "tasks", []byte(payload))
		if err != nil {
			log.Printf("Failed to publish: %v", err)
		} else {
			log.Printf("Memory: Published task %d", i+1)
		}
	}

	time.Sleep(1 * time.Second)
}
