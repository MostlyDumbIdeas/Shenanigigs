package main

import (
	"context"
	"log"

	"shenanigigs/common/database/schema"
	"shenanigigs/common/database/schema/migrations"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{
			Database: "shenanigigs",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		logger.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}
	defer conn.Close()

	ctx := context.Background()

	migrator := schema.NewMigrator(conn, logger)

	if err := migrator.CreateMigrationsTable(ctx); err != nil {
		logger.Fatal("Failed to create migrations table", zap.Error(err))
	}

	applied, err := migrator.GetAppliedMigrations(ctx)
	if err != nil {
		logger.Fatal("Failed to get applied migrations", zap.Error(err))
	}

	migrations := []schema.Migration{
		migrations.CreateJobsTable,
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; !ok {
			logger.Info("Applying migration",
				zap.Int("version", migration.Version),
				zap.String("description", migration.Description),
			)

			if err := migrator.ApplyMigration(ctx, migration); err != nil {
				logger.Fatal("Failed to apply migration",
					zap.Int("version", migration.Version),
					zap.Error(err),
				)
			}

			logger.Info("Successfully applied migration",
				zap.Int("version", migration.Version),
			)
		} else {
			logger.Info("Migration already applied",
				zap.Int("version", migration.Version),
				zap.String("description", migration.Description),
			)
		}
	}

	logger.Info("All migrations completed successfully")
}
