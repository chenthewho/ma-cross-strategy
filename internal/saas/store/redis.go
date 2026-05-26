// Package store provides Redis client for caching.
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps go-redis for cache operations.
// Used for: champion gene cache, session cache, AI signal cache.
// NOT used for: signal passing between components.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis connection.
func NewRedisClient(addr string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// Get retrieves a value from cache.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // cache miss
	}
	return val, err
}

// Set stores a value in cache with optional TTL.
func (r *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Del removes a key from cache.
func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// GetJSON retrieves and unmarshals a JSON value from cache.
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	val, err := r.Get(ctx, key)
	if err != nil || val == "" {
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("unmarshal cache: %w", err)
	}
	return true, nil
}

// SetJSON marshals and stores a JSON value in cache.
func (r *RedisClient) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	return r.Set(ctx, key, string(data), ttl)
}

// Close shuts down the Redis connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}
