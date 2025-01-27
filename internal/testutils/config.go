package testutils

import (
	"os"
	"testing"
)

const (
	TestBucketPrefix = "s3auditor-test-"
)

// RequireAWSEnv skips the test if AWS credentials are not configured
func RequireAWSEnv(t *testing.T) {
	t.Helper()

	required := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping test: %s environment variable not set", env)
		}
	}
}
