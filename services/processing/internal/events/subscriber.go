package events

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"shenanigigs/processing/internal/processor"
)

type Handler struct {
	logger       *zap.Logger
	nc           *nats.Conn
	tracer       trace.Tracer
	jobProcessor *processor.JobProcessor
	sub          *nats.Subscription
}

func NewHandler(logger *zap.Logger, nc *nats.Conn, tracer trace.Tracer, jobProcessor *processor.JobProcessor) *Handler {
	return &Handler{
		logger:       logger,
		nc:           nc,
		tracer:       tracer,
		jobProcessor: jobProcessor,
	}
}

func (h *Handler) RegisterSubscriptions(lc fx.Lifecycle) error {
	sub, err := h.nc.QueueSubscribe("jobs.new", "processing-service", h.handleJobPosting)
	if err != nil {
		return fmt.Errorf("subscribe to jobs.new: %w", err)
	}

	h.sub = sub
	h.logger.Info("Registered NATS subscriptions")

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return h.sub.Unsubscribe()
		},
	})

	return nil
}

func (h *Handler) handleJobPosting(msg *nats.Msg) {
	ctx, span := h.tracer.Start(context.Background(), "handleJobPosting")
	defer span.End()

	if err := h.jobProcessor.ProcessJobPosting(ctx, msg.Data); err != nil {
		h.logger.Error("Failed to process job posting",
			zap.Error(err),
			zap.String("subject", msg.Subject),
		)
		return
	}

	h.logger.Info("Successfully processed job posting",
		zap.String("subject", msg.Subject),
	)
}
