package main

import (
	"context"
	"github.com/AndZPW/data-enricher-and-dispatcher/internal/client"
	"github.com/AndZPW/data-enricher-and-dispatcher/internal/config"
	log "github.com/AndZPW/data-enricher-and-dispatcher/internal/logger"
	"github.com/AndZPW/data-enricher-and-dispatcher/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {

	cfg, err := config.ParseConfig()

	logger := log.MustInitLogger(cfg.ENV)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	if err != nil {
		logger.Error(err.Error())
	}

	apiClient := client.NewAPIClient(cfg, logger)
	if err != nil {
		logger.Error("Failed to create API client", zap.Error(err))
	}

	dispatcher := service.NewDispatcher(apiClient, logger)

	logger.Info("Starting users processing")
	if err = dispatcher.ProcessUsers(ctx); err != nil {
		logger.Error("Processing finished with error", zap.Error(err))
	} else {
		logger.Info("Processing finished successfully")
	}

	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
	}
}
