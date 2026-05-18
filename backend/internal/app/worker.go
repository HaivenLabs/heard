package app

import (
	"context"
	"log"
	"time"
)

type Worker struct {
	cfg   Config
	store *Store
}

func NewWorker(cfg Config, store *Store) *Worker {
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
			processed, err := w.store.ProcessNextOutboxEvent(ctx)
			if err != nil {
				log.Printf("worker error: %v", err)
				continue
			}
			if processed {
				log.Printf("worker processed outbox event")
			}
		}
	}
}
