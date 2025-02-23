package schema

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

type Migrator struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

func NewMigrator(conn clickhouse.Conn, logger *zap.Logger) *Migrator {
	return &Migrator{
		conn:   conn,
		logger: logger,
	}
}

func (m *Migrator) CreateMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version Int32,
			description String,
			applied_at DateTime,
			PRIMARY KEY (version)
		) ENGINE = MergeTree()
	`

	if err := m.conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

func (m *Migrator) GetAppliedMigrations(ctx context.Context) (map[int]time.Time, error) {
	query := "SELECT version, applied_at FROM migrations ORDER BY version"

	rows, err := m.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]time.Time)
	for rows.Next() {
		var version int
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		applied[version] = appliedAt
	}

	return applied, nil
}

func (m *Migrator) ApplyMigration(ctx context.Context, migration Migration) error {
	if err := m.conn.Exec(ctx, migration.Up); err != nil {
		return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
	}

	if err := m.conn.Exec(ctx, `
		INSERT INTO migrations (version, description, applied_at)
		VALUES (?, ?, now())
	`, migration.Version, migration.Description); err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	return nil
}

func (m *Migrator) RollbackMigration(ctx context.Context, migration Migration) error {
	if err := m.conn.Exec(ctx, migration.Down); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
	}

	if err := m.conn.Exec(ctx, "DELETE FROM migrations WHERE version = ?", migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
	}

	return nil
}
