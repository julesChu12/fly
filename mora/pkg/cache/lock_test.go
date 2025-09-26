package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultLockOptions(t *testing.T) {
	opts := DefaultLockOptions()

	expected := LockOptions{
		TTL:         DefaultLockTTL,
		RetryDelay:  DefaultRetryDelay,
		MaxRetries:  DefaultMaxRetries,
		LockTimeout: DefaultLockTimeout,
	}

	if opts != expected {
		t.Errorf("DefaultLockOptions() = %+v, want %+v", opts, expected)
	}
}

func TestGenerateLockValue(t *testing.T) {
	t.Run("generate unique values", func(t *testing.T) {
		value1 := generateLockValue()
		value2 := generateLockValue()

		if value1 == "" {
			t.Error("generateLockValue() returned empty string")
		}

		if value2 == "" {
			t.Error("generateLockValue() returned empty string")
		}

		if value1 == value2 {
			t.Error("generateLockValue() returned same value twice")
		}
	})
}

func TestDistributedLock_AccessorMethods(t *testing.T) {
	client := &Client{}
	key := "test_key"
	value := "test_value"
	ttl := 5 * time.Second

	lock := &DistributedLock{
		client: client,
		key:    key,
		value:  value,
		ttl:    ttl,
	}

	t.Run("Key()", func(t *testing.T) {
		if got := lock.Key(); got != key {
			t.Errorf("Key() = %v, want %v", got, key)
		}
	})

	t.Run("Value()", func(t *testing.T) {
		if got := lock.Value(); got != value {
			t.Errorf("Value() = %v, want %v", got, value)
		}
	})
}

func TestLockMethodsExist(t *testing.T) {
	// Test that lock methods exist without requiring Redis connection
	cfg := DefaultConfig()
	client := New(cfg)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	t.Run("Lock Methods Exist", func(t *testing.T) {
		// These will fail due to no Redis, but we're testing method existence
		_, err := client.TryLock(ctx, "test_key", time.Minute)
		if err == nil {
			t.Log("Redis is available, TryLock succeeded")
		} else {
			t.Logf("Redis not available (expected): %v", err)
		}

		_, err = client.Lock(ctx, "test_key")
		if err == nil {
			t.Log("Redis is available, Lock succeeded")
		} else {
			t.Logf("Redis not available (expected): %v", err)
		}

		// Test WithLock
		executed := false
		fn := func() error {
			executed = true
			return nil
		}

		err = client.WithLock(ctx, "test_key", fn)
		if err == nil {
			t.Log("Redis is available, WithLock succeeded")
			if !executed {
				t.Error("Function should have been executed")
			}
		} else {
			t.Logf("Redis not available (expected): %v", err)
			if executed {
				t.Error("Function should not have been executed without Redis")
			}
		}
	})
}

func TestLockIntegration(t *testing.T) {
	// Only run if Redis is available
	cfg := DefaultConfig()
	client := New(cfg)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test if Redis is available
	if err := client.Ping(ctx); err != nil {
		t.Skipf("Redis not available, skipping lock integration tests: %v", err)
	}

	// Extend context for actual tests since Redis is available
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("TryLock Integration", func(t *testing.T) {
		key := "test_lock_integration"
		ttl := 5 * time.Second

		// First lock should succeed
		lock1, err := client.TryLock(ctx, key, ttl)
		if err != nil {
			t.Errorf("First TryLock() error = %v", err)
			return
		}
		defer lock1.Unlock(ctx)

		if lock1.Key() != key {
			t.Errorf("lock.Key() = %v, want %v", lock1.Key(), key)
		}

		if lock1.Value() == "" {
			t.Error("lock.Value() is empty")
		}

		// Second lock should fail
		lock2, err := client.TryLock(ctx, key, ttl)
		if err == nil {
			defer lock2.Unlock(ctx)
			t.Error("Second TryLock() should have failed")
		} else if !errors.Is(err, ErrLockNotAcquired) {
			t.Errorf("Second TryLock() error = %v, want %v", err, ErrLockNotAcquired)
		}
	})

	t.Run("Lock with retry Integration", func(t *testing.T) {
		key := "test_lock_retry_integration"

		opts := LockOptions{
			TTL:         2 * time.Second,
			RetryDelay:  100 * time.Millisecond,
			MaxRetries:  3,
			LockTimeout: 1 * time.Second,
		}

		lock, err := client.Lock(ctx, key, opts)
		if err != nil {
			t.Errorf("Lock() error = %v", err)
			return
		}
		defer lock.Unlock(ctx)

		if lock == nil {
			t.Error("Lock() returned nil lock")
		}
	})

	t.Run("Lock Operations Integration", func(t *testing.T) {
		key := "test_lock_ops_integration"

		lock, err := client.TryLock(ctx, key, 10*time.Second)
		if err != nil {
			t.Errorf("TryLock() error = %v", err)
			return
		}

		// Test IsLocked
		isLocked, err := lock.IsLocked(ctx)
		if err != nil {
			t.Errorf("IsLocked() error = %v", err)
		} else if !isLocked {
			t.Error("IsLocked() = false, want true")
		}

		// Test GetTTL
		ttl, err := lock.GetTTL(ctx)
		if err != nil {
			t.Errorf("GetTTL() error = %v", err)
		} else if ttl <= 0 {
			t.Errorf("GetTTL() = %v, want > 0", ttl)
		}

		// Test Extend
		newTTL := 15 * time.Second
		err = lock.Extend(ctx, newTTL)
		if err != nil {
			t.Errorf("Extend() error = %v", err)
		}

		if lock.ttl != newTTL {
			t.Errorf("lock.ttl = %v, want %v", lock.ttl, newTTL)
		}

		// Test Unlock
		err = lock.Unlock(ctx)
		if err != nil {
			t.Errorf("Unlock() error = %v", err)
		}

		// Verify unlock
		isLocked, err = lock.IsLocked(ctx)
		if err != nil {
			t.Errorf("IsLocked() after unlock error = %v", err)
		} else if isLocked {
			t.Error("IsLocked() after unlock = true, want false")
		}
	})

	t.Run("WithLock Integration", func(t *testing.T) {
		key := "test_withlock_integration"
		executed := false

		fn := func() error {
			executed = true
			return nil
		}

		err := client.WithLock(ctx, key, fn)
		if err != nil {
			t.Errorf("WithLock() error = %v", err)
		}

		if !executed {
			t.Error("Function was not executed")
		}

		// Test with function that returns error
		expectedErr := errors.New("function error")
		fnWithError := func() error {
			return expectedErr
		}

		err = client.WithLock(ctx, key+"_error", fnWithError)
		if !errors.Is(err, expectedErr) {
			t.Errorf("WithLock() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("Lock Error Cases Integration", func(t *testing.T) {
		// Test unlock non-existent lock
		lock := &DistributedLock{
			client: client,
			key:    "non_existent_lock",
			value:  "fake_value",
			ttl:    5 * time.Second,
		}

		err := lock.Unlock(ctx)
		if !errors.Is(err, ErrLockNotOwned) {
			t.Errorf("Unlock() on non-existent lock error = %v, want %v", err, ErrLockNotOwned)
		}

		// Test extend non-owned lock
		err = lock.Extend(ctx, 10*time.Second)
		if !errors.Is(err, ErrLockNotOwned) {
			t.Errorf("Extend() on non-owned lock error = %v, want %v", err, ErrLockNotOwned)
		}
	})
}