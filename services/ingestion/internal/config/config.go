package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HNAPIBaseURL       string
	HNSearchAPIBaseURL string
	HNAPITimeout       time.Duration

	PollingInterval time.Duration
	MaxRetries      int
	RetryDelay      time.Duration

	NATSURL         string
	NATSConnTimeout time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration
}

func LoadConfig() (*Config, error) {
	config := &Config{
		HNAPIBaseURL:       getEnvString("HN_API_BASE_URL", "https://hacker-news.firebaseio.com/v0"),
		HNSearchAPIBaseURL: getEnvString("HN_SEARCH_API_BASE_URL", "https://hn.algolia.com/api/v1"),
		HNAPITimeout:       getEnvDuration("HN_API_TIMEOUT", 10*time.Second),
		PollingInterval:    getEnvDuration("POLLING_INTERVAL", 15*time.Minute),
		MaxRetries:         getEnvInt("MAX_RETRIES", 3),
		RetryDelay:         getEnvDuration("RETRY_DELAY", 30*time.Second),
		NATSURL:            getEnvString("NATS_URL", "nats://localhost:4222"),
		NATSConnTimeout:    getEnvDuration("NATS_CONN_TIMEOUT", 10*time.Second),

		RedisAddr:     getEnvString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnvString("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		CacheTTL:      getEnvDuration("CACHE_TTL", 24*time.Hour),
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
