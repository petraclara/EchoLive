package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pulseroom/api/internal/config"
	httpserver "github.com/pulseroom/api/internal/http"
	"github.com/pulseroom/api/internal/repository"
	"github.com/pulseroom/api/internal/ws"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := repository.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := repository.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	store := repository.NewStore(pool)
	hub := ws.NewHub()
	srv := httpserver.NewServer(cfg, store, hub)

	addr := ":" + cfg.Port
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("PulseRoom API listening on %s\n", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
