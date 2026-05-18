package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/haivenlabs/heard/backend/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	ctx := context.Background()

	store, err := app.NewStore(ctx, cfg)
	if err != nil {
		log.Fatalf("create store: %v", err)
	}
	defer store.Close()

	if err := store.RunMigrations(ctx); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	if err := store.SeedDemoData(ctx); err != nil {
		log.Fatalf("seed demo data: %v", err)
	}

	worker := app.NewWorker(cfg, store)
	workerCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := worker.Run(workerCtx); err != nil {
		log.Fatalf("run worker: %v", err)
	}
}
