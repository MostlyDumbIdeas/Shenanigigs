package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shenanigigs/ingestion/internal/api"
	"shenanigigs/ingestion/internal/config"
	"shenanigigs/ingestion/internal/messaging"
	"shenanigigs/ingestion/internal/scheduler"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	cfg := &config.Config{
		HNAPIBaseURL:       "https://hacker-news.firebaseio.com/v0",
		HNSearchAPIBaseURL: "https://hn.algolia.com/api/v1",
		HNAPITimeout:       10 * time.Second,
		PollingInterval:    30 * time.Second,
	}

	logger.Info("starting ingestion service",
		zap.String("hn_api_url", cfg.HNAPIBaseURL),
		zap.Duration("api_timeout", cfg.HNAPITimeout),
		zap.Duration("polling_interval", cfg.PollingInterval))

	hnClient := api.NewJobSourceClient(logger, cfg)

	publisher, err := messaging.NewPublisher(logger, cfg)
	if err != nil {
		logger.Fatal("failed to create NATS publisher", zap.Error(err))
	}
	defer publisher.Close()

	jobScheduler := scheduler.NewJobScheduler(hnClient, publisher, logger, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := jobScheduler.Start(ctx); err != nil {
			logger.Error("job scheduler failed", zap.Error(err))
		}
	}()

	logger.Info("ingestion service started successfully")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down...")
	jobScheduler.Stop()
	logger.Info("shutdown complete")
}
