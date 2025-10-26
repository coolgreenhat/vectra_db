package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"vectraDB/internal/api"
	"vectraDB/internal/config"
	"vectraDB/internal/logger"
	"vectraDB/internal/middleware"
	"vectraDB/internal/store"
)

var version = "v0.1.0"

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	logger.Info("Starting VectraDB", "version", version)

	// Initialize store
	storeConfig := store.Config{
		DBPath:    cfg.Database.Path,
		Timeout:   cfg.Database.Timeout,
		MaxConns:  100,
		BatchSize: 1000,
	}

	store, err := store.NewBoltStore(storeConfig)
	if err != nil {
		logger.Fatal("Failed to initialize store", "error", err)
	}
	defer store.Close()

	// Initialize handler
	handler := api.NewHandler(store)

	// Setup router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RealIPMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.CompressMiddleware())

	// Mount routes
	r.Mount("/api/v1", handler.Routes())

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
