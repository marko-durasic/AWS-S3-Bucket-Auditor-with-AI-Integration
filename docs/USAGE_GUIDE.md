# AWS S3 Bucket Auditor - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Basic Usage](#basic-usage)
6. [Advanced Usage](#advanced-usage)
7. [Common Use Cases](#common-use-cases)
8. [Troubleshooting](#troubleshooting)
9. [Best Practices](#best-practices)

## Getting Started

The AWS S3 Bucket Auditor is a command-line tool that helps you assess the security posture of your S3 buckets by analyzing their configuration and detecting sensitive data using Amazon Macie.

### Key Features

- **Interactive CLI**: User-friendly menu-driven interface
- **Comprehensive Auditing**: Checks encryption, versioning, public access, and sensitive data
- **Real-time Analysis**: Uses AWS Macie for sensitive data detection
- **Detailed Reporting**: Color-coded reports with security recommendations
- **Batch Operations**: Support for auditing multiple buckets

## Prerequisites

### AWS Requirements

1. **AWS Account**: Active AWS account with appropriate permissions
2. **AWS CLI**: Configured with credentials (optional but recommended)
3. **IAM Permissions**: Required permissions for S3, Macie, and STS services

### Required IAM Permissions

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
            "Action": [
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}
```

### System Requirements

- **Go**: Version 1.22.6 or later
- **Operating System**: Linux, macOS, or Windows
- **Network**: Internet connectivity for AWS API calls
- **Memory**: Minimum 512MB RAM
- **Disk Space**: 100MB for application and logs

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/marko-durasic/aws-s3-bucket-auditor.git
cd aws-s3-bucket-auditor

# Install dependencies
go mod download

# Build the application
go build -o s3auditor cmd/s3auditor/main.go

# Run the application
./s3auditor
```

### Option 2: Direct Run

```bash
# Clone and run directly
git clone https://github.com/marko-durasic/aws-s3-bucket-auditor.git
cd aws-s3-bucket-auditor
go run cmd/s3auditor/main.go
```

## Configuration

### AWS Credentials

The application uses the AWS SDK's default credential chain:

1. **Environment Variables**:
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

2. **AWS Credentials File** (`~/.aws/credentials`):
   ```ini
   [default]
   aws_access_key_id = your_access_key
   aws_secret_access_key = your_secret_key
   region = us-east-1
   ```

3. **AWS Profile**:
   ```bash
   export AWS_PROFILE=your_profile_name
   ```

4. **IAM Roles** (for EC2 instances)

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MACIE_JOB_TIMEOUT_MINUTES` | Timeout for Macie classification jobs | 40 | `60` |
| `AWS_REGION` | Default AWS region | us-east-1 | `us-west-2` |
| `AWS_PROFILE` | AWS profile to use | default | `production` |

### Configuration Example

```bash
# Set environment variables
export AWS_REGION=us-west-2
export AWS_PROFILE=security-audit
export MACIE_JOB_TIMEOUT_MINUTES=60

# Run the auditor
./s3auditor
```

## Basic Usage

### Starting the Application

```bash
./s3auditor
```

You'll see the welcome screen:
```
   _____ ____     ___             ___ __
  / ___/|_  /    /   | __  ______/ (_) /_____  _____
  \__ \ / /     / /| |/ / / / __  / / __/ __ \/ ___/
 ___/ // /_    / ___ / /_/ / /_/ / / /_/ /_/ / /
/____/____/   /_/  |_\__,_/\__,_/_/\__/\____/_/

Welcome to the AWS S3 Bucket Auditor!
```

### Main Menu Options

1. **List S3 Buckets**: Browse and view bucket details
2. **Audit a Bucket**: Perform comprehensive security audit
3. **Exit**: Close the application

### Listing Buckets

Select "List S3 Buckets" to see all buckets in your account:

```
S3 Buckets List (Enter for details, Ctrl+C or Exit option to return)
▸ my-app-logs (us-east-1)
  company-backups (us-west-2)
  website-assets (eu-west-1)
  [ Exit ]
```

**Features**:
- Search by typing bucket name or region
- Arrow keys for navigation
- Enter to view detailed information
- Ctrl+C or select "Exit" to return

### Viewing Bucket Details

Select any bucket to view its security configuration:

```
Bucket Details:
=====================================================================
Name              : my-app-logs
Region            : us-east-1
Encryption        : AES256
Versioning        : Enabled
Public Access     : No
---------------------------------------------------------------------
```

### Auditing a Bucket

Select "Audit a Bucket" for comprehensive security analysis:

1. **Bucket Selection**: Choose from available buckets
2. **Security Analysis**: Automated checks for:
   - Region identification
   - Public access configuration
   - Encryption status
   - Versioning settings
3. **Sensitive Data Scan**: Macie classification job
4. **Report Generation**: Detailed security report

### Sample Audit Report

```
🔍 Macie classification job created with Job ID: s3-audit-my-bucket-1234567890

Performing Macie Classification... [████████████████████████████████] 100%

S3 Bucket Security Audit Report:
=====================================================================
Bucket Name      : my-sensitive-bucket
Region           : us-east-1
Public Access    : false
Encryption       : aws:kms
Versioning       : Enabled
Sensitive Data   : true
Audit Duration   : 2m15s
---------------------------------------------------------------------

🛑 Finding ID: 12345-abcde-67890-fghij
Details: [Sensitive data detected in bucket]
```

## Advanced Usage

### Programmatic Usage

You can integrate the auditor components into your own applications:

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
    // Initialize clients
    ctx := context.Background()
    clients, err := awsutils.NewAWSClients(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create scanner
    stsClient := sts.NewFromConfig(clients.Config)
    scanner := audit.NewScanner(
        clients.Config,
        clients.S3Client,
        clients.MacieClient,
        stsClient,
    )
    
    // Audit specific bucket
    err = scanner.AuditBucket("my-bucket")
    if err != nil {
        log.Printf("Audit failed: %v", err)
    }
}
```

### Batch Auditing

Audit multiple buckets programmatically:

```go
func auditAllBuckets() {
    clients, _ := awsutils.NewAWSClients(context.Background())
    buckets, _ := awsutils.ListBuckets(clients.S3Client)
    
    stsClient := sts.NewFromConfig(clients.Config)
    scanner := audit.NewScanner(
        clients.Config,
        clients.S3Client,
        clients.MacieClient,
        stsClient,
    )
    
    for _, bucket := range buckets {
        log.Printf("Auditing bucket: %s", bucket.Name)
        err := scanner.AuditBucket(bucket.Name)
        if err != nil {
            log.Printf("Failed to audit %s: %v", bucket.Name, err)
        }
    }
}
```

### Custom Reporting

Create custom reports using the audit data:

```go
func generateCustomReport(bucketName string) {
    // Perform individual checks
    isPublic, _ := awsutils.IsBucketPublic(s3Client, bucketName)
    encryption, _ := awsutils.GetBucketEncryption(s3Client, bucketName)
    versioning, _ := awsutils.GetBucketVersioning(s3Client, bucketName)
    
    // Generate custom report
    report := map[string]interface{}{
        "bucket":     bucketName,
        "public":     isPublic,
        "encrypted":  encryption != "Not Enabled",
        "versioned":  versioning == "Enabled",
        "score":      calculateSecurityScore(isPublic, encryption, versioning),
    }
    
    // Output as JSON
    jsonReport, _ := json.MarshalIndent(report, "", "  ")
    fmt.Println(string(jsonReport))
}
```

## Common Use Cases

### 1. Security Compliance Audit

**Scenario**: Quarterly security review of all S3 buckets

**Steps**:
1. Run the auditor
2. Select "List S3 Buckets" to get overview
3. For each critical bucket, select "Audit a Bucket"
4. Document findings and remediation actions

**Automation**:
```bash
#!/bin/bash
# Automated security audit script
export MACIE_JOB_TIMEOUT_MINUTES=30
./s3auditor 2>&1 | tee audit-$(date +%Y%m%d).log
```

### 2. Data Classification Review

**Scenario**: Identify buckets containing sensitive data

**Focus Areas**:
- Enable Macie sensitive data detection
- Review buckets with positive findings
- Implement additional security controls

### 3. Public Access Assessment

**Scenario**: Ensure no buckets are inadvertently public

**Quick Check**:
```go
func checkPublicBuckets() {
    clients, _ := awsutils.NewAWSClients(context.Background())
    buckets, _ := awsutils.ListBuckets(clients.S3Client)
    
    for _, bucket := range buckets {
        isPublic, err := awsutils.IsBucketPublic(clients.S3Client, bucket.Name)
        if err != nil {
            log.Printf("Error checking %s: %v", bucket.Name, err)
            continue
        }
        
        if isPublic {
            log.Printf("WARNING: Bucket %s is public!", bucket.Name)
        }
    }
}
```

### 4. Encryption Compliance

**Scenario**: Verify all buckets use encryption

**Implementation**:
```go
func checkEncryptionCompliance() {
    clients, _ := awsutils.NewAWSClients(context.Background())
    buckets, _ := awsutils.ListBuckets(clients.S3Client)
    
    var unencrypted []string
    
    for _, bucket := range buckets {
        encryption, err := awsutils.GetBucketEncryption(clients.S3Client, bucket.Name)
        if err != nil || encryption == "Not Enabled" {
            unencrypted = append(unencrypted, bucket.Name)
        }
    }
    
    if len(unencrypted) > 0 {
        log.Printf("Unencrypted buckets: %v", unencrypted)
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Permission Denied

**Error**: `AccessDenied: User is not authorized to perform operation`

**Solution**:
- Verify IAM permissions
- Check AWS credentials configuration
- Ensure Macie is enabled in the region

#### 2. Macie Job Timeout

**Error**: `timeout waiting for Macie classification job completion`

**Solutions**:
- Increase timeout: `export MACIE_JOB_TIMEOUT_MINUTES=60`
- Check Macie service limits
- Verify bucket size and object count

#### 3. Region Mismatch

**Error**: `bucket is in a different region`

**Solution**:
- The auditor automatically handles cross-region buckets
- Ensure credentials have access to all regions

#### 4. Network Connectivity

**Error**: `connection timeout` or `network unreachable`

**Solutions**:
- Check internet connectivity
- Verify AWS endpoint accessibility
- Check proxy/firewall settings

### Debugging

Enable verbose logging:

```bash
# Set Go debug flags
export GODEBUG=netdns=go

# Run with verbose output
./s3auditor 2>&1 | tee debug.log
```

Check log file for detailed information:

```bash
tail -f s3_audit.log
```

### Performance Issues

#### Large Number of Buckets

For accounts with many buckets:

1. **Pagination**: The CLI automatically handles large bucket lists
2. **Filtering**: Use search functionality to find specific buckets
3. **Batch Processing**: Consider programmatic approach for bulk operations

#### Macie Job Performance

Factors affecting Macie performance:

- **Bucket Size**: Larger buckets take longer to scan
- **Object Count**: More objects increase processing time
- **Data Types**: Complex file types require more analysis time

## Best Practices

### Security

1. **Principle of Least Privilege**: Use minimal required IAM permissions
2. **Credential Management**: Use IAM roles instead of long-term keys when possible
3. **Audit Logs**: Regularly review audit logs for security events
4. **Sensitive Data**: Handle Macie findings securely

### Operational

1. **Regular Audits**: Schedule periodic security reviews
2. **Automation**: Integrate auditing into CI/CD pipelines
3. **Documentation**: Maintain records of audit findings and remediation
4. **Monitoring**: Set up alerts for security configuration changes

### Performance

1. **Timeout Configuration**: Adjust Macie timeout based on bucket sizes
2. **Regional Optimization**: Run audits from the same region as buckets when possible
3. **Batch Operations**: Group similar operations for efficiency
4. **Resource Limits**: Monitor AWS service limits and quotas

### Example Automation Script

```bash
#!/bin/bash
# S3 Security Audit Automation Script

set -e

# Configuration
AUDIT_DATE=$(date +%Y%m%d)
REPORT_DIR="./audit-reports"
LOG_FILE="$REPORT_DIR/audit-$AUDIT_DATE.log"

# Create report directory
mkdir -p "$REPORT_DIR"

# Set environment
export MACIE_JOB_TIMEOUT_MINUTES=45
export AWS_REGION=us-east-1

# Run audit
echo "Starting S3 security audit - $AUDIT_DATE" | tee "$LOG_FILE"
./s3auditor 2>&1 | tee -a "$LOG_FILE"

# Generate summary
echo "Audit completed - $AUDIT_DATE" | tee -a "$LOG_FILE"
echo "Report saved to: $LOG_FILE"

# Optional: Send to monitoring system
# curl -X POST "https://monitoring.example.com/audit" \
#      -H "Content-Type: application/json" \
#      -d "{\"date\":\"$AUDIT_DATE\",\"status\":\"completed\"}"
```

This usage guide provides comprehensive information for effectively using the AWS S3 Bucket Auditor in various scenarios and environments.