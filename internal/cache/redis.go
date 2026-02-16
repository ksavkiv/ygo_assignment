package cache

import (
	"context"
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("cache get %s: %w", key, err)
	}

	var d destination.Destination
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("cache unmarshal %s: %w", key, err)
	}
	return &d, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, d *destination.Destination) error {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("cache marshal %s: %w", key, err)
	}
	return c.client.Set(ctx, key, data, defaultTTL).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
