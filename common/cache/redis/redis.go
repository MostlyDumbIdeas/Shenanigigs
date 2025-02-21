package redis

import (
	"context"
	"encoding"
	"time"

	"shenanigigs/common/cache"
	"shenanigigs/common/telemetry"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type Cache struct {
	client *redis.Client
	tracer trace.Tracer
}

func New(opts cache.Options) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.RedisURL,
		Password: opts.RedisPassword,
		DB:       opts.RedisDB,
	})

	return &Cache{
		client: client,
		tracer: telemetry.GetTracer("redis-cache"),
	}
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	ctx, span := c.tracer.Start(ctx, "cache.Set")
	defer func() {
		span.SetAttributes(telemetry.Int("cache.operation.duration_ms", int(time.Since(start).Milliseconds())))
		span.End()
	}()

	span.SetAttributes(
		telemetry.String("cache.key", key),
		telemetry.String("cache.operation", "set"),
		telemetry.Int("cache.ttl_seconds", int(ttl.Seconds())),
	)

	if ttl == 0 {
		err := c.client.Set(ctx, key, value, cache.DefaultOptions().DefaultTTL).Err()
		if err != nil {
			span.RecordError(err)
		}
		return err
	}

	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (c *Cache) Get(ctx context.Context, key string, value interface{}) error {
	start := time.Now()
	ctx, span := c.tracer.Start(ctx, "cache.Get")
	defer func() {
		span.SetAttributes(telemetry.Int("cache.operation.duration_ms", int(time.Since(start).Milliseconds())))
		span.End()
	}()

	span.SetAttributes(
		telemetry.String("cache.key", key),
		telemetry.String("cache.operation", "get"),
	)

	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		span.SetAttributes(telemetry.String("cache.result", "miss"))
		return cache.ErrNotFound
	}
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.SetAttributes(telemetry.String("cache.result", "hit"))

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
	start := time.Now()
	ctx, span := c.tracer.Start(ctx, "cache.Delete")
	defer func() {
		span.SetAttributes(telemetry.Int("cache.operation.duration_ms", int(time.Since(start).Milliseconds())))
		span.End()
	}()

	span.SetAttributes(
		telemetry.String("cache.key", key),
		telemetry.String("cache.operation", "delete"),
	)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (c *Cache) Clear(ctx context.Context) error {
	start := time.Now()
	ctx, span := c.tracer.Start(ctx, "cache.Clear")
	defer func() {
		span.SetAttributes(telemetry.Int("cache.operation.duration_ms", int(time.Since(start).Milliseconds())))
		span.End()
	}()

	span.SetAttributes(telemetry.String("cache.operation", "clear"))

	err := c.client.FlushDB(ctx).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (c *Cache) Close() error {
	return c.client.Close()
}
