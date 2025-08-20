package worker

import (
	"context"
	"sync"
	"time"

	"github.com/GlazedCurd/PlataTest/internal/db"
	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/GlazedCurd/PlataTest/internal/quotafetcher"
	"go.uber.org/zap"
)

type Worker struct {
	db           db.DB
	log          *zap.Logger
	tick         time.Duration
	quotaFetcher quotafetcher.QuotaFetcher
	numWorkers   int
}

func NewWorker(db db.DB, tick time.Duration, numWorkers int, logger *zap.Logger, quotaFetcher quotafetcher.QuotaFetcher) *Worker {
	return &Worker{db: db, tick: tick, numWorkers: numWorkers, log: logger, quotaFetcher: quotaFetcher}
}

func (w *Worker) worker(ctx context.Context, task chan *model.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range task {
		w.log.Info("Processing task", zap.Any("task", task))
		quota, err := w.quotaFetcher.FetchQuota(ctx, task.Code, w.log.With(zap.Uint64("task_id", task.ID)))
		if err != nil {
			w.log.Error("Fetching quota", zap.Error(err))
			task.Status = model.STATUS_FAILED
			_, err = w.db.UpdateTask(ctx, task)
			if err != nil {
				w.log.Error("Task quote status", zap.Error(err))
			}
			continue
		}
		task.Price = &quota
		task.Status = model.STATUS_SUCCESS
		_, err = w.db.UpdateTask(ctx, task)
		if err != nil {
			w.log.Error("Task quote status", zap.Error(err))
		}
		w.log.Info("Fetched quota", zap.Any("quota", quota))
	}
}

func (w *Worker) doWork() {
	ctx, cancel := context.WithTimeout(context.Background(), w.tick)
	defer cancel()
	tasks, err := w.db.GetRecentlyTasksToProcess(ctx)
	if err != nil {
		w.log.Error("Get recently tasks to process", zap.Error(err))
		return
	}
	chanTasks := make(chan *model.Task)

	var wg sync.WaitGroup
	for i := 0; i < w.numWorkers; i++ {
		wg.Add(1)
		go w.worker(ctx, chanTasks, &wg)
	}

	for _, task := range tasks {
		chanTasks <- &task
	}
	close(chanTasks)
	wg.Wait()
}

func (w *Worker) Start() {
	w.log.Info("Worker started")
	defer w.log.Info("Worker stopped")

	ticker := time.Tick(w.tick)
	for range ticker {
		w.log.Info("Worker is working...")
		w.doWork()
	}
}
