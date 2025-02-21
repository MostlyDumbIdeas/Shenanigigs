package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"shenanigigs/common/telemetry"
	"shenanigigs/ingestion/internal/errors"
	"shenanigigs/ingestion/internal/models"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	jobPostingSubject = "job.posting.fetched"
	connectTimeout    = 10 * time.Second
)

type JobPostingEvent struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PostedAt    time.Time `json:"posted_at"`
	RawText     string    `json:"raw_text"`
}

type Publisher struct {
	nc     *nats.Conn
	logger *zap.Logger
}

func NewPublisher(natsURL string, logger *zap.Logger) (*Publisher, error) {
	opts := []nats.Option{
		nats.Timeout(connectTimeout),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
	}

	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	return &Publisher{
		nc:     nc,
		logger: logger,
	}, nil
}

func (p *Publisher) PublishJobPosting(ctx context.Context, post *models.JobPosting) error {
	_, span := tracer.Start(ctx, "PublishJobPosting")
	defer span.End()

	event := JobPostingEvent{
		ID:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		PostedAt:    post.PostedAt,
		RawText:     post.RawText,
	}

	data, err := json.Marshal(event)
	if err != nil {
		span.RecordError(err)
		return errors.Internal("marshaling event", err)
	}

	span.SetAttributes(
		telemetry.String("nats.subject", jobPostingSubject),
		telemetry.Int("message.size", len(data)),
	)

	if err := p.nc.Publish(jobPostingSubject, data); err != nil {
		span.RecordError(err)
		p.logger.Error("failed to publish job posting",
			zap.Int("id", event.ID),
			zap.Error(err))
		return errors.Internal("publishing event", err)
	}

	p.logger.Debug("published job posting event",
		zap.Int("id", event.ID),
		zap.String("title", event.Title),
		zap.Time("posted_at", event.PostedAt))

	return nil
}

var tracer = telemetry.GetTracer("shenanigigs/ingestion/events")

func (p *Publisher) Close() error {
	if p.nc != nil {
		p.nc.Close()
	}
	return nil
}
