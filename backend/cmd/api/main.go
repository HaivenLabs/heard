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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	providerClient := &http.Client{Timeout: 10 * time.Second}
	if err := app.ConfigurePassageGoogle(ctx, cfg, providerClient); err != nil {
		// A Passage outage must not take down public guest-feedback routes or
		// already authenticated sessions. Retry the durable configuration while
		// this instance remains healthy.
		log.Printf("identity provider configuration is temporarily unavailable: %v", err)
		go retryIdentityProviderConfiguration(ctx, cfg, providerClient)
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
	app.InitializeIdentityCache(ctx, identity, store)

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

func retryIdentityProviderConfiguration(ctx context.Context, cfg app.Config, client *http.Client) {
	for {
		timer := time.NewTimer(30 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		if err := app.ConfigurePassageGoogle(ctx, cfg, client); err != nil {
			log.Printf("identity provider configuration retry failed: %v", err)
			continue
		}
		log.Printf("identity provider configuration restored")
		return
	}
}
