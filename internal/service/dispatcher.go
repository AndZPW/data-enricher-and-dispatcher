package service

import (
	"context"
	"fmt"
	"github.com/AndZPW/data-enricher-and-dispatcher/internal/client"
	"strings"

	"go.uber.org/zap"
)

type Dispatcher struct {
	apiClient *client.APIClient
	logger    *zap.Logger
}

func NewDispatcher(apiClient *client.APIClient, logger *zap.Logger) *Dispatcher {
	return &Dispatcher{
		apiClient: apiClient,
		logger:    logger,
	}
}

func (d *Dispatcher) ProcessUsers(ctx context.Context) error {
	users, err := d.apiClient.GetUsers(ctx)
	if err != nil {
		d.logger.Error("Failed to get users", zap.Error(err))
		return fmt.Errorf("failed to get users: %w", err)
	}

	d.logger.Info("Starting users processing",
		zap.Int("total_users", len(users)))

	bizCount := 0
	nonBizCount := 0

	for _, user := range users {
		select {
		case <-ctx.Done():
			d.logger.Warn("Processing interrupted by context")
			return ctx.Err()
		default:
			if strings.HasSuffix(user.Email, ".biz") {
				bizCount++
				if err := d.apiClient.SendUsers(ctx, user); err != nil {
					d.logger.Error("Failed to send user to API B",
						zap.String("user_name", user.Name),
						zap.String("user_email", user.Email),
						zap.Error(err))
				}
			} else {
				nonBizCount++
				d.logger.Info("Skipping user - non .biz domain",
					zap.String("user_name", user.Name),
					zap.String("user_email", user.Email))
			}
		}
	}

	d.logger.Info("Processing completed",
		zap.Int("biz_users_processed", bizCount),
		zap.Int("non_biz_users_skipped", nonBizCount))

	return nil
}
