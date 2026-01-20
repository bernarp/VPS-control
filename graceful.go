package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func (app *application) Run() error {
	router := setupRouter(app.logger)
	app.registerRoutes(router)

	srv := app.newServer(router)

	serverErrors := make(chan error, 1)

	go func() {
		zap.String("port", app.cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-quit:
		zap.String("signal", sig.String())
		shutdownTimeout := 25 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			app.logger.Error("Graceful shutdown failed", zap.Error(err))
			_ = srv.Close()
		} else {
			app.logger.Info("HTTP server stopped successfully")
		}
	}

	app.closeResources()

	if err := app.logger.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Logger sync error: %v\n", err)
	}

	app.logger.Info("Application terminated gracefully")
	return nil
}
