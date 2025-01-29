package testutils

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CreateTestBucket creates a bucket for testing and returns a cleanup function
// The cleanup function ensures all resources are properly deleted
func CreateTestBucket(t *testing.T, s3Client *s3.Client, bucketName string) func() {
	t.Helper()

	_, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		t.Fatalf("Failed to create test bucket: %v", err)
	}

	// Return cleanup function
	return func() {
		// First delete all objects in the bucket
		listOutput, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			t.Errorf("Failed to list objects in bucket: %v", err)
			return
		}

		// Delete each object
		for _, obj := range listOutput.Contents {
			_, err := s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			})
			if err != nil {
				t.Errorf("Failed to delete object %s: %v", *obj.Key, err)
			}
		}

		// Now delete the empty bucket
		_, err = s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			t.Errorf("Failed to cleanup test bucket: %v", err)
		}
	}
}

// SetupTestData uploads test files to a bucket
func SetupTestData(t *testing.T, s3Client *s3.Client, bucketName string) error {
	t.Helper()

	testFiles := map[string]string{
		"test.txt":      "This is a test file",
		"sensitive.txt": "Credit card: 4111-1111-1111-1111",
	}

	for filename, content := range testFiles {
		_, err := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(filename),
			Body:   strings.NewReader(content),
		})
		if err != nil {
			return fmt.Errorf("failed to upload test file %s: %w", filename, err)
		}
	}

	return nil
}
