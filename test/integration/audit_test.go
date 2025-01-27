package integration

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/audit"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/config"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Setup test environment
	if err := setup(); err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	teardown()
	os.Exit(code)
}

func setup() error {
	// Setup test buckets and configurations
	return nil
}

func teardown() {
	// Cleanup test resources
}

func TestAuditFlow(t *testing.T) {
	// Increase test timeout for Macie jobs
	t.Parallel() // Allow tests to run in parallel
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("AWS_ACCESS_KEY_ID not set")
	}

	testutils.RequireAWSEnv(t)

	// Use config package to get values
	bucketPrefix := config.GetTestBucketPrefix()

	clients, err := awsutils.NewAWSClients(context.Background())
	assert.NoError(t, err)

	tests := []struct {
		name      string
		setupData bool
		wantErr   bool
	}{
		{
			name:      "empty-bucket",
			setupData: false,
			wantErr:   false,
		},
		{
			name:      "sensitive-data",
			setupData: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucketName := testutils.GetTestBucketName(bucketPrefix + tt.name)
			cleanup := testutils.CreateTestBucket(t, clients.S3Client, bucketName)
			defer cleanup()

			if tt.setupData {
				err := testutils.SetupTestData(t, clients.S3Client, bucketName)
				assert.NoError(t, err)
			}

			s3Client := s3.NewFromConfig(clients.Config)
			macieClient := macie2.NewFromConfig(clients.Config)
			stsClient := sts.NewFromConfig(clients.Config)

			scanner := audit.NewScanner(clients.Config, s3Client, macieClient, stsClient)
			err := scanner.AuditBucket(bucketName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
