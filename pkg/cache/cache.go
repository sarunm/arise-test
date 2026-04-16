package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type redisCache struct {
	client *redis.Client
}

func New() Cache {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT"),
		),
		PoolSize:        envInt("REDIS_POOL_SIZE", 20),
		MinIdleConns:    envInt("REDIS_MIN_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Duration(envInt("REDIS_CONN_MAX_LIFETIME_MIN", 30)) * time.Minute,
		ConnMaxIdleTime: time.Duration(envInt("REDIS_CONN_MAX_IDLE_TIME_MIN", 5)) * time.Minute,
	})
	return &redisCache{client: client}
}

func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}
