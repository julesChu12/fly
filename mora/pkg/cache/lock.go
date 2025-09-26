package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Default lock settings
	DefaultLockTTL     = 30 * time.Second
	DefaultRetryDelay  = 100 * time.Millisecond
	DefaultMaxRetries  = 10
	DefaultLockTimeout = 5 * time.Second
)

var (
	// ErrLockNotAcquired is returned when a lock cannot be acquired
	ErrLockNotAcquired = errors.New("lock not acquired")
	// ErrLockNotOwned is returned when trying to release a lock not owned by current process
	ErrLockNotOwned = errors.New("lock not owned by current process")
)

// DistributedLock represents a distributed lock
type DistributedLock struct {
	client *Client
	key    string
	value  string
	ttl    time.Duration
}

// LockOptions contains options for acquiring a lock
type LockOptions struct {
	TTL         time.Duration // Lock TTL
	RetryDelay  time.Duration // Delay between retry attempts
	MaxRetries  int           // Maximum number of retry attempts
	LockTimeout time.Duration // Total timeout for acquiring the lock
}

// DefaultLockOptions returns default lock options
func DefaultLockOptions() LockOptions {
	return LockOptions{
		TTL:         DefaultLockTTL,
		RetryDelay:  DefaultRetryDelay,
		MaxRetries:  DefaultMaxRetries,
		LockTimeout: DefaultLockTimeout,
	}
}

// TryLock attempts to acquire a distributed lock without retries
func (c *Client) TryLock(ctx context.Context, key string, ttl time.Duration) (*DistributedLock, error) {
	value := generateLockValue()

	// Use SET with NX (only if not exists) and EX (expiration)
	result, err := c.rdb.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !result {
		return nil, ErrLockNotAcquired
	}

	return &DistributedLock{
		client: c,
		key:    key,
		value:  value,
		ttl:    ttl,
	}, nil
}

// Lock acquires a distributed lock with retry logic
func (c *Client) Lock(ctx context.Context, key string, opts ...LockOptions) (*DistributedLock, error) {
	options := DefaultLockOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Create a context with timeout for the entire lock acquisition process
	lockCtx, cancel := context.WithTimeout(ctx, options.LockTimeout)
	defer cancel()

	var lastErr error
	retries := 0

	for {
		select {
		case <-lockCtx.Done():
			if lastErr != nil {
				return nil, fmt.Errorf("lock acquisition timeout: %w", lastErr)
			}
			return nil, fmt.Errorf("lock acquisition timeout: %w", lockCtx.Err())
		default:
		}

		lock, err := c.TryLock(lockCtx, key, options.TTL)
		if err == nil {
			return lock, nil
		}

		if !errors.Is(err, ErrLockNotAcquired) {
			return nil, err
		}

		lastErr = err
		retries++

		if retries > options.MaxRetries {
			return nil, fmt.Errorf("max retries exceeded: %w", ErrLockNotAcquired)
		}

		// Wait before retrying
		select {
		case <-lockCtx.Done():
			return nil, fmt.Errorf("lock acquisition timeout during retry: %w", lockCtx.Err())
		case <-time.After(options.RetryDelay):
		}
	}
}

// Unlock releases the distributed lock
func (lock *DistributedLock) Unlock(ctx context.Context) error {
	// Lua script to ensure we only delete the lock if we own it
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := lock.client.rdb.Eval(ctx, script, []string{lock.key}, lock.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotOwned
	}

	return nil
}

// Extend extends the lock's TTL
func (lock *DistributedLock) Extend(ctx context.Context, ttl time.Duration) error {
	// Lua script to extend TTL only if we own the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := lock.client.rdb.Eval(ctx, script, []string{lock.key}, lock.value, int64(ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotOwned
	}

	lock.ttl = ttl
	return nil
}

// IsLocked checks if the lock is still held by this process
func (lock *DistributedLock) IsLocked(ctx context.Context) (bool, error) {
	value, err := lock.client.rdb.Get(ctx, lock.key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return value == lock.value, nil
}

// GetTTL returns the remaining TTL of the lock
func (lock *DistributedLock) GetTTL(ctx context.Context) (time.Duration, error) {
	return lock.client.rdb.TTL(ctx, lock.key).Result()
}

// Key returns the lock key
func (lock *DistributedLock) Key() string {
	return lock.key
}

// Value returns the lock value
func (lock *DistributedLock) Value() string {
	return lock.value
}

// WithLock executes a function while holding a distributed lock
func (c *Client) WithLock(ctx context.Context, key string, fn func() error, opts ...LockOptions) error {
	lock, err := c.Lock(ctx, key, opts...)
	if err != nil {
		return err
	}

	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			// Log the unlock error but don't return it
			// You might want to integrate with your logger here
		}
	}()

	return fn()
}

// generateLockValue generates a unique value for the lock
func generateLockValue() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based value
		return fmt.Sprintf("lock_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
