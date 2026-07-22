package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ccawmiku/disk-insight/internal/analytics"
	"github.com/ccawmiku/disk-insight/internal/api"
	"github.com/ccawmiku/disk-insight/internal/config"
	"github.com/ccawmiku/disk-insight/internal/scanner"
	"github.com/ccawmiku/disk-insight/internal/scheduler"
	"github.com/ccawmiku/disk-insight/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	settings, err := config.FromEnvironment()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	dataStore, err := store.Open(settings.DatabasePath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer dataStore.Close()
	if err := dataStore.SyncRoots(context.Background(), settings.Roots); err != nil {
		logger.Error("sync scan roots", "error", err)
		os.Exit(1)
	}
	dashboard := analytics.NewDashboardService(dataStore)
	scanManager := scanner.New(dataStore, dashboard.Invalidate)
	handler := api.New(dataStore, scanManager, dashboard, settings.WebPath, logger)
	server := &http.Server{Addr: settings.Address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go scheduler.New(dataStore, scanManager).Run(ctx)
	if err := scheduler.StartInitialScans(ctx, dataStore, scanManager); err != nil {
		logger.Error("start initial scans", "error", err)
		os.Exit(1)
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	logger.Info("disk-insight started", "address", settings.Address, "roots", len(settings.Roots))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
