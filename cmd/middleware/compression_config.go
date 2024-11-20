package middleware

import (
	"fmt"

	"github.com/lokeam/bravo-kilo/internal/shared/utils"
)

func LoadCompressionConfig() (*AdaptiveCompressionConfig, error) {
	// Start with default
	cfg := DefaultCompressionConfig

	// Load values with validation using shared utils
	cfg.MinSize = utils.GetEnvIntWithFallback("COMPRESSION_MIN_SIZE", cfg.MinSize)
	cfg.Level = utils.GetEnvIntWithFallback("COMPRESSION_LEVEL", cfg.Level)
	cfg.MaxRetries = utils.GetEnvIntWithFallback("COMPRESSION_MAX_RETRIES", cfg.MaxRetries)
	cfg.RetryDelay = utils.GetEnvDurationWithFallback("COMPRESSION_RETRY_DELAY", cfg.RetryDelay)
	cfg.UpdateInterval = utils.GetEnvDurationWithFallback("COMPRESSION_UPDATE_INTERVAL", cfg.UpdateInterval)
	cfg.TargetLatency = utils.GetEnvDurationWithFallback("COMPRESSION_TARGET_LATENCY", cfg.TargetLatency)

	if err := cfg.Validate(); err != nil {
		return &cfg, fmt.Errorf("invalid compression config: %w", err)
}

	return &cfg, nil
}