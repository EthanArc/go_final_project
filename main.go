package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EthanArc/go_final_project/pkg/db"
	"github.com/EthanArc/go_final_project/pkg/server"
)

func main() {
	// 1. Use structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Initialize database
	if err := db.Init("scheduler.db"); err != nil {
		logger.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()

	// 3. Initialize server
	srv := server.StartServer(logger)

	// 4. Setup graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// 5. Start server in a background goroutine
	go func() {
		logger.Info("Starting server on port :7540")
		if err := srv.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server startup error", "error", err)
			os.Exit(1)
		}
	}()

	// 6. Wait for termination signal
	sig := <-shutdownChan
	logger.Info("Shutdown signal received", "signal", sig.String())

	// 7. Execute graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	} else {
		logger.Info("Server stopped cleanly")
	}
}
