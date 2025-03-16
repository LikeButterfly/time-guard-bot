package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Implements Storage using Redis
type RedisStorage struct {
	client *redis.Client
}

// Creates a new Redis storage
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) { // FIXME? addr, password, db in one struct? I think no
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// Closes the Redis connection
func (rs *RedisStorage) Close() error {
	return rs.client.Close()
}
