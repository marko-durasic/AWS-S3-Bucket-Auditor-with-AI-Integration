# AWS S3 Bucket Auditor - Technical Reference

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Design Patterns](#design-patterns)
3. [Package Structure](#package-structure)
4. [Testing Framework](#testing-framework)
5. [Error Handling](#error-handling)
6. [Performance Considerations](#performance-considerations)
7. [Security Implementation](#security-implementation)
8. [Extension Points](#extension-points)
9. [Development Guidelines](#development-guidelines)

## Architecture Overview

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │   UI Layer      │    │  Main Entry     │
│                 │    │                 │    │                 │
│ - Menu System   │    │ - Display       │    │ - Application   │
│ - User Input    │    │ - Colors        │    │   Bootstrap     │
│ - Navigation    │    │ - Progress      │    │ - Client Init   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────┐
         │                Business Logic                       │
         └─────────────────────────────────────────────────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────┐
    │                            │                            │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Audit Engine   │    │  AWS Utilities  │    │  Configuration  │
│                 │    │                 │    │                 │
│ - Scanner       │    │ - S3 Client     │    │ - Environment   │
│ - Report Gen    │    │ - Macie Client  │    │ - Timeouts      │
│ - Data Flow     │    │ - STS Client    │    │ - Defaults      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────┐
         │                  Data Models                        │
         │                                                     │
         │ - BucketBasicInfo  - BucketInfo  - AWSClients       │
         └─────────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────┐
         │                 AWS Services                        │
         │                                                     │
         │    S3 API    │    Macie API    │    STS API         │
         └─────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### CLI Layer (`internal/cli`)
- **Purpose**: User interaction and workflow orchestration
- **Responsibilities**:
  - Menu navigation and user input handling
  - Bucket selection and filtering
  - Audit workflow coordination
  - Display formatting and presentation

#### UI Layer (`internal/ui`)
- **Purpose**: Visual presentation and user feedback
- **Responsibilities**:
  - Welcome screen and branding
  - Color-coded messaging (error, success, info)
  - Progress indication
  - ASCII art and visual elements

#### Audit Engine (`internal/audit`)
- **Purpose**: Core security analysis functionality
- **Responsibilities**:
  - Bucket security scanning
  - Macie integration and job management
  - Report generation and formatting
  - Concurrent processing coordination

#### AWS Utilities (`internal/awsutils`)
- **Purpose**: AWS service abstraction and integration
- **Responsibilities**:
  - Client initialization and configuration
  - Service-specific operations (S3, Macie, STS)
  - Interface definitions for testability
  - Error handling and retry logic

#### Configuration (`internal/config`)
- **Purpose**: Application configuration management
- **Responsibilities**:
  - Environment variable handling
  - Default value management
  - Timeout and limit configuration

#### Logger (`internal/logger`)
- **Purpose**: Centralized logging functionality
- **Responsibilities**:
  - File-based logging setup
  - Log rotation and management
  - Structured logging support

## Design Patterns

### 1. Interface Segregation Pattern

**Implementation**: AWS client interfaces are segregated by service

```go
// S3ClientAPI - Only S3 operations needed
type S3ClientAPI interface {
    ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
    GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
    // ... other S3 methods
}

// MacieClientAPI - Only Macie operations needed
type MacieClientAPI interface {
    CreateClassificationJob(ctx context.Context, params *macie2.CreateClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.CreateClassificationJobOutput, error)
    // ... other Macie methods
}
```

**Benefits**:
- Improved testability with focused mock interfaces
- Reduced coupling between components
- Clear separation of concerns

### 2. Dependency Injection Pattern

**Implementation**: Scanner accepts interfaces rather than concrete types

```go
type Scanner struct {
    cfg         aws.Config
    s3Client    awsutils.S3ClientAPI    // Interface, not concrete type
    macieClient awsutils.MacieClientAPI // Interface, not concrete type
    stsClient   awsutils.STSClientAPI   // Interface, not concrete type
}

func NewScanner(cfg aws.Config, s3Client awsutils.S3ClientAPI, macieClient awsutils.MacieClientAPI, stsClient awsutils.STSClientAPI) *Scanner {
    return &Scanner{
        cfg:         cfg,
        s3Client:    s3Client,
        macieClient: macieClient,
        stsClient:   stsClient,
    }
}
```

**Benefits**:
- Easy unit testing with mock implementations
- Flexible client substitution
- Improved modularity

### 3. Factory Pattern

**Implementation**: Centralized AWS client creation

```go
type AWSClients struct {
    Config      aws.Config
    S3Client    *s3.Client
    MacieClient *macie2.Client
}

func NewAWSClients(ctx context.Context) (*AWSClients, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, err
    }

    return &AWSClients{
        Config:      cfg,
        S3Client:    s3.NewFromConfig(cfg),
        MacieClient: macie2.NewFromConfig(cfg),
    }, nil
}
```

**Benefits**:
- Centralized configuration management
- Consistent client initialization
- Simplified error handling

### 4. Strategy Pattern

**Implementation**: Different audit strategies can be implemented

```go
type AuditStrategy interface {
    AuditBucket(bucketName string) (*models.BucketInfo, error)
}

type StandardAuditStrategy struct {
    scanner *Scanner
}

func (s *StandardAuditStrategy) AuditBucket(bucketName string) (*models.BucketInfo, error) {
    // Standard audit implementation
}

type QuickAuditStrategy struct {
    scanner *Scanner
}

func (q *QuickAuditStrategy) AuditBucket(bucketName string) (*models.BucketInfo, error) {
    // Quick audit implementation (skip Macie)
}
```

### 5. Template Method Pattern

**Implementation**: Audit workflow with customizable steps

```go
func (s *Scanner) AuditBucket(bucketName string) error {
    // Template method defining audit workflow
    info := models.BucketInfo{Name: bucketName}
    
    // Step 1: Basic information gathering
    if err := s.gatherBasicInfo(&info); err != nil {
        return err
    }
    
    // Step 2: Security configuration analysis
    if err := s.analyzeSecurityConfig(&info); err != nil {
        return err
    }
    
    // Step 3: Sensitive data detection
    if err := s.detectSensitiveData(&info); err != nil {
        return err
    }
    
    // Step 4: Report generation
    PrintBucketReport(info)
    return nil
}
```

## Package Structure

### Directory Organization

```
aws-s3-bucket-auditor/
├── cmd/
│   └── s3auditor/
│       └── main.go              # Application entry point
├── internal/
│   ├── audit/                   # Core audit functionality
│   │   ├── scanner.go           # Main audit engine
│   │   ├── scanner_test.go      # Scanner tests
│   │   └── report.go            # Report generation
│   ├── awsutils/                # AWS service utilities
│   │   ├── client.go            # Client factory
│   │   ├── s3.go                # S3 operations
│   │   ├── s3_test.go           # S3 tests
│   │   ├── macie.go             # Macie interface
│   │   └── sts.go               # STS interface
│   ├── cli/                     # Command-line interface
│   │   ├── menu.go              # Menu system
│   │   └── display.go           # Bucket display
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Main configuration
│   │   └── test_config.go       # Test configuration
│   ├── logger/                  # Logging utilities
│   │   └── logger.go            # Logger initialization
│   ├── models/                  # Data models
│   │   └── bucket.go            # Bucket data structures
│   ├── testutils/               # Testing utilities
│   │   ├── config.go            # Test configuration
│   │   ├── helpers.go           # Test helpers
│   │   └── bucket_names.go      # Test bucket naming
│   └── ui/                      # User interface
│       └── display.go           # UI display functions
├── test/                        # Integration tests
│   └── integration/
│       └── audit_test.go        # End-to-end tests
├── tests/                       # Additional test files
├── docs/                        # Documentation
│   ├── API_DOCUMENTATION.md     # API reference
│   ├── USAGE_GUIDE.md           # Usage instructions
│   └── TECHNICAL_REFERENCE.md   # Technical details
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # Project overview
```

### Package Dependencies

```
cmd/s3auditor
    └── internal/awsutils
    └── internal/cli
    └── internal/logger
    └── internal/ui

internal/cli
    └── internal/audit
    └── internal/awsutils
    └── internal/models
    └── internal/ui

internal/audit
    └── internal/awsutils
    └── internal/config
    └── internal/models

internal/awsutils
    └── internal/models

internal/ui
    └── (no internal dependencies)

internal/config
    └── (no internal dependencies)

internal/logger
    └── (no internal dependencies)

internal/models
    └── (no internal dependencies)
```

## Testing Framework

### Unit Testing Strategy

#### Mock Interfaces

```go
// Example mock implementation for testing
type MockS3Client struct {
    ListBucketsFunc      func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
    GetBucketLocationFunc func(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}

func (m *MockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
    if m.ListBucketsFunc != nil {
        return m.ListBucketsFunc(ctx, params, optFns...)
    }
    return nil, errors.New("not implemented")
}
```

#### Test Structure

```go
func TestScanner_AuditBucket(t *testing.T) {
    tests := []struct {
        name           string
        bucketName     string
        s3Client       awsutils.S3ClientAPI
        macieClient    awsutils.MacieClientAPI
        stsClient      awsutils.STSClientAPI
        expectedError  bool
        expectedResult *models.BucketInfo
    }{
        {
            name:       "successful audit",
            bucketName: "test-bucket",
            s3Client:   &MockS3Client{/* mock setup */},
            // ... other mocks
            expectedError: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scanner := audit.NewScanner(aws.Config{}, tt.s3Client, tt.macieClient, tt.stsClient)
            err := scanner.AuditBucket(tt.bucketName)
            
            if tt.expectedError && err == nil {
                t.Error("expected error but got none")
            }
            if !tt.expectedError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

### Integration Testing

#### Test Environment Setup

```go
func TestMain(m *testing.M) {
    // Setup test environment
    if err := setup(); err != nil {
        log.Fatalf("Failed to setup test environment: %v", err)
    }
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    teardown()
    
    os.Exit(code)
}

func setup() error {
    // Verify AWS credentials
    testutils.RequireAWSEnv(&testing.T{})
    
    // Initialize test clients
    ctx := context.Background()
    clients, err := awsutils.NewAWSClients(ctx)
    if err != nil {
        return err
    }
    
    // Create test buckets if needed
    return createTestResources(clients)
}
```

#### Test Utilities

```go
// Test bucket creation with cleanup
func CreateTestBucket(t *testing.T, s3Client *s3.Client, bucketName string) func() {
    // Create bucket
    _, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
        Bucket: aws.String(bucketName),
    })
    require.NoError(t, err)
    
    // Return cleanup function
    return func() {
        // Delete all objects first
        deleteAllObjects(t, s3Client, bucketName)
        
        // Delete bucket
        _, err := s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
            Bucket: aws.String(bucketName),
        })
        if err != nil {
            t.Logf("Failed to delete test bucket %s: %v", bucketName, err)
        }
    }
}

// Generate unique test bucket names
func GetTestBucketName(prefix string) string {
    timestamp := time.Now().Unix()
    randomSuffix := rand.Intn(10000)
    return fmt.Sprintf("%s-%d-%d", prefix, timestamp, randomSuffix)
}
```

### Test Coverage

#### Running Tests

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./test/integration/...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark tests
go test -bench=. ./internal/...
```

#### Coverage Goals

- **Unit Tests**: >80% coverage for all packages
- **Integration Tests**: End-to-end workflow coverage
- **Performance Tests**: Benchmark critical paths

## Error Handling

### Error Categories

#### 1. AWS Service Errors

```go
func handleAWSError(err error) error {
    var awsErr smithy.APIError
    if errors.As(err, &awsErr) {
        switch awsErr.ErrorCode() {
        case "NoSuchBucket":
            return fmt.Errorf("bucket not found: %w", err)
        case "AccessDenied":
            return fmt.Errorf("insufficient permissions: %w", err)
        case "InvalidBucketName":
            return fmt.Errorf("invalid bucket name: %w", err)
        default:
            return fmt.Errorf("AWS service error: %w", err)
        }
    }
    return err
}
```

#### 2. Network Errors

```go
func isNetworkError(err error) bool {
    var netErr net.Error
    return errors.As(err, &netErr) && netErr.Timeout()
}

func handleNetworkError(err error) error {
    if isNetworkError(err) {
        return fmt.Errorf("network timeout - check connectivity: %w", err)
    }
    return err
}
```

#### 3. Configuration Errors

```go
func validateConfiguration() error {
    if os.Getenv("AWS_REGION") == "" {
        return errors.New("AWS_REGION environment variable is required")
    }
    
    timeout := config.GetMacieTimeout()
    if timeout < time.Minute {
        return errors.New("MACIE_JOB_TIMEOUT_MINUTES must be at least 1 minute")
    }
    
    return nil
}
```

### Error Propagation Strategy

#### 1. Fail Fast for Critical Errors

```go
func (s *Scanner) AuditBucket(bucketName string) error {
    // Validate inputs immediately
    if bucketName == "" {
        return errors.New("bucket name cannot be empty")
    }
    
    // Check bucket existence early
    _, err := awsutils.GetBucketRegion(s.s3Client, bucketName)
    if err != nil {
        return fmt.Errorf("bucket validation failed: %w", err)
    }
    
    // Continue with audit...
}
```

#### 2. Graceful Degradation for Non-Critical Errors

```go
func (s *Scanner) gatherBucketInfo(bucketName string) *models.BucketInfo {
    info := &models.BucketInfo{Name: bucketName}
    
    // Try to get encryption (non-critical)
    if encryption, err := awsutils.GetBucketEncryption(s.s3Client, bucketName); err != nil {
        log.Printf("Warning: Could not get encryption status for %s: %v", bucketName, err)
        info.Encryption = "Unknown"
    } else {
        info.Encryption = encryption
    }
    
    // Continue with other checks...
    return info
}
```

## Performance Considerations

### Concurrency Model

#### Goroutine Usage

```go
func (s *Scanner) AuditBucket(bucketName string) error {
    wg := sync.WaitGroup{}
    wg.Add(1)
    
    go func(bucketName string) {
        defer wg.Done()
        
        // Perform audit in goroutine
        bucketInfo := s.performAudit(bucketName)
        PrintBucketReport(bucketInfo)
    }(bucketName)
    
    wg.Wait()
    return nil
}
```

#### Parallel Operations

```go
func (s *Scanner) gatherBucketInfoParallel(bucketName string) *models.BucketInfo {
    info := &models.BucketInfo{Name: bucketName}
    
    // Channel for collecting results
    type result struct {
        field string
        value interface{}
        err   error
    }
    
    results := make(chan result, 4)
    
    // Start parallel operations
    go func() {
        region, err := awsutils.GetBucketRegion(s.s3Client, bucketName)
        results <- result{"region", region, err}
    }()
    
    go func() {
        encryption, err := awsutils.GetBucketEncryption(s.s3Client, bucketName)
        results <- result{"encryption", encryption, err}
    }()
    
    go func() {
        versioning, err := awsutils.GetBucketVersioning(s.s3Client, bucketName)
        results <- result{"versioning", versioning, err}
    }()
    
    go func() {
        isPublic, err := awsutils.IsBucketPublic(s.s3Client, bucketName)
        results <- result{"public", isPublic, err}
    }()
    
    // Collect results
    for i := 0; i < 4; i++ {
        r := <-results
        switch r.field {
        case "region":
            if r.err == nil {
                info.Region = r.value.(string)
            }
        case "encryption":
            if r.err == nil {
                info.Encryption = r.value.(string)
            }
        // ... handle other fields
        }
    }
    
    return info
}
```

### Memory Management

#### Efficient Data Structures

```go
// Use slices with known capacity
func ListBuckets(s3Client S3ClientAPI) ([]models.BucketBasicInfo, error) {
    bucketNames, err := getBucketNames(context.Background(), s3Client)
    if err != nil {
        return nil, err
    }
    
    // Pre-allocate slice with known capacity
    buckets := make([]models.BucketBasicInfo, 0, len(bucketNames))
    
    for _, name := range bucketNames {
        region, err := GetBucketRegion(s3Client, name)
        if err != nil {
            region = "unknown"
        }
        
        buckets = append(buckets, models.BucketBasicInfo{
            Name:   name,
            Region: region,
        })
    }
    
    return buckets, nil
}
```

#### Resource Cleanup

```go
func (s *Scanner) checkSensitiveData(bucketName string) (bool, error) {
    // Ensure cleanup of resources
    defer func() {
        // Cleanup Macie job if needed
        if jobID != "" {
            s.cleanupMacieJob(jobID)
        }
    }()
    
    // Create and monitor Macie job
    jobID, err := s.createMacieJob(bucketName)
    if err != nil {
        return false, err
    }
    
    return s.waitForJobCompletion(jobID)
}
```

### Caching Strategy

#### Client Reuse

```go
type clientCache struct {
    mu      sync.RWMutex
    clients map[string]*AWSClients
}

func (c *clientCache) GetClients(region string) (*AWSClients, error) {
    c.mu.RLock()
    if clients, exists := c.clients[region]; exists {
        c.mu.RUnlock()
        return clients, nil
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Double-check pattern
    if clients, exists := c.clients[region]; exists {
        return clients, nil
    }
    
    // Create new clients for region
    clients, err := createRegionalClients(region)
    if err != nil {
        return nil, err
    }
    
    c.clients[region] = clients
    return clients, nil
}
```

## Security Implementation

### Credential Handling

#### Secure Credential Chain

```go
func NewAWSClients(ctx context.Context) (*AWSClients, error) {
    // Use AWS SDK default credential chain
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithRegion(getDefaultRegion()),
        config.WithRetryMode(aws.RetryModeAdaptive),
        config.WithRetryMaxAttempts(3),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    return &AWSClients{
        Config:      cfg,
        S3Client:    s3.NewFromConfig(cfg),
        MacieClient: macie2.NewFromConfig(cfg),
    }, nil
}
```

#### Sensitive Data Protection

```go
func (s *Scanner) handleMacieFindings(findings []types.Finding) {
    for _, finding := range findings {
        // Log finding ID but not sensitive details
        log.Printf("Macie finding detected: ID=%s, Type=%s", 
            aws.ToString(finding.Id),
            finding.Type)
        
        // Display sanitized information to user
        color.Magenta("🛑 Finding ID: %s", aws.ToString(finding.Id))
        
        // Do not log or display actual sensitive data
    }
}
```

### Input Validation

#### Bucket Name Validation

```go
func validateBucketName(bucketName string) error {
    if bucketName == "" {
        return errors.New("bucket name cannot be empty")
    }
    
    if len(bucketName) < 3 || len(bucketName) > 63 {
        return errors.New("bucket name must be between 3 and 63 characters")
    }
    
    // Basic DNS-compliant validation
    matched, _ := regexp.MatchString(`^[a-z0-9.-]+$`, bucketName)
    if !matched {
        return errors.New("bucket name contains invalid characters")
    }
    
    return nil
}
```

#### Parameter Sanitization

```go
func sanitizeInput(input string) string {
    // Remove potentially dangerous characters
    re := regexp.MustCompile(`[^\w\-.]`)
    return re.ReplaceAllString(input, "")
}
```

## Extension Points

### Custom Audit Strategies

#### Interface Definition

```go
type AuditStrategy interface {
    Name() string
    AuditBucket(ctx context.Context, bucketName string) (*models.BucketInfo, error)
    SupportsFeature(feature string) bool
}

type QuickAuditStrategy struct {
    s3Client awsutils.S3ClientAPI
}

func (q *QuickAuditStrategy) Name() string {
    return "quick-audit"
}

func (q *QuickAuditStrategy) AuditBucket(ctx context.Context, bucketName string) (*models.BucketInfo, error) {
    // Implement quick audit without Macie
    info := &models.BucketInfo{Name: bucketName}
    
    // Only check basic S3 configurations
    region, _ := awsutils.GetBucketRegion(q.s3Client, bucketName)
    info.Region = region
    
    encryption, _ := awsutils.GetBucketEncryption(q.s3Client, bucketName)
    info.Encryption = encryption
    
    return info, nil
}

func (q *QuickAuditStrategy) SupportsFeature(feature string) bool {
    return feature != "sensitive-data-detection"
}
```

### Custom Reporters

#### Reporter Interface

```go
type Reporter interface {
    GenerateReport(info *models.BucketInfo) error
    Format() string
}

type JSONReporter struct {
    output io.Writer
}

func (j *JSONReporter) GenerateReport(info *models.BucketInfo) error {
    data, err := json.MarshalIndent(info, "", "  ")
    if err != nil {
        return err
    }
    
    _, err = j.output.Write(data)
    return err
}

func (j *JSONReporter) Format() string {
    return "json"
}
```

### Plugin Architecture

#### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Initialize(config map[string]interface{}) error
    Execute(ctx context.Context, data interface{}) (interface{}, error)
    Cleanup() error
}

type PluginManager struct {
    plugins map[string]Plugin
}

func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
    if err := plugin.Initialize(nil); err != nil {
        return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
    }
    
    pm.plugins[plugin.Name()] = plugin
    return nil
}
```

## Development Guidelines

### Code Style

#### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Types**: PascalCase (exported), camelCase (unexported)
- **Functions**: PascalCase (exported), camelCase (unexported)
- **Variables**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE for package-level

#### Documentation Standards

```go
// ListBuckets retrieves all S3 buckets in the account along with their regions.
// It returns a slice of BucketBasicInfo containing bucket names and regions.
// 
// The function handles cross-region buckets automatically and will attempt
// to determine the region for each bucket. If region detection fails for
// a bucket, it will be marked as "unknown".
//
// Example:
//   buckets, err := awsutils.ListBuckets(s3Client)
//   if err != nil {
//       return fmt.Errorf("failed to list buckets: %w", err)
//   }
//
// Returns:
//   - []models.BucketBasicInfo: slice of bucket information
//   - error: non-nil if the operation fails
func ListBuckets(s3Client S3ClientAPI) ([]models.BucketBasicInfo, error) {
    // Implementation...
}
```

### Testing Guidelines

#### Test Organization

```go
func TestListBuckets(t *testing.T) {
    t.Parallel() // Enable parallel execution
    
    tests := []struct {
        name          string
        mockS3Client  *MockS3Client
        expectedCount int
        expectError   bool
    }{
        {
            name: "successful bucket listing",
            mockS3Client: &MockS3Client{
                ListBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
                    return &s3.ListBucketsOutput{
                        Buckets: []types.Bucket{
                            {Name: aws.String("bucket1")},
                            {Name: aws.String("bucket2")},
                        },
                    }, nil
                },
            },
            expectedCount: 2,
            expectError:   false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Enable parallel execution for subtests
            
            buckets, err := awsutils.ListBuckets(tt.mockS3Client)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Len(t, buckets, tt.expectedCount)
        })
    }
}
```

#### Performance Testing

```go
func BenchmarkListBuckets(b *testing.B) {
    mockClient := &MockS3Client{
        ListBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
            // Simulate realistic response
            buckets := make([]types.Bucket, 100)
            for i := range buckets {
                buckets[i] = types.Bucket{
                    Name: aws.String(fmt.Sprintf("bucket-%d", i)),
                }
            }
            return &s3.ListBucketsOutput{Buckets: buckets}, nil
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := awsutils.ListBuckets(mockClient)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Contribution Guidelines

#### Pull Request Process

1. **Fork and Branch**: Create feature branch from main
2. **Implement**: Follow coding standards and add tests
3. **Test**: Ensure all tests pass and coverage is maintained
4. **Document**: Update documentation for API changes
5. **Review**: Submit PR with clear description

#### Code Review Checklist

- [ ] Code follows project style guidelines
- [ ] All tests pass and coverage is maintained
- [ ] Documentation is updated
- [ ] Error handling is appropriate
- [ ] Security considerations are addressed
- [ ] Performance impact is considered

This technical reference provides comprehensive coverage of the architectural decisions, implementation patterns, and development practices used in the AWS S3 Bucket Auditor project.