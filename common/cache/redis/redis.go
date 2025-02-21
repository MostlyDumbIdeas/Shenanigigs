package redis

import (
	"context"
	"encoding"
	"time"

	"shenanigigs/common/cache"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(opts cache.Options) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.RedisURL,
		Password: opts.RedisPassword,
		DB:       opts.RedisDB,
	})

	return &Cache{client: client}
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		return c.client.Set(ctx, key, value, cache.DefaultOptions().DefaultTTL).Err()
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string, value interface{}) error {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return cache.ErrNotFound
	}
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case *string:
		*v = string(val)
	case encoding.BinaryUnmarshaler:
		return v.UnmarshalBinary(val)
	default:
		return cache.ErrInvalidValue
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) Clear(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
