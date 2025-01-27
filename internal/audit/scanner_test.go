package audit

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	macie2types "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMacieClient mocks the Macie2 client interface
type MockMacieClient struct {
	mock.Mock
}

func (m *MockMacieClient) CreateClassificationJob(ctx context.Context, params *macie2.CreateClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.CreateClassificationJobOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*macie2.CreateClassificationJobOutput), args.Error(1)
}

func (m *MockMacieClient) DescribeClassificationJob(ctx context.Context, params *macie2.DescribeClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.DescribeClassificationJobOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*macie2.DescribeClassificationJobOutput), args.Error(1)
}

func (m *MockMacieClient) ListFindings(ctx context.Context, params *macie2.ListFindingsInput, optFns ...func(*macie2.Options)) (*macie2.ListFindingsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*macie2.ListFindingsOutput), args.Error(1)
}

func (m *MockMacieClient) GetFindings(ctx context.Context, params *macie2.GetFindingsInput, optFns ...func(*macie2.Options)) (*macie2.GetFindingsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*macie2.GetFindingsOutput), args.Error(1)
}

type mockS3Client struct {
	mock.Mock
}

// Implement required S3 methods
func (m *mockS3Client) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetBucketLocationOutput), args.Error(1)
}

func (m *mockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func (m *mockS3Client) GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetBucketEncryptionOutput), args.Error(1)
}

func (m *mockS3Client) GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetBucketVersioningOutput), args.Error(1)
}

func (m *mockS3Client) GetPublicAccessBlock(ctx context.Context, params *s3.GetPublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetPublicAccessBlockOutput), args.Error(1)
}

func (m *mockS3Client) GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetBucketAclOutput), args.Error(1)
}

// Add mock STS client
type mockSTSClient struct {
	mock.Mock
}

func (m *mockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params)
	return &sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
		Arn:     aws.String("arn:aws:iam::123456789012:user/test"),
		UserId:  aws.String("AIDATEST"),
	}, args.Error(1)
}

func TestScanner_AuditBucket(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		setupMocks func(*MockMacieClient, *mockS3Client, *mockSTSClient)
		wantErr    bool
	}{
		{
			name:       "Successful audit",
			bucketName: "test-bucket",
			setupMocks: func(m *MockMacieClient, s *mockS3Client, sts *mockSTSClient) {
				// Setup S3 mocks
				s.On("GetBucketLocation", mock.Anything, mock.Anything).Return(
					&s3.GetBucketLocationOutput{
						LocationConstraint: "us-east-1",
					}, nil)
				s.On("GetBucketEncryption", mock.Anything, mock.Anything).Return(
					&s3.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &s3types.ServerSideEncryptionConfiguration{
							Rules: []s3types.ServerSideEncryptionRule{
								{
									ApplyServerSideEncryptionByDefault: &s3types.ServerSideEncryptionByDefault{
										SSEAlgorithm: s3types.ServerSideEncryptionAes256,
									},
								},
							},
						},
					}, nil)
				s.On("GetBucketVersioning", mock.Anything, mock.Anything).Return(
					&s3.GetBucketVersioningOutput{
						Status: "Enabled",
					}, nil)
				s.On("GetPublicAccessBlock", mock.Anything, mock.Anything).Return(
					&s3.GetPublicAccessBlockOutput{
						PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
							BlockPublicAcls:       aws.Bool(true),
							BlockPublicPolicy:     aws.Bool(true),
							IgnorePublicAcls:      aws.Bool(true),
							RestrictPublicBuckets: aws.Bool(true),
						},
					}, nil)

				// Mock STS response - remove the return value since it's hardcoded in the mock
				sts.On("GetCallerIdentity", mock.Anything, mock.Anything).Return(nil, nil)

				// Setup Macie mocks
				m.On("CreateClassificationJob", mock.Anything, mock.Anything).Return(
					&macie2.CreateClassificationJobOutput{
						JobId: aws.String("test-job-id"),
					}, nil)
				m.On("DescribeClassificationJob", mock.Anything, mock.Anything).Return(
					&macie2.DescribeClassificationJobOutput{
						JobStatus: macie2types.JobStatusComplete,
					}, nil)
				m.On("ListFindings", mock.Anything, mock.Anything).Return(
					&macie2.ListFindingsOutput{
						FindingIds: []string{"test-finding"},
					}, nil)
				m.On("GetFindings", mock.Anything, mock.Anything).Return(
					&macie2.GetFindingsOutput{
						Findings: []macie2types.Finding{},
					}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMacie := new(MockMacieClient)
			mockS3 := new(mockS3Client)
			mockSTS := new(mockSTSClient)
			tt.setupMocks(mockMacie, mockS3, mockSTS)

			scanner := NewScanner(aws.Config{
				Region: "us-east-1",
			}, mockS3, mockMacie, mockSTS)
			err := scanner.AuditBucket(tt.bucketName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
