package app

import (
	"context"
	"log"
	"time"
)

type Worker struct {
	cfg   Config
	store outboxProcessor
}

type outboxProcessor interface {
	ProcessNextOutboxEvent(context.Context) (bool, error)
}

type OutboxBatchResult struct {
	Attempted int
	Succeeded int
	Failed    int
}

func NewWorker(cfg Config, store outboxProcessor) *Worker {
	return &Worker{cfg: cfg, store: store}
}

func (w *Worker) Run(ctx context.Context) error {
	interval := time.Duration(w.cfg.WorkerPollIntervalMS) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			result := w.processBatch(ctx)
			if result.Attempted > 0 {
				log.Printf(`{"event":"worker.outbox_batch","attempted":%d,"succeeded":%d,"failed":%d}`, result.Attempted, result.Succeeded, result.Failed)
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) OutboxBatchResult {
	result := OutboxBatchResult{}
	batchSize := w.cfg.WorkerBatchSize
	if batchSize < 1 {
		batchSize = 1
	}
	for result.Attempted < batchSize {
		processed, err := w.store.ProcessNextOutboxEvent(ctx)
		if !processed && err == nil {
			break
		}
		result.Attempted++
		if err != nil {
			result.Failed++
			log.Printf("worker outbox event failed: %v", err)
			if !processed {
				break
			}
			continue
		}
		result.Succeeded++
	}
	return result
}
