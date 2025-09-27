package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	expected := Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
	}

	if cfg != expected {
		t.Errorf("DefaultConfig() = %+v, want %+v", cfg, expected)
	}
}

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	client := New(cfg)

	if client == nil {
		t.Fatal("New() returned nil client")
	}

	if client.rdb == nil {
		t.Fatal("New() client has nil redis client")
	}

	// Clean up
	client.Close()
}

func TestClient_MethodsExist(t *testing.T) {
	// Test that all methods exist and can be called without Redis
	cfg := DefaultConfig()
	client := New(cfg)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test all methods exist (they will fail due to no Redis, but methods should exist)
	t.Run("Basic Operations Methods", func(t *testing.T) {
		// These will fail but we're testing method existence
		_, err := client.Get(ctx, "test")
		if err == nil {
			t.Log("Redis is available, operations succeeded")
		} else {
			t.Logf("Redis not available (expected): %v", err)
		}

		client.Set(ctx, "test", "value", time.Minute)
		client.Exists(ctx, "test")
		client.Delete(ctx, "test")
		client.Expire(ctx, "test", time.Minute)
		client.TTL(ctx, "test")
	})

	t.Run("Hash Operations Methods", func(t *testing.T) {
		client.HSet(ctx, "hash", "field", "value")
		client.HGet(ctx, "hash", "field")
		client.HGetAll(ctx, "hash")
		client.HDel(ctx, "hash", "field")
	})

	t.Run("List Operations Methods", func(t *testing.T) {
		client.LPush(ctx, "list", "value")
		client.RPush(ctx, "list", "value")
		client.LPop(ctx, "list")
		client.RPop(ctx, "list")
		client.LRange(ctx, "list", 0, -1)
	})

	t.Run("Set Operations Methods", func(t *testing.T) {
		client.SAdd(ctx, "set", "member")
		client.SMembers(ctx, "set")
		client.SIsMember(ctx, "set", "member")
		client.SRem(ctx, "set", "member")
	})

	t.Run("Advanced Operations", func(t *testing.T) {
		rdb := client.GetClient()
		if rdb == nil {
			t.Error("GetClient() returned nil")
		}

		pipe := client.Pipeline()
		if pipe == nil {
			t.Error("Pipeline() returned nil")
		}

		txPipe := client.TxPipeline()
		if txPipe == nil {
			t.Error("TxPipeline() returned nil")
		}
	})
}

func TestClient_WithRedis(t *testing.T) {
	// Only run if Redis is available
	cfg := DefaultConfig()
	client := New(cfg)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test if Redis is available
	if err := client.Ping(ctx); err != nil {
		t.Skipf("Redis not available, skipping integration tests: %v", err)
	}

	// If we get here, Redis is available
	t.Run("Basic Operations Integration", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		// Set value
		err := client.Set(ctx, key, value, time.Hour)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		// Get value
		result, err := client.Get(ctx, key)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		} else if result != value {
			t.Errorf("Get() = %v, want %v", result, value)
		}

		// Check existence
		exists, err := client.Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		} else if !exists {
			t.Error("Exists() = false, want true")
		}

		// Delete
		err = client.Delete(ctx, key)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		// Verify deletion
		exists, err = client.Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists() after delete error = %v", err)
		} else if exists {
			t.Error("Exists() after delete = true, want false")
		}
	})

	t.Run("Hash Operations Integration", func(t *testing.T) {
		key := "test_hash"
		field := "test_field"
		value := "test_value"

		// Set hash field
		err := client.HSet(ctx, key, field, value)
		if err != nil {
			t.Errorf("HSet() error = %v", err)
		}

		// Get hash field
		result, err := client.HGet(ctx, key, field)
		if err != nil {
			t.Errorf("HGet() error = %v", err)
		} else if result != value {
			t.Errorf("HGet() = %v, want %v", result, value)
		}

		// Get all hash fields
		allFields, err := client.HGetAll(ctx, key)
		if err != nil {
			t.Errorf("HGetAll() error = %v", err)
		} else if allFields[field] != value {
			t.Errorf("HGetAll()[%s] = %v, want %v", field, allFields[field], value)
		}

		// Delete hash field
		err = client.HDel(ctx, key, field)
		if err != nil {
			t.Errorf("HDel() error = %v", err)
		}

		// Clean up
		client.Delete(ctx, key)
	})
}

// Helper functions

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common Redis connection errors
	errStr := err.Error()
	return redis.HasErrorPrefix(err, "dial") ||
		redis.HasErrorPrefix(err, "connection") ||
		errStr == "redis: connection pool timeout" ||
		errStr == "redis: client is closed"
}
