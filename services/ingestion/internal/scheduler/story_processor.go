package scheduler

import (
	"context"
	"strings"
	"sync/atomic"

	"shenanigigs/ingestion/internal/models"

	"go.uber.org/zap"
)

type storyProcessor struct {
	scheduler *JobScheduler
	logger    *zap.Logger
}

func newStoryProcessor(scheduler *JobScheduler, logger *zap.Logger) *storyProcessor {
	return &storyProcessor{
		scheduler: scheduler,
		logger:    logger,
	}
}

func (p *storyProcessor) processStory(ctx context.Context, id int, stats *jobProcessingStats, commentChan chan int) {
	post, err := p.scheduler.hnClient.GetItem(ctx, id)
	if err != nil {
		p.logger.Error("failed to fetch story", zap.Int("id", id), zap.Error(err))
		return
	}

	if p.isHiringThread(post) {
		atomic.AddInt32(&stats.hiringThreadsFound, 1)
		p.logger.Info("found hiring thread",
			zap.Int("id", post.ID),
			zap.String("title", post.Title),
			zap.String("author", post.By),
			zap.Int64("time", post.Time),
			zap.Int("comments_count", len(post.Kids)))

		for _, commentID := range post.Kids {
			commentChan <- commentID
		}
	}
}

func (p *storyProcessor) isHiringThread(post *models.SourcePost) bool {
	title := strings.ToLower(post.Title)
	return (strings.Contains(title, "who is hiring?") || strings.Contains(title, "ask hn: who is hiring?")) &&
		(post.By == "whoishiring")
}

func (p *storyProcessor) feedStories(stories []int, storyChan chan int) {
	for _, id := range stories {
		storyChan <- id
	}
	close(storyChan)
}
