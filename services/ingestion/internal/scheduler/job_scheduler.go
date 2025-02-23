package scheduler

import (
	"context"
	"sync"
	"time"

	"shenanigigs/common/telemetry"
	"shenanigigs/ingestion/internal/api"
	"shenanigigs/ingestion/internal/config"
	"shenanigigs/ingestion/internal/errors"
	"shenanigigs/ingestion/internal/messaging"

	"go.uber.org/zap"
)

var tracer = telemetry.GetTracer("shenanigigs/ingestion/scheduler")

type JobScheduler struct {
	hnClient       api.JobSourceClient
	publisher      messaging.Publisher
	logger         *zap.Logger
	config         *config.Config
	mutex          sync.Mutex
	isActive       bool
	workerManager  *workerManager
	storyProcessor *storyProcessor
}

func NewJobScheduler(hnClient api.JobSourceClient, publisher messaging.Publisher, logger *zap.Logger, config *config.Config) *JobScheduler {
	scheduler := &JobScheduler{
		hnClient:  hnClient,
		publisher: publisher,
		logger:    logger,
		config:    config,
	}
	scheduler.workerManager = newWorkerManager(scheduler, logger)
	scheduler.storyProcessor = newStoryProcessor(scheduler, logger)
	return scheduler
}

func (s *JobScheduler) Start(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "JobScheduler.Start")
	defer span.End()

	s.mutex.Lock()
	if s.isActive {
		s.mutex.Unlock()
		return nil
	}
	s.isActive = true
	s.mutex.Unlock()

	ticker := time.NewTicker(s.config.PollingInterval)
	defer ticker.Stop()

	if err := s.fetchWhoIsHiring(ctx); err != nil {
		s.logger.Error("initial fetch failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.fetchWhoIsHiring(ctx); err != nil {
				s.logger.Error("periodic fetch failed", zap.Error(err))
			}
		}
	}
}

func (s *JobScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.isActive = false
}

type jobProcessingStats struct {
	hiringThreadsFound int32
	commentsProcessed  int32
}

func (s *JobScheduler) fetchWhoIsHiring(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "JobScheduler.fetchWhoIsHiring")
	defer span.End()

	s.logger.Info("starting to fetch who is hiring posts")
	stories, err := s.hnClient.SearchHiringThreads(ctx)
	if err != nil {
		span.RecordError(err)
		return errors.Internal("failed to search hiring threads", err)
	}
	span.SetAttributes(telemetry.Int("stories.count", len(stories)))
	s.logger.Info("found hiring threads", zap.Int("count", len(stories)))

	stats := &jobProcessingStats{}
	storyChan := make(chan int)
	commentChan := make(chan int)
	doneChan := make(chan bool)

	wg := s.startWorkers(ctx, stats, storyChan, commentChan, doneChan)

	go s.feedStories(stories, storyChan)

	go func() {
		wg.Wait()
		close(commentChan)
		close(doneChan)
	}()

	return s.waitForCompletion(ctx, doneChan, stats)
}

func (s *JobScheduler) startWorkers(ctx context.Context, stats *jobProcessingStats, storyChan chan int, commentChan chan int, doneChan chan bool) *sync.WaitGroup {
	return s.workerManager.startWorkers(ctx, stats, storyChan, commentChan, doneChan)
}

func (s *JobScheduler) processStory(ctx context.Context, id int, stats *jobProcessingStats, commentChan chan int) {
	s.storyProcessor.processStory(ctx, id, stats, commentChan)
}

func (s *JobScheduler) feedStories(stories []int, storyChan chan int) {
	s.storyProcessor.feedStories(stories, storyChan)
}

func (s *JobScheduler) waitForCompletion(ctx context.Context, doneChan chan bool, stats *jobProcessingStats) error {
	ctx, span := tracer.Start(ctx, "JobScheduler.waitForCompletion")
	defer span.End()

	select {
	case <-ctx.Done():
		span.RecordError(ctx.Err())
		return ctx.Err()
	case <-doneChan:
		span.SetAttributes(
			telemetry.Int("hiring_threads_found", int(stats.hiringThreadsFound)),
			telemetry.Int("comments_processed", int(stats.commentsProcessed)),
		)
		s.logger.Info("completed fetching who is hiring posts",
			zap.Int("hiring_threads_found", int(stats.hiringThreadsFound)),
			zap.Int("comments_processed", int(stats.commentsProcessed)))
		return nil
	}
}

func (s *JobScheduler) processComment(ctx context.Context, commentID int) error {
	ctx, span := tracer.Start(ctx, "JobScheduler.processComment")
	span.SetAttributes(telemetry.Int("comment_id", commentID))
	defer span.End()

	comment, err := s.hnClient.GetItem(ctx, commentID)
	if err != nil {
		span.RecordError(err)
		return errors.Internal("failed to fetch comment", err)
	}

	jobPosting := comment.ToJobPosting()
	s.logger.Debug("processing job posting",
		zap.String("comment_id", jobPosting.ID),
		zap.Time("posted_at", jobPosting.PostedAt))

	if err := s.publisher.PublishJobPosting(ctx, jobPosting); err != nil {
		span.RecordError(err)
		return errors.Internal("failed to publish job posting", err)
	}

	return nil
}
