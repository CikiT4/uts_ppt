package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"legal-consultation-api/internal/config"
	"legal-consultation-api/internal/database"
	"legal-consultation-api/internal/router"
)

func main() {
	// Load config
	config.Load()
	cfg := config.AppConfig

	// Connect database
	database.Connect()
	defer database.Close()

	// Ensure upload directory exists
	os.MkdirAll(cfg.UploadDir, 0755)

	// Setup routes
	r := router.Setup(database.DB)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 %s running on port %s (env: %s)", cfg.AppName, cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
