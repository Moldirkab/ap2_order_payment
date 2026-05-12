package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

func (r *RedisCache) Set(key string, value string, ttl int) error {
	return r.client.Set(
		context.Background(),
		key,
		value,
		time.Duration(ttl)*time.Second,
	).Err()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(context.Background(), key).Err()
}
