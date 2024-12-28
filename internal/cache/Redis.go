package cache

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	Client *redis.Client
}

// InitializeCache initializes a Redis client and tests the connection.
func InitializeCache(addr string, password string, db int) (*RedisCache, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // If no password, set it to an empty string ""
		DB:       db,
	})

	// Test the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{Client: client}, nil
}

// Set cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := r.Client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Printf("Failed to set cache for key: %s, error: %v", key, err)
	}
	return err
}

// Get cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Cache miss for key: %s", key)
	} else if err != nil {
		log.Printf("Error fetching key %s from cache: %v", key, err)
	} else {
		log.Printf("Cache hit for key: %s", key)
	}
	return result, err
}

// Delete cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Failed to delete cache for key: %s, error: %v", key, err)
	}
	return err
}