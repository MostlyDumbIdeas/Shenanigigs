package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("key not found in cache")
	ErrInvalidValue = errors.New("invalid value for cache")
	ErrClosed       = errors.New("cache is closed")
	ErrInvalidKey   = errors.New("invalid cache key")
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	Get(ctx context.Context, key string, value interface{}) error

	Delete(ctx context.Context, key string) error

	Clear(ctx context.Context) error

	Close() error
}

type Options struct {
	DefaultTTL time.Duration

	CleanupInterval time.Duration

	RedisURL string

	RedisPassword string

	RedisDB int
}

func DefaultOptions() Options {
	return Options{
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute * 5,
	}
}
