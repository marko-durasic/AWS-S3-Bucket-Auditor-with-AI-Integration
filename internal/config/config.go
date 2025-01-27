package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultMacieTimeout = 40 * time.Minute
)

// GetMacieTimeout returns the Macie job timeout duration from environment variable
// or falls back to default value (40 minutes)
func GetMacieTimeout() time.Duration {
	timeoutStr := os.Getenv("MACIE_JOB_TIMEOUT_MINUTES")
	if timeoutStr == "" {
		return defaultMacieTimeout
	}

	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return defaultMacieTimeout
	}

	return time.Duration(timeout) * time.Minute
}
