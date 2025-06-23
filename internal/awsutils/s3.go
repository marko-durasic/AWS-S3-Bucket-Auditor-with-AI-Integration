package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/models"
)

// S3ClientAPI defines the interface for S3 operations we use
type S3ClientAPI interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
	GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error)
	GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetPublicAccessBlock(ctx context.Context, params *s3.GetPublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error)
	GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
}

// ListBuckets returns a list of bucket names and their regions
func ListBuckets(s3Client S3ClientAPI) ([]models.BucketBasicInfo, error) {
	bucketNames, err := getBucketNames(context.Background(), s3Client)
	if err != nil {
		return nil, err
	}

	buckets := make([]models.BucketBasicInfo, len(bucketNames))
	for i, name := range bucketNames {
		region, err := GetBucketRegion(s3Client, name)
		if err != nil {
			region = "unknown"
		}
		buckets[i] = models.BucketBasicInfo{
			Name:   name,
			Region: region,
		}
	}

	return buckets, nil
}

// GetBucketRegion retrieves the region of the specified S3 bucket
func GetBucketRegion(s3Client S3ClientAPI, bucketName string) (string, error) {
	locOutput, err := s3Client.GetBucketLocation(context.Background(), &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "", err
	}

	region := string(locOutput.LocationConstraint)
	if region == "" {
		region = "us-east-1"
	}
	return region, nil
}

// IsBucketPublic checks if the bucket is publicly accessible
func IsBucketPublic(s3Client S3ClientAPI, bucketName string) (bool, error) {
	// Check Public Access Block configuration
	pabOutput, err := s3Client.GetPublicAccessBlock(context.Background(), &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		config := pabOutput.PublicAccessBlockConfiguration
		if config != nil &&
			aws.ToBool(config.BlockPublicAcls) &&
			aws.ToBool(config.BlockPublicPolicy) &&
			aws.ToBool(config.IgnorePublicAcls) &&
			aws.ToBool(config.RestrictPublicBuckets) {
			return false, nil
		}
	}

	// Check bucket ACL
	aclOutput, err := s3Client.GetBucketAcl(context.Background(), &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return false, err
	}

	for _, grant := range aclOutput.Grants {
		if grant.Grantee != nil && grant.Grantee.URI != nil {
			if *grant.Grantee.URI == "http://acs.amazonaws.com/groups/global/AllUsers" ||
				*grant.Grantee.URI == "http://acs.amazonaws.com/groups/global/AuthenticatedUsers" {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetBucketEncryption checks if server-side encryption is enabled
func GetBucketEncryption(s3Client S3ClientAPI, bucketName string) (string, error) {
	encryptionOutput, err := s3Client.GetBucketEncryption(context.Background(), &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "Not Enabled", err
	}

	if len(encryptionOutput.ServerSideEncryptionConfiguration.Rules) > 0 {
		return string(encryptionOutput.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm), nil
	}

	return "Not Enabled", nil
}

// GetBucketVersioning checks if versioning is enabled
func GetBucketVersioning(s3Client S3ClientAPI, bucketName string) (string, error) {
	versioningOutput, err := s3Client.GetBucketVersioning(context.Background(), &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "Unknown", err
	}

	if versioningOutput.Status == "Enabled" {
		return "Enabled", nil
	}

	return "Disabled", nil
}

// GetBucketNames returns a slice of bucket names
func getBucketNames(ctx context.Context, s3Client S3ClientAPI) ([]string, error) {
	result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	bucketNames := make([]string, len(result.Buckets))
	for i, bucket := range result.Buckets {
		bucketNames[i] = *bucket.Name
	}
	return bucketNames, nil
}
