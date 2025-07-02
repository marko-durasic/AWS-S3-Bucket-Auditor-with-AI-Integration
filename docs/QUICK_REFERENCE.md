# AWS S3 Bucket Auditor - Quick Reference

## Essential Commands

### Building and Running

```bash
# Build the application
go build -o s3auditor cmd/s3auditor/main.go

# Run directly
go run cmd/s3auditor/main.go

# Run tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Environment Setup

```bash
# Required AWS credentials
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=us-east-1

# Optional configuration
export MACIE_JOB_TIMEOUT_MINUTES=60
export AWS_PROFILE=your_profile
```

## Core API Functions

### AWS Client Initialization

```go
import "github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"

// Create AWS clients
ctx := context.Background()
clients, err := awsutils.NewAWSClients(ctx)
if err != nil {
    log.Fatal(err)
}
```

### S3 Operations

```go
// List all buckets
buckets, err := awsutils.ListBuckets(clients.S3Client)

// Get bucket region
region, err := awsutils.GetBucketRegion(clients.S3Client, "bucket-name")

// Check if bucket is public
isPublic, err := awsutils.IsBucketPublic(clients.S3Client, "bucket-name")

// Get encryption status
encryption, err := awsutils.GetBucketEncryption(clients.S3Client, "bucket-name")

// Check versioning
versioning, err := awsutils.GetBucketVersioning(clients.S3Client, "bucket-name")
```

### Audit Operations

```go
import (
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/audit"
    "github.com/aws/aws-sdk-go-v2/service/sts"
)

// Create scanner
stsClient := sts.NewFromConfig(clients.Config)
scanner := audit.NewScanner(
    clients.Config,
    clients.S3Client,
    clients.MacieClient,
    stsClient,
)

// Audit a bucket
err := scanner.AuditBucket("bucket-name")
```

### UI Functions

```go
import "github.com/marko-durasic/aws-s3-bucket-auditor/internal/ui"

// Display messages
ui.ShowWelcomeScreen()
ui.ShowError("Error: %v", err)
ui.ShowSuccess("Operation completed successfully")
```

### CLI Functions

```go
import "github.com/marko-durasic/aws-s3-bucket-auditor/internal/cli"

// Show main menu
choice, err := cli.PromptMainMenu()

// Select bucket
bucketName, err := cli.PromptForBucketSelection(s3Client)

// Display buckets list
cli.DisplayBucketsList(s3Client, buckets)

// Handle audit workflow
cli.HandleBucketAudit(cfg, s3Client, macieClient)
```

## Data Models

### BucketBasicInfo

```go
type BucketBasicInfo struct {
    Name   string  // S3 bucket name
    Region string  // AWS region
}
```

### BucketInfo

```go
type BucketInfo struct {
    Name             string        // S3 bucket name
    Region           string        // AWS region
    IsPublic         bool          // Public access status
    Encryption       string        // Encryption algorithm
    VersioningStatus string        // Versioning status
    SensitiveData    bool          // Macie findings
    AuditDuration    time.Duration // Audit execution time
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MACIE_JOB_TIMEOUT_MINUTES` | 40 | Macie job timeout |
| `AWS_REGION` | us-east-1 | Default AWS region |
| `AWS_PROFILE` | default | AWS profile |

### Get Configuration

```go
import "github.com/marko-durasic/aws-s3-bucket-auditor/internal/config"

timeout := config.GetMacieTimeout()
```

## Testing

### Mock Interfaces

```go
type MockS3Client struct {
    ListBucketsFunc func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
}

func (m *MockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
    if m.ListBucketsFunc != nil {
        return m.ListBucketsFunc(ctx, params, optFns...)
    }
    return nil, errors.New("not implemented")
}
```

### Test Utilities

```go
import "github.com/marko-durasic/aws-s3-bucket-auditor/internal/testutils"

// Generate test bucket name
bucketName := testutils.GetTestBucketName("test-prefix")

// Create test bucket with cleanup
cleanup := testutils.CreateTestBucket(t, s3Client, bucketName)
defer cleanup()

// Check AWS environment
testutils.RequireAWSEnv(t)
```

## Common Patterns

### Error Handling

```go
// Standard error handling
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// AWS-specific error handling
var awsErr smithy.APIError
if errors.As(err, &awsErr) {
    switch awsErr.ErrorCode() {
    case "NoSuchBucket":
        return fmt.Errorf("bucket not found: %w", err)
    case "AccessDenied":
        return fmt.Errorf("insufficient permissions: %w", err)
    }
}
```

### Concurrent Operations

```go
// Using goroutines with WaitGroup
wg := sync.WaitGroup{}
wg.Add(1)

go func() {
    defer wg.Done()
    // Perform operation
}()

wg.Wait()
```

### Resource Cleanup

```go
// Defer cleanup
defer func() {
    if resource != nil {
        resource.Close()
    }
}()
```

## Security Best Practices

### Input Validation

```go
func validateBucketName(name string) error {
    if name == "" {
        return errors.New("bucket name cannot be empty")
    }
    if len(name) < 3 || len(name) > 63 {
        return errors.New("invalid bucket name length")
    }
    return nil
}
```

### Credential Security

```go
// Use AWS SDK credential chain (recommended)
cfg, err := config.LoadDefaultConfig(ctx)

// Avoid hardcoded credentials
// ❌ Don't do this:
// cfg := aws.Config{
//     Credentials: credentials.NewStaticCredentialsProvider("key", "secret", ""),
// }
```

## Performance Tips

### Pre-allocate Slices

```go
// Good: pre-allocate with known capacity
buckets := make([]models.BucketBasicInfo, 0, len(bucketNames))

// Avoid: growing slice dynamically
var buckets []models.BucketBasicInfo
```

### Parallel Operations

```go
// Use channels for parallel operations
results := make(chan result, numOperations)

for i := 0; i < numOperations; i++ {
    go func(i int) {
        result := performOperation(i)
        results <- result
    }(i)
}

// Collect results
for i := 0; i < numOperations; i++ {
    r := <-results
    // Process result
}
```

## Debugging

### Enable Verbose Logging

```bash
# Go debugging
export GODEBUG=netdns=go

# Run with output capture
./s3auditor 2>&1 | tee debug.log
```

### Check Log Files

```bash
# Monitor audit log
tail -f s3_audit.log

# Search for errors
grep -i error s3_audit.log
```

## Integration Examples

### Custom Audit Script

```go
package main

import (
    "context"
    "log"
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/audit"
    "github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
    ctx := context.Background()
    clients, err := awsutils.NewAWSClients(ctx)
    if err != nil {
        log.Fatal(err)
    }

    stsClient := sts.NewFromConfig(clients.Config)
    scanner := audit.NewScanner(
        clients.Config,
        clients.S3Client,
        clients.MacieClient,
        stsClient,
    )

    buckets := []string{"bucket1", "bucket2", "bucket3"}
    for _, bucket := range buckets {
        log.Printf("Auditing: %s", bucket)
        if err := scanner.AuditBucket(bucket); err != nil {
            log.Printf("Failed to audit %s: %v", bucket, err)
        }
    }
}
```

### Batch Security Check

```go
func checkSecurityCompliance(buckets []string) map[string]bool {
    clients, _ := awsutils.NewAWSClients(context.Background())
    results := make(map[string]bool)

    for _, bucket := range buckets {
        isPublic, _ := awsutils.IsBucketPublic(clients.S3Client, bucket)
        encryption, _ := awsutils.GetBucketEncryption(clients.S3Client, bucket)
        versioning, _ := awsutils.GetBucketVersioning(clients.S3Client, bucket)

        // Compliance: not public, encrypted, versioned
        compliant := !isPublic && 
                    encryption != "Not Enabled" && 
                    versioning == "Enabled"
        
        results[bucket] = compliant
    }

    return results
}
```

## Troubleshooting Quick Fixes

### Permission Issues

```bash
# Check AWS credentials
aws sts get-caller-identity

# Verify region access
aws s3 ls --region us-east-1
```

### Macie Issues

```bash
# Check Macie status
aws macie2 get-macie-session --region us-east-1

# Increase timeout
export MACIE_JOB_TIMEOUT_MINUTES=120
```

### Network Issues

```bash
# Test connectivity
ping s3.amazonaws.com

# Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY
```

## IAM Policy Template

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListAllMyBuckets",
                "s3:GetBucketLocation",
                "s3:GetBucketEncryption",
                "s3:GetBucketVersioning",
                "s3:GetPublicAccessBlock",
                "s3:GetBucketAcl"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "macie2:CreateClassificationJob",
                "macie2:DescribeClassificationJob",
                "macie2:ListFindings",
                "macie2:GetFindings"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "sts:GetCallerIdentity",
            "Resource": "*"
        }
    ]
}
```

This quick reference provides the essential information needed for daily development and usage of the AWS S3 Bucket Auditor.