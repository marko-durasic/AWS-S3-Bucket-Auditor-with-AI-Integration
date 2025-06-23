package awsutils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock S3 client interface
type mockS3Client struct {
	mock.Mock
}

// Required interface methods
func (m *mockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func (m *mockS3Client) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.GetBucketLocationOutput), args.Error(1)
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

func TestGetBucketEncryption(t *testing.T) {
	tests := []struct {
		name          string
		bucketName    string
		mockSetup     func(*mockS3Client)
		expectedValue string
		expectError   bool
	}{
		{
			name:       "Bucket with AES256 encryption",
			bucketName: "encrypted-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetBucketEncryption", mock.Anything, mock.Anything).Return(
					&s3.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{
									ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
										SSEAlgorithm: types.ServerSideEncryptionAes256,
									},
								},
							},
						},
					}, nil)
			},
			expectedValue: "AES256",
			expectError:   false,
		},
		{
			name:       "Bucket without encryption",
			bucketName: "unencrypted-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetBucketEncryption", mock.Anything, mock.Anything).Return(
					&s3.GetBucketEncryptionOutput{}, &types.NoSuchBucket{})
			},
			expectedValue: "Not Enabled",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockS3Client)
			tt.mockSetup(mockClient)

			result, err := GetBucketEncryption(mockClient, tt.bucketName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestGetBucketVersioning(t *testing.T) {
	tests := []struct {
		name          string
		bucketName    string
		mockSetup     func(*mockS3Client)
		expectedValue string
		expectError   bool
	}{
		{
			name:       "Bucket with versioning enabled",
			bucketName: "versioned-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetBucketVersioning", mock.Anything, mock.Anything).Return(
					&s3.GetBucketVersioningOutput{
						Status: "Enabled",
					}, nil)
			},
			expectedValue: "Enabled",
			expectError:   false,
		},
		{
			name:       "Bucket with versioning disabled",
			bucketName: "unversioned-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetBucketVersioning", mock.Anything, mock.Anything).Return(
					&s3.GetBucketVersioningOutput{}, nil)
			},
			expectedValue: "Disabled",
			expectError:   false,
		},
		{
			name:       "Error getting versioning status",
			bucketName: "error-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetBucketVersioning", mock.Anything, mock.Anything).Return(
					&s3.GetBucketVersioningOutput{}, &types.NoSuchBucket{})
			},
			expectedValue: "Unknown",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockS3Client)
			tt.mockSetup(mockClient)

			result, err := GetBucketVersioning(mockClient, tt.bucketName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestIsBucketPublic(t *testing.T) {
	tests := []struct {
		name          string
		bucketName    string
		mockSetup     func(*mockS3Client)
		expectedValue bool
		expectError   bool
	}{
		{
			name:       "Bucket with public access blocked",
			bucketName: "private-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetPublicAccessBlock", mock.Anything, mock.Anything).Return(
					&s3.GetPublicAccessBlockOutput{
						PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
							BlockPublicAcls:       aws.Bool(true),
							BlockPublicPolicy:     aws.Bool(true),
							IgnorePublicAcls:      aws.Bool(true),
							RestrictPublicBuckets: aws.Bool(true),
						},
					}, nil)
			},
			expectedValue: false,
			expectError:   false,
		},
		{
			name:       "Bucket with public ACL",
			bucketName: "public-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetPublicAccessBlock", mock.Anything, mock.Anything).Return(
					&s3.GetPublicAccessBlockOutput{
						PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
							BlockPublicAcls:       aws.Bool(false),
							BlockPublicPolicy:     aws.Bool(false),
							IgnorePublicAcls:      aws.Bool(false),
							RestrictPublicBuckets: aws.Bool(false),
						},
					}, nil)
				m.On("GetBucketAcl", mock.Anything, mock.Anything).Return(
					&s3.GetBucketAclOutput{
						Grants: []types.Grant{
							{
								Grantee: &types.Grantee{
									URI: aws.String("http://acs.amazonaws.com/groups/global/AllUsers"),
								},
							},
						},
					}, nil)
			},
			expectedValue: true,
			expectError:   false,
		},
		{
			name:       "Bucket with incomplete public access block",
			bucketName: "partial-block-bucket",
			mockSetup: func(m *mockS3Client) {
				m.On("GetPublicAccessBlock", mock.Anything, mock.Anything).Return(
					&s3.GetPublicAccessBlockOutput{
						PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
							BlockPublicAcls: aws.Bool(true),
							// BlockPublicPolicy missing
							IgnorePublicAcls:      aws.Bool(true),
							RestrictPublicBuckets: aws.Bool(true),
						},
					}, nil)
				m.On("GetBucketAcl", mock.Anything, mock.Anything).Return(
					&s3.GetBucketAclOutput{}, nil)
			},
			expectedValue: false,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockS3Client)
			tt.mockSetup(mockClient)

			result, err := IsBucketPublic(mockClient, tt.bucketName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}
