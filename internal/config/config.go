// Package config loads configuration from environment variables — the only source of
// environment-specific configuration (library-docs/05-architecture/cross-cutting.md).
package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	// InternalJWTSecret mints the short-lived service-to-service token this
	// worker uses to call lms-circulation-api and lms-catalog-api —
	// deliberately separate from any end-user RS256 session token
	// (rules/2-anexos/C-api-hexagonal.md, numeral 5.3.7).
	InternalJWTSecret string

	CirculationServiceURL string
	CatalogServiceURL     string

	// PollInterval — ADR-009-worker-scheduling-model.md: 15 minutes,
	// polling/cron, not event-driven (AT-002).
	PollInterval time.Duration

	LogLevel string
}

func Load() (*Config, error) {
	pollInterval, err := time.ParseDuration(getEnv("POLL_INTERVAL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
	}

	cfg := &Config{
		InternalJWTSecret: getEnv("INTERNAL_JWT_SECRET", ""),

		CirculationServiceURL: getEnv("CIRCULATION_SERVICE_URL", "http://circulation-service:8080"),
		CatalogServiceURL:     getEnv("CATALOG_SERVICE_URL", "http://catalog-service:8080"),

		PollInterval: pollInterval,

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	if cfg.InternalJWTSecret == "" {
		return nil, fmt.Errorf("INTERNAL_JWT_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
