package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AndZPW/data-enricher-and-dispatcher/internal/config"
	"github.com/AndZPW/data-enricher-and-dispatcher/internal/model"
	"io"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type APIClient struct {
	config *config.Config
	client *http.Client
	logger *zap.Logger
}

func NewAPIClient(cfg *config.Config, logger *zap.Logger) *APIClient {

	return &APIClient{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}
}

func (c *APIClient) GetUsers(parentCtx context.Context) ([]model.User, error) {

	c.logger.Info("Fetching users from source API", zap.String("url", c.config.APIAURL))

	ctx, cancel := context.WithTimeout(parentCtx, time.Duration(c.config.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.APIAURL, nil)
	if err != nil {
		c.logger.Error("Failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to get users", zap.Error(err))
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			c.logger.Error("Failed to close response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Unexpected status code",
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var users []model.User
	if err = json.Unmarshal(body, &users); err != nil {
		c.logger.Error("Failed to unmarshal users", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	c.logger.Info("Successfully fetched users", zap.Int("count", len(users)))

	return users, nil
}

func (c *APIClient) SendUsers(ctx context.Context, user model.User) error {
	data := model.FromUserToUserForB(user)

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Error("Failed to marshal user data",
			zap.Error(err),
			zap.String("user_name", user.Name))
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	baseDelay := time.Duration(c.config.RetryDelay) * time.Millisecond
	var allErrors []error

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			c.logger.Warn("Context cancelled, aborting retries",
				zap.String("user_name", user.Name),
				zap.Int("attempt", attempt+1),
				zap.Error(ctx.Err()))
			allErrors = append(allErrors, fmt.Errorf("attempt %d: context cancelled: %w", attempt+1, ctx.Err()))
			return fmt.Errorf("retry attempts failed: %w", errors.Join(allErrors...))
		}

		if attempt > 0 {
			delay := calculateBackoffWithJitter(baseDelay, attempt)
			c.logger.Info("Retrying request",
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
				zap.String("user_name", user.Name))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				c.logger.Warn("Context cancelled during backoff delay",
					zap.String("user_name", user.Name),
					zap.Int("attempt", attempt+1))
				allErrors = append(allErrors, fmt.Errorf("attempt %d: context cancelled during backoff: %w", attempt+1, ctx.Err()))
				return fmt.Errorf("retry attempts failed: %w", errors.Join(allErrors...))
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIBURL, bytes.NewBuffer(jsonData))
		if err != nil {
			err = fmt.Errorf("attempt %d: failed to create request: %w", attempt+1, err)
			allErrors = append(allErrors, err)
			c.logger.Warn("Failed to create request",
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			err = fmt.Errorf("attempt %d: failed to send to API B: %w", attempt+1, err)
			allErrors = append(allErrors, err)
			c.logger.Warn("Failed to send request to API B",
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			continue
		}

		func() {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				c.logger.Info("Successfully sent user to API B",
					zap.String("user_name", user.Name),
					zap.Int("attempt", attempt+1))
				allErrors = nil
				return
			}

			err = fmt.Errorf("attempt %d: unexpected status code from API B: %d", attempt+1, resp.StatusCode)
			allErrors = append(allErrors, err)
			c.logger.Warn("Received non-2xx status from API B",
				zap.Int("status_code", resp.StatusCode),
				zap.Int("attempt", attempt+1))
		}()

		if len(allErrors) == 0 {
			return nil
		}
	}

	if len(allErrors) > 0 {
		combinedError := fmt.Errorf("all retry attempts failed: %w", errors.Join(allErrors...))
		c.logger.Error("Failed to send user to API B after retries",
			zap.String("user_name", user.Name),
			zap.Int("max_retries", c.config.MaxRetries),
			zap.Error(combinedError))
		return combinedError
	}

	return nil
}

func calculateBackoffWithJitter(baseDelay time.Duration, attempt int) time.Duration {
	maxDelay := baseDelay * time.Duration(1<<uint(attempt))
	jitter := time.Duration(rand.Int63n(int64(maxDelay / 2)))
	return maxDelay/2 + jitter
}
