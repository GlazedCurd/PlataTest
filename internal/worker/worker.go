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

func (w *Worker) worker(ctx context.Context, update chan *model.Update, wg *sync.WaitGroup) {
	defer wg.Done()
	for update := range update {
		w.log.Info("Processing update", zap.Any("update", update))
		quota, err := w.quotaFetcher.FetchQuota(ctx, update.Code, w.log.With(zap.Uint64("update_id", update.ID)))
		if err != nil {
			w.log.Error("Failed to fetch quota", zap.Error(err))
			update.Status = model.STATUS_FAILED
			_, err = w.db.UpdateUpdate(ctx, update)
			if err != nil {
				w.log.Error("Failed to update quote status", zap.Error(err))
			}
			continue
		}
		update.Price = &quota
		update.Status = model.STATUS_SUCCESS
		_, err = w.db.UpdateUpdate(ctx, update)
		if err != nil {
			w.log.Error("Failed to update quote status", zap.Error(err))
		}
		w.log.Info("Fetched quota", zap.Any("quota", quota))
	}
}

func (w *Worker) doWork() {
	ctx, cancel := context.WithTimeout(context.Background(), w.tick)
	defer cancel()
	updates, err := w.db.GetRecentlyUpdatesToProcess(ctx)
	if err != nil {
		w.log.Error("Failed to get recently updates to process", zap.Error(err))
		return
	}
	chanUpdates := make(chan *model.Update)

	var wg sync.WaitGroup
	for i := 0; i < w.numWorkers; i++ {
		wg.Add(1)
		go w.worker(ctx, chanUpdates, &wg)
	}

	for _, update := range updates {
		chanUpdates <- &update
	}
	close(chanUpdates)
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
