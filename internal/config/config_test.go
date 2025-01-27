package config

import (
	"os"
	"testing"
	"time"
)

func TestGetMacieTimeout(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedValue time.Duration
		setup         func()
		cleanup       func()
	}{
		{
			name:          "Default value when env not set",
			envValue:      "",
			expectedValue: defaultMacieTimeout,
			setup:         func() { os.Unsetenv("MACIE_JOB_TIMEOUT_MINUTES") },
			cleanup:       func() {},
		},
		{
			name:          "Custom value from env",
			envValue:      "60",
			expectedValue: 60 * time.Minute,
			setup:         func() { os.Setenv("MACIE_JOB_TIMEOUT_MINUTES", "60") },
			cleanup:       func() { os.Unsetenv("MACIE_JOB_TIMEOUT_MINUTES") },
		},
		{
			name:          "Invalid env value falls back to default",
			envValue:      "invalid",
			expectedValue: defaultMacieTimeout,
			setup:         func() { os.Setenv("MACIE_JOB_TIMEOUT_MINUTES", "invalid") },
			cleanup:       func() { os.Unsetenv("MACIE_JOB_TIMEOUT_MINUTES") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			got := GetMacieTimeout()
			if got != tt.expectedValue {
				t.Errorf("GetMacieTimeout() = %v, want %v", got, tt.expectedValue)
			}
		})
	}
}
