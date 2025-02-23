package messaging

import (
	"context"
	"encoding/json"
	"time"

	"shenanigigs/common/telemetry"
	"shenanigigs/ingestion/internal/config"
	"shenanigigs/ingestion/internal/errors"
	"shenanigigs/ingestion/internal/models"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var tracer = telemetry.GetTracer("shenanigigs/ingestion/messaging")

const (
	JobPostingsSubject = "jobs.new"
)

type Publisher interface {
	PublishJobPosting(ctx context.Context, posting *models.JobPosting) error
	Close()
}

type natsPublisher struct {
	conn   *nats.Conn
	logger *zap.Logger
}

func NewPublisher(logger *zap.Logger, config *config.Config) (Publisher, error) {
	opts := []nats.Option{
		nats.Timeout(config.NATSConnTimeout),
		nats.ReconnectWait(time.Second),
		nats.MaxReconnects(-1),
	}

	conn, err := nats.Connect(config.NATSURL, opts...)
	if err != nil {
		return nil, errors.Internal("connecting to NATS", err)
	}

	return &natsPublisher{
		conn:   conn,
		logger: logger,
	}, nil
}

func (p *natsPublisher) PublishJobPosting(ctx context.Context, posting *models.JobPosting) error {
	_, span := tracer.Start(ctx, "PublishJobPosting")
	defer span.End()

	data, err := json.Marshal(posting)
	if err != nil {
		span.RecordError(err)
		return errors.Internal("marshaling job posting", err)
	}

	span.SetAttributes(
		telemetry.String("nats.subject", JobPostingsSubject),
		telemetry.Int("message.size", len(data)),
	)

	if err := p.conn.Publish(JobPostingsSubject, data); err != nil {
		span.RecordError(err)
		p.logger.Error("failed to publish job posting",
			zap.String("id", posting.ID),
			zap.Error(err))
		return errors.Internal("publishing to NATS", err)
	}

	p.logger.Debug("published job posting",
		zap.String("id", posting.ID),
		zap.String("subject", JobPostingsSubject))
	return nil
}

func (p *natsPublisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}
