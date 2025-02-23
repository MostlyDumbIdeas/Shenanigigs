package processor

import (
	"context"
	"fmt"

	"shenanigigs/common/telemetry"
	"shenanigigs/processing/internal/config"
	"shenanigigs/processing/internal/models"
	"shenanigigs/processing/internal/parser"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type JobProcessor struct {
	logger *zap.Logger
	db     clickhouse.Conn
	nats   *nats.Conn
	tracer trace.Tracer
	config *config.Config
}

func NewJobProcessor(logger *zap.Logger, db clickhouse.Conn, nc *nats.Conn, config *config.Config) *JobProcessor {
	tracer := telemetry.GetTracer("shenanigigs/processing/processor")
	return &JobProcessor{
		logger: logger,
		db:     db,
		nats:   nc,
		tracer: tracer,
		config: config,
	}
}

func (p *JobProcessor) ProcessJobPosting(ctx context.Context, rawData []byte) error {
	ctx, span := p.tracer.Start(ctx, "ProcessJobPosting")
	defer span.End()

	parsedPosting, err := parser.ParseJobPosting(string(rawData))
	if err != nil {
		p.logger.Error("Failed to parse job posting", zap.Error(err))
		return fmt.Errorf("parse job posting: %w", err)
	}

	if err := p.storeJobPosting(ctx, parsedPosting); err != nil {
		p.logger.Error("Failed to store job posting", zap.Error(err))
		return fmt.Errorf("store job posting: %w", err)
	}

	return nil
}

func (p *JobProcessor) storeJobPosting(ctx context.Context, posting *models.JobPosting) error {
	query := `
		INSERT INTO jobs (
			id, title, company, location, description, technologies,
			experience_level, compensation_min, compensation_max,
			compensation_currency, compensation_period, remote_policy,
			source, source_url, created_at, updated_at, raw_data
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	if err := p.db.Exec(ctx, query,
		posting.ID,
		posting.Title,
		posting.Company,
		posting.Location,
		posting.Description,
		posting.Technologies,
		posting.ExperienceLevel,
		posting.CompensationMin,
		posting.CompensationMax,
		posting.CompensationCurrency,
		posting.CompensationPeriod,
		posting.RemotePolicy,
		posting.Source,
		posting.SourceURL,
		posting.CreatedAt,
		posting.UpdatedAt,
		posting.RawData,
	); err != nil {
		return fmt.Errorf("insert job posting: %w", err)
	}

	return nil
}
