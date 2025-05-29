package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ENV        string `envconfig:"ENV" default:"DEV"`
	APIAURL    string `envconfig:"API_A_URL" default:"https://jsonplaceholder.typicode.com/users"`
	APIBURL    string `envconfig:"API_B_URL" default:"https://webhook.site" required:"true"`
	MaxRetries int    `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay int    `envconfig:"RETRY_DELAY_MS" default:"1000"`
	Timeout    int    `envconfig:"TIMEOUT" default:"10"`
}

func ParseConfig() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	err := envconfig.Process("", &cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}
	return &cfg, nil
}
