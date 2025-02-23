package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	NATSURL         string
	NATSConnTimeout time.Duration

	ClickHouseDSN          string
	ClickHouseMaxOpenConns int
	ClickHouseMaxIdleConns int
	ClickHouseConnMaxLife  time.Duration
	ClickHouseUsername     string
	ClickHousePassword     string
	ClickHouseDatabase     string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration

	BatchSize         int
	ProcessingTimeout time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
}

func LoadConfig() (*Config, error) {
	config := &Config{
		NATSURL:         getEnvString("NATS_URL", "nats://localhost:4222"),
		NATSConnTimeout: getEnvDuration("NATS_CONN_TIMEOUT", 10*time.Second),

		ClickHouseDSN:          getEnvString("CLICKHOUSE_DSN", "localhost:9000"),
		ClickHouseMaxOpenConns: getEnvInt("CLICKHOUSE_MAX_OPEN_CONNS", 10),
		ClickHouseMaxIdleConns: getEnvInt("CLICKHOUSE_MAX_IDLE_CONNS", 5),
		ClickHouseConnMaxLife:  getEnvDuration("CLICKHOUSE_CONN_MAX_LIFE", time.Hour),
		ClickHouseUsername:     getEnvString("CLICKHOUSE_USERNAME", "default"),
		ClickHousePassword:     getEnvString("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDatabase:     getEnvString("CLICKHOUSE_DATABASE", "shenanigigs"),

		RedisAddr:     getEnvString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnvString("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		CacheTTL:      getEnvDuration("CACHE_TTL", 24*time.Hour),

		BatchSize:         getEnvInt("BATCH_SIZE", 100),
		ProcessingTimeout: getEnvDuration("PROCESSING_TIMEOUT", 5*time.Minute),
		MaxRetries:        getEnvInt("MAX_RETRIES", 3),
		RetryDelay:        getEnvDuration("RETRY_DELAY", 30*time.Second),
	}

	return config, nil
}

func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
