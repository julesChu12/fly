package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/julesChu12/fly/mora/pkg/cache"
)

// Adapter mora缓存适配器
type Adapter struct {
	client *cache.Client
	prefix string
}

// NewAdapter 创建缓存适配器
func NewAdapter(redisAddr, redisPassword string, redisDB int, prefix string) (*Adapter, error) {
	config := cache.Config{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	client := cache.New(config)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, err
	}

	return &Adapter{
		client: client,
		prefix: prefix,
	}, nil
}

// Set 设置缓存值
func (a *Adapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cacheKey := a.buildKey(key)

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return a.client.Set(ctx, cacheKey, data, ttl)
}

// Get 获取缓存值
func (a *Adapter) Get(ctx context.Context, key string, dest interface{}) error {
	cacheKey := a.buildKey(key)

	data, err := a.client.Get(ctx, cacheKey)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// Delete 删除缓存
func (a *Adapter) Delete(ctx context.Context, key string) error {
	cacheKey := a.buildKey(key)
	return a.client.Delete(ctx, cacheKey)
}

// Exists 检查缓存是否存在
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := a.buildKey(key)
	return a.client.Exists(ctx, cacheKey)
}

// buildKey 构建完整的缓存键
func (a *Adapter) buildKey(key string) string {
	if a.prefix != "" {
		return a.prefix + ":" + key
	}
	return key
}

// Close 关闭连接
func (a *Adapter) Close() error {
	return a.client.Close()
}