// Command worker is the entry point for lms-worker — asynchronous jobs:
// due-date notices, fines accrual, catalog reindexing
// (ADR-009-worker-scheduling-model.md). Polling/cron-triggered, not
// event-driven in the message-broker sense (AT-002) — see this repo's
// README for the correction to its own earlier "event-driven" wording.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/code-corhuila/lms-worker/internal/application/usecase"
	"github.com/code-corhuila/lms-worker/internal/config"
	"github.com/code-corhuila/lms-worker/internal/infrastructure/httpclient"
	"github.com/code-corhuila/lms-worker/internal/infrastructure/logger"
	"github.com/code-corhuila/lms-worker/internal/infrastructure/scheduler"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		println("config error:", err.Error())
		return err
	}

	zapLog, err := logger.New(cfg.LogLevel)
	if err != nil {
		return err
	}
	defer func() { _ = zapLog.Sync() }()
	log := zapLog.Sugar()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	circulation := httpclient.NewCirculationClient(cfg.CirculationServiceURL, cfg.InternalJWTSecret)
	catalog := httpclient.NewCatalogClient(cfg.CatalogServiceURL, cfg.InternalJWTSecret)

	jobs := map[string]scheduler.Job{
		"due-date-notices": usecase.NewDueDateNotices(circulation, log),
		"fines-accrual":    usecase.NewFinesAccrual(circulation),
		"catalog-reindex":  usecase.NewCatalogReindex(catalog),
	}

	ticker := scheduler.NewTicker(cfg.PollInterval, jobs, log)

	log.Infow("lms-worker started", "pollInterval", cfg.PollInterval)
	ticker.Run(ctx)

	log.Info("shutting down")
	return nil
}
