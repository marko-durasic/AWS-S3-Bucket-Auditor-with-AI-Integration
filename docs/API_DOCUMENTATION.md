# AWS S3 Bucket Auditor - API Documentation

## Overview

The AWS S3 Bucket Auditor is a Go application that provides comprehensive security auditing capabilities for Amazon S3 buckets. It integrates with AWS S3, Macie, and STS services to analyze bucket configurations, detect sensitive data, and generate detailed security reports.

## Table of Contents

1. [Core Components](#core-components)
2. [Public APIs](#public-apis)
3. [Data Models](#data-models)
4. [Interfaces](#interfaces)
5. [CLI Components](#cli-components)
6. [UI Components](#ui-components)
7. [Configuration](#configuration)
8. [Usage Examples](#usage-examples)

## Core Components

### Main Application Entry Point

#### `cmd/s3auditor/main.go`

**Function: `main()`**
- **Description**: Application entry point that initializes logging, AWS clients, and starts the interactive CLI menu
- **Dependencies**: 
  - AWS SDK v2 (S3, Macie2, STS)
  - Internal packages for AWS utilities, CLI, logging, and UI
- **Flow**:
  1. Initializes logger with file output
  2. Shows welcome screen
  3. Creates AWS clients
  4. Runs interactive menu loop

**Example Usage**:
```bash
go run cmd/s3auditor/main.go
```

## Public APIs

### AWS Utilities Package (`internal/awsutils`)

#### Client Management

**Function: `NewAWSClients(ctx context.Context) (*AWSClients, error)`**
- **Description**: Creates and initializes AWS service clients using default configuration
- **Parameters**:
  - `ctx`: Context for the operation
- **Returns**: 
  - `*AWSClients`: Struct containing S3, Macie, and STS clients with shared AWS config
  - `error`: Error if client initialization fails
- **Example**:
```go
ctx := context.Background()
clients, err := awsutils.NewAWSClients(ctx)
if err != nil {
    log.Fatal(err)
}
```

#### S3 Operations

**Function: `ListBuckets(s3Client S3ClientAPI) ([]models.BucketBasicInfo, error)`**
- **Description**: Retrieves all S3 buckets in the account with their regions
- **Parameters**:
  - `s3Client`: S3 client interface implementation
- **Returns**:
  - `[]models.BucketBasicInfo`: Slice of bucket information
  - `error`: Error if listing fails
- **Example**:
```go
buckets, err := awsutils.ListBuckets(clients.S3Client)
if err != nil {
    log.Printf("Error listing buckets: %v", err)
}
```

**Function: `GetBucketRegion(s3Client S3ClientAPI, bucketName string) (string, error)`**
- **Description**: Determines the AWS region where the specified bucket is located
- **Parameters**:
  - `s3Client`: S3 client interface
  - `bucketName`: Name of the S3 bucket
- **Returns**:
  - `string`: AWS region (defaults to "us-east-1" if empty)
  - `error`: Error if region retrieval fails
- **Example**:
```go
region, err := awsutils.GetBucketRegion(s3Client, "my-bucket")
if err != nil {
    log.Printf("Error getting bucket region: %v", err)
}
```

**Function: `IsBucketPublic(s3Client S3ClientAPI, bucketName string) (bool, error)`**
- **Description**: Analyzes bucket public access configuration and ACLs to determine if bucket is publicly accessible
- **Parameters**:
  - `s3Client`: S3 client interface
  - `bucketName`: Name of the S3 bucket
- **Returns**:
  - `bool`: True if bucket is public, false otherwise
  - `error`: Error if analysis fails
- **Security Checks**:
  - Public Access Block configuration
  - Bucket ACL permissions for AllUsers and AuthenticatedUsers
- **Example**:
```go
isPublic, err := awsutils.IsBucketPublic(s3Client, "my-bucket")
if err != nil {
    log.Printf("Error checking public access: %v", err)
}
if isPublic {
    log.Printf("WARNING: Bucket is publicly accessible")
}
```

**Function: `GetBucketEncryption(s3Client S3ClientAPI, bucketName string) (string, error)`**
- **Description**: Retrieves the server-side encryption configuration for the bucket
- **Parameters**:
  - `s3Client`: S3 client interface
  - `bucketName`: Name of the S3 bucket
- **Returns**:
  - `string`: Encryption algorithm (e.g., "AES256", "aws:kms") or "Not Enabled"
  - `error`: Error if encryption check fails
- **Example**:
```go
encryption, err := awsutils.GetBucketEncryption(s3Client, "my-bucket")
if err != nil {
    log.Printf("Error getting encryption status: %v", err)
}
log.Printf("Encryption: %s", encryption)
```

**Function: `GetBucketVersioning(s3Client S3ClientAPI, bucketName string) (string, error)`**
- **Description**: Checks if versioning is enabled on the bucket
- **Parameters**:
  - `s3Client`: S3 client interface
  - `bucketName`: Name of the S3 bucket
- **Returns**:
  - `string`: "Enabled", "Disabled", or "Unknown"
  - `error`: Error if versioning check fails
- **Example**:
```go
versioning, err := awsutils.GetBucketVersioning(s3Client, "my-bucket")
if err != nil {
    log.Printf("Error getting versioning status: %v", err)
}
log.Printf("Versioning: %s", versioning)
```

### Audit Package (`internal/audit`)

#### Scanner Component

**Function: `NewScanner(cfg aws.Config, s3Client awsutils.S3ClientAPI, macieClient awsutils.MacieClientAPI, stsClient awsutils.STSClientAPI) *Scanner`**
- **Description**: Creates a new audit scanner with all required AWS service clients
- **Parameters**:
  - `cfg`: AWS configuration
  - `s3Client`: S3 client interface
  - `macieClient`: Macie client interface
  - `stsClient`: STS client interface
- **Returns**: `*Scanner`: Configured scanner instance
- **Example**:
```go
scanner := audit.NewScanner(cfg, s3Client, macieClient, stsClient)
```

**Method: `(s *Scanner) AuditBucket(bucketName string) error`**
- **Description**: Performs comprehensive security audit of a single S3 bucket
- **Parameters**:
  - `bucketName`: Name of the bucket to audit
- **Returns**: `error`: Error if audit fails
- **Audit Checks**:
  - Bucket region identification
  - Public access analysis
  - Encryption status
  - Versioning configuration
  - Sensitive data detection using Macie
- **Features**:
  - Concurrent execution using goroutines
  - Progress tracking with visual progress bar
  - Comprehensive error handling and logging
- **Example**:
```go
err := scanner.AuditBucket("my-sensitive-bucket")
if err != nil {
    log.Printf("Audit failed: %v", err)
}
```

#### Report Generation

**Function: `PrintBucketReport(info models.BucketInfo)`**
- **Description**: Generates and displays a formatted security audit report
- **Parameters**:
  - `info`: Complete bucket audit information
- **Output**: Colorized console report with:
  - Bucket name and region
  - Public access status (highlighted if public)
  - Encryption configuration
  - Versioning status
  - Sensitive data detection results
  - Audit execution duration
- **Example**:
```go
bucketInfo := models.BucketInfo{
    Name: "my-bucket",
    Region: "us-east-1",
    IsPublic: false,
    Encryption: "AES256",
    VersioningStatus: "Enabled",
    SensitiveData: false,
    AuditDuration: time.Minute * 2,
}
audit.PrintBucketReport(bucketInfo)
```

### CLI Package (`internal/cli`)

#### Menu System

**Function: `PromptMainMenu() (string, error)`**
- **Description**: Displays interactive main menu with navigation options
- **Returns**:
  - `string`: Selected menu option
  - `error`: Error if menu interaction fails
- **Menu Options**:
  - "List S3 Buckets": Browse and view bucket details
  - "Audit a Bucket": Perform security audit on selected bucket
  - "Exit": Terminate application
- **Features**:
  - Keyboard navigation with arrow keys
  - Ctrl+C interrupt handling
- **Example**:
```go
choice, err := cli.PromptMainMenu()
if err != nil {
    log.Printf("Menu error: %v", err)
}
```

**Function: `PromptForBucketSelection(s3Client *s3.Client) (string, error)`**
- **Description**: Interactive bucket selection with search functionality
- **Parameters**:
  - `s3Client`: S3 client for bucket listing
- **Returns**:
  - `string`: Selected bucket name
  - `error`: Error if selection fails or user cancels
- **Features**:
  - Real-time search by bucket name or region
  - Paginated display for large bucket lists
  - Exit option for returning to main menu
- **Example**:
```go
bucketName, err := cli.PromptForBucketSelection(s3Client)
if err == promptui.ErrInterrupt {
    return // User cancelled
}
if err != nil {
    log.Printf("Selection error: %v", err)
}
```

**Function: `HandleBucketAudit(cfg aws.Config, s3Client *s3.Client, macieClient *macie2.Client)`**
- **Description**: Complete workflow for bucket audit including selection and execution
- **Parameters**:
  - `cfg`: AWS configuration
  - `s3Client`: S3 service client
  - `macieClient`: Macie service client
- **Workflow**:
  1. Prompts user for bucket selection
  2. Creates audit scanner
  3. Executes comprehensive audit
  4. Displays results
- **Example**:
```go
cli.HandleBucketAudit(cfg, s3Client, macieClient)
```

#### Display Functions

**Function: `DisplayBucketsList(s3Client *s3.Client, buckets []models.BucketBasicInfo)`**
- **Description**: Interactive bucket browser with detailed view capability
- **Parameters**:
  - `s3Client`: S3 client for additional bucket information
  - `buckets`: List of buckets to display
- **Features**:
  - Interactive navigation
  - Detailed bucket information on selection
  - Search functionality
  - Real-time security status indicators
- **Displayed Information**:
  - Bucket name and region
  - Encryption status
  - Versioning configuration
  - Public access status (color-coded)
- **Example**:
```go
buckets, _ := awsutils.ListBuckets(s3Client)
cli.DisplayBucketsList(s3Client, buckets)
```

### UI Package (`internal/ui`)

#### Display Functions

**Function: `ShowWelcomeScreen()`**
- **Description**: Displays ASCII art welcome banner and application introduction
- **Features**:
  - Stylized "S3 Auditor" ASCII art
  - Colorized welcome message
- **Example**:
```go
ui.ShowWelcomeScreen()
```

**Function: `ShowError(format string, args ...interface{})`**
- **Description**: Displays error messages with red color formatting
- **Parameters**:
  - `format`: Printf-style format string
  - `args`: Format arguments
- **Example**:
```go
ui.ShowError("Failed to connect to AWS: %v", err)
```

**Function: `ShowSuccess(format string, args ...interface{})`**
- **Description**: Displays success messages with green color formatting
- **Parameters**:
  - `format`: Printf-style format string
  - `args`: Format arguments
- **Example**:
```go
ui.ShowSuccess("Audit completed successfully!")
```

### Configuration Package (`internal/config`)

**Function: `GetMacieTimeout() time.Duration`**
- **Description**: Retrieves Macie job timeout from environment variable or default
- **Environment Variable**: `MACIE_JOB_TIMEOUT_MINUTES`
- **Default**: 40 minutes
- **Returns**: `time.Duration`: Timeout duration for Macie classification jobs
- **Example**:
```go
timeout := config.GetMacieTimeout()
log.Printf("Using Macie timeout: %v", timeout)
```

### Logger Package (`internal/logger`)

**Function: `InitLogger(logPath string) error`**
- **Description**: Initializes file-based logging for the application
- **Parameters**:
  - `logPath`: Path to log file
- **Returns**: `error`: Error if log file creation fails
- **Features**:
  - Append mode for persistent logging
  - File permissions: 0666
- **Example**:
```go
err := logger.InitLogger("s3_audit.log")
if err != nil {
    fmt.Printf("Logger initialization failed: %v", err)
}
```

## Data Models

### BucketBasicInfo (`internal/models`)

```go
type BucketBasicInfo struct {
    Name   string  // S3 bucket name
    Region string  // AWS region where bucket is located
}
```

**Usage**: Basic bucket information for listing and selection

### BucketInfo (`internal/models`)

```go
type BucketInfo struct {
    Name             string        // S3 bucket name
    Region           string        // AWS region
    IsPublic         bool          // Public access status
    Encryption       string        // Encryption algorithm or "Not Enabled"
    VersioningStatus string        // "Enabled", "Disabled", or "Unknown"
    SensitiveData    bool          // True if Macie detected sensitive data
    AuditDuration    time.Duration // Time taken for audit completion
}
```

**Usage**: Complete audit results for reporting and analysis

### AWSClients (`internal/awsutils`)

```go
type AWSClients struct {
    Config      aws.Config      // Shared AWS configuration
    S3Client    *s3.Client      // S3 service client
    MacieClient *macie2.Client  // Macie service client
}
```

**Usage**: Container for all AWS service clients with shared configuration

## Interfaces

### S3ClientAPI (`internal/awsutils`)

```go
type S3ClientAPI interface {
    ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
    GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
    GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error)
    GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
    GetPublicAccessBlock(ctx context.Context, params *s3.GetPublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error)
    GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
}
```

**Purpose**: Abstraction for S3 operations enabling testing and dependency injection

### MacieClientAPI (`internal/awsutils`)

```go
type MacieClientAPI interface {
    CreateClassificationJob(ctx context.Context, params *macie2.CreateClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.CreateClassificationJobOutput, error)
    DescribeClassificationJob(ctx context.Context, params *macie2.DescribeClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.DescribeClassificationJobOutput, error)
    ListFindings(ctx context.Context, params *macie2.ListFindingsInput, optFns ...func(*macie2.Options)) (*macie2.ListFindingsOutput, error)
    GetFindings(ctx context.Context, params *macie2.GetFindingsInput, optFns ...func(*macie2.Options)) (*macie2.GetFindingsOutput, error)
}
```

**Purpose**: Interface for Macie sensitive data detection operations

### STSClientAPI (`internal/awsutils`)

```go
type STSClientAPI interface {
    GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}
```

**Purpose**: Interface for AWS Security Token Service operations

## Usage Examples

### Complete Audit Workflow

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
    // Initialize AWS clients
    ctx := context.Background()
    clients, err := awsutils.NewAWSClients(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create audit scanner
    stsClient := sts.NewFromConfig(clients.Config)
    scanner := audit.NewScanner(
        clients.Config,
        clients.S3Client,
        clients.MacieClient,
        stsClient,
    )
    
    // Perform audit
    bucketName := "my-important-bucket"
    err = scanner.AuditBucket(bucketName)
    if err != nil {
        log.Printf("Audit failed: %v", err)
    }
}
```

### Bucket Security Analysis

```go
package main

import (
    "log"
    
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
)

func analyzeSecurityPosture(s3Client awsutils.S3ClientAPI, bucketName string) {
    // Check public access
    isPublic, err := awsutils.IsBucketPublic(s3Client, bucketName)
    if err != nil {
        log.Printf("Error checking public access: %v", err)
        return
    }
    
    // Check encryption
    encryption, err := awsutils.GetBucketEncryption(s3Client, bucketName)
    if err != nil {
        log.Printf("Error checking encryption: %v", err)
        return
    }
    
    // Check versioning
    versioning, err := awsutils.GetBucketVersioning(s3Client, bucketName)
    if err != nil {
        log.Printf("Error checking versioning: %v", err)
        return
    }
    
    // Security recommendations
    if isPublic {
        log.Printf("WARNING: Bucket %s is publicly accessible", bucketName)
    }
    
    if encryption == "Not Enabled" {
        log.Printf("RECOMMENDATION: Enable encryption for bucket %s", bucketName)
    }
    
    if versioning == "Disabled" {
        log.Printf("RECOMMENDATION: Enable versioning for bucket %s", bucketName)
    }
}
```

### Custom UI Integration

```go
package main

import (
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/ui"
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/models"
    "github.com/marko-durasic/aws-s3-bucket-auditor/internal/audit"
    "time"
)

func customAuditReport(bucketName string, results map[string]interface{}) {
    ui.ShowWelcomeScreen()
    
    // Create bucket info from results
    info := models.BucketInfo{
        Name:             bucketName,
        Region:           results["region"].(string),
        IsPublic:         results["isPublic"].(bool),
        Encryption:       results["encryption"].(string),
        VersioningStatus: results["versioning"].(string),
        SensitiveData:    results["sensitiveData"].(bool),
        AuditDuration:    results["duration"].(time.Duration),
    }
    
    // Display formatted report
    audit.PrintBucketReport(info)
    
    if info.IsPublic {
        ui.ShowError("CRITICAL: Bucket is publicly accessible!")
    } else {
        ui.ShowSuccess("Bucket access is properly restricted")
    }
}
```

## Error Handling

All public functions return errors that should be handled appropriately:

- **Network errors**: AWS API connectivity issues
- **Permission errors**: Insufficient IAM permissions
- **Resource errors**: Bucket not found or inaccessible
- **Timeout errors**: Macie job execution timeouts
- **Configuration errors**: Invalid AWS configuration

## Dependencies

- **AWS SDK Go v2**: Core AWS service integration
- **promptui**: Interactive CLI components
- **fatih/color**: Terminal color output
- **progressbar/v3**: Progress visualization
- **go-figure**: ASCII art generation

## Security Considerations

- **IAM Permissions**: Requires appropriate S3, Macie, and STS permissions
- **Sensitive Data**: Macie findings may contain sensitive information
- **Logging**: Audit logs may contain bucket names and security details
- **Network**: Requires internet connectivity for AWS API calls

This documentation provides comprehensive coverage of all public APIs, functions, and components in the AWS S3 Bucket Auditor application. Each function includes detailed descriptions, parameters, return values, and practical usage examples.