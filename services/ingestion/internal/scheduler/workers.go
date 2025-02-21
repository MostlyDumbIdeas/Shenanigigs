package scheduler

import (
	"context"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

type workerManager struct {
	scheduler *JobScheduler
	logger    *zap.Logger
}

func newWorkerManager(scheduler *JobScheduler, logger *zap.Logger) *workerManager {
	return &workerManager{
		scheduler: scheduler,
		logger:    logger,
	}
}

func (w *workerManager) startWorkers(ctx context.Context, stats *jobProcessingStats, storyChan chan int, commentChan chan int, doneChan chan bool) *sync.WaitGroup {
	var wg sync.WaitGroup

	w.startCommentWorkers(ctx, &wg, stats, commentChan)
	w.startStoryWorkers(ctx, &wg, stats, storyChan, commentChan)

	return &wg
}

func (w *workerManager) startCommentWorkers(ctx context.Context, wg *sync.WaitGroup, stats *jobProcessingStats, commentChan chan int) {
	const numWorkers = 10
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for commentID := range commentChan {
				if err := w.scheduler.processComment(ctx, commentID); err != nil {
					w.logger.Error("failed to process comment",
						zap.Int("comment_id", commentID),
						zap.Error(err))
					continue
				}
				atomic.AddInt32(&stats.commentsProcessed, 1)
			}
		}()
	}
}

func (w *workerManager) startStoryWorkers(ctx context.Context, wg *sync.WaitGroup, stats *jobProcessingStats, storyChan chan int, commentChan chan int) {
	const storyWorkers = 5
	for i := 0; i < storyWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range storyChan {
				w.scheduler.processStory(ctx, id, stats, commentChan)
			}
		}()
	}
}
