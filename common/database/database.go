package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

type Options struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	Username        string
	Password        string
	Database        string
}

type Database struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

func New(ctx context.Context, opts Options, logger *zap.Logger) (*Database, error) {
	hostAndParams := strings.Split(opts.DSN, "?")
	host := hostAndParams[0]

	conn, err := clickhouse.Open(&clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{host},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Auth: clickhouse.Auth{
			Database: opts.Database,
			Username: opts.Username,
			Password: opts.Password,
		},
		DialTimeout:     time.Second * 30,
		MaxOpenConns:    opts.MaxOpenConns,
		MaxIdleConns:    opts.MaxIdleConns,
		ConnMaxLifetime: opts.ConnMaxLifetime,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return &Database{
		conn:   conn,
		logger: logger,
	}, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) Conn() clickhouse.Conn {
	return db.conn
}
