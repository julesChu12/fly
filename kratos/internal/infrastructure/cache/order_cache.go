package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/julesChu12/fly/mora/pkg/cache"
	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *cache.Client
}

func NewOrderCache(redisAddr string, db int) (*OrderCache, error) {
	cfg := cache.Config{
		Addr: redisAddr,
		DB:   db,
	}

	client := cache.New(cfg)
	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &OrderCache{client: client}, nil
}

func (c *OrderCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil // Cache miss
		}
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (c *OrderCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, string(data), ttl)
}

func (c *OrderCache) Delete(ctx context.Context, key string) error {
	return c.client.Delete(ctx, key)
}

func (c *OrderCache) Close() error {
	return c.client.Close()
}

// Cache key builders
func OrderCacheKey(id uint) string {
	return fmt.Sprintf("order:%d", id)
}

func OrderListCacheKey(tenantID uint, filters string) string {
	return fmt.Sprintf("orders:%d:%s", tenantID, filters)
}
