package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/haivenlabs/heard/backend/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	if err := cfg.ValidateAPI(); err != nil {
		log.Fatalf("invalid runtime configuration: %v", err)
	}
	ctx := context.Background()
	if err := app.ConfigurePassageGoogle(ctx, cfg, &http.Client{Timeout: 10 * time.Second}); err != nil {
		log.Fatalf("configure Google sign-in: %v", err)
	}

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

	identity, err := app.NewIdentityProvider(cfg)
	if err != nil {
		log.Fatalf("configure Passage identity provider: %v", err)
	}

	server := app.NewServer(cfg, store, identity)
	httpServer := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    64 * 1024,
	}

	go func() {
		log.Printf("heard api listening on :%s", cfg.AppPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
