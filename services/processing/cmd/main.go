package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"shenanigigs/common/database"
	"shenanigigs/common/telemetry"
	"shenanigigs/processing/internal/config"
	"shenanigigs/processing/internal/events"
	"shenanigigs/processing/internal/processor"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newLogger(cfg *config.Config) (*zap.Logger, error) {
	return zap.NewProduction()
}

func newNATSConnection(cfg *config.Config) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Timeout(cfg.NATSConnTimeout),
		nats.Name("processing-service"),
		nats.RetryOnFailedConnect(true),
	}
	return nats.Connect(cfg.NATSURL, opts...)
}

func newClickHouseConnection(cfg *config.Config, logger *zap.Logger) (clickhouse.Conn, error) {
	db, err := database.New(context.Background(), database.Options{
		DSN:             cfg.ClickHouseDSN,
		MaxOpenConns:    cfg.ClickHouseMaxOpenConns,
		MaxIdleConns:    cfg.ClickHouseMaxIdleConns,
		ConnMaxLifetime: cfg.ClickHouseConnMaxLife,
		Username:        cfg.ClickHouseUsername,
		Password:        cfg.ClickHousePassword,
		Database:        cfg.ClickHouseDatabase,
	}, logger)
	if err != nil {
		return nil, err
	}
	return db.Conn(), nil
}

func newTracer() trace.Tracer {
	return telemetry.GetTracer("shenanigigs/processing")
}

func main() {
	app := fx.New(
		fx.Provide(
			config.LoadConfig,
			newLogger,
			newNATSConnection,
			newClickHouseConnection,
			processor.NewJobProcessor,
			events.NewHandler,
			newTracer,
		),
		fx.Invoke(
			func(handler *events.Handler, lc fx.Lifecycle) error {
				return handler.RegisterSubscriptions(lc)
			},
		),
	)

	startCtx := context.Background()
	if err := app.Start(startCtx); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	stopCtx := context.Background()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}
