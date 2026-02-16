package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"destination-data-aggregation-api/internal/destination"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 5 * time.Minute

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (*destination.Destination, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		slog.Error("redis: failed to get key", "key", key, "error", err)
		return nil, fmt.Errorf("cache get %s: %w", key, err)
	}

	var d destination.Destination
	if err := json.Unmarshal(data, &d); err != nil {
		slog.Error("redis: failed to unmarshal cached value", "key", key, "error", err)
		return nil, fmt.Errorf("cache unmarshal %s: %w", key, err)
	}
	return &d, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, d *destination.Destination) error {
	data, err := json.Marshal(d)
	if err != nil {
		slog.Error("redis: failed to marshal value for cache", "key", key, "error", err)
		return fmt.Errorf("cache marshal %s: %w", key, err)
	}
	if err := c.client.Set(ctx, key, data, defaultTTL).Err(); err != nil {
		slog.Error("redis: failed to set key", "key", key, "error", err)
		return err
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		slog.Error("redis: failed to delete key", "key", key, "error", err)
		return err
	}
	return nil
}
