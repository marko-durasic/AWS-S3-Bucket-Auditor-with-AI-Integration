package awsutils

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fatih/color"
)

// ListBuckets retrieves a list of all S3 buckets
func ListBuckets(s3Client *s3.Client) {
	color.Cyan("\nListing S3 Buckets...\n")
	log.Println("Listing S3 Buckets...")
	result, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		color.Red("Error: Unable to list S3 buckets: %v", err)
		log.Printf("Error: Unable to list S3 buckets: %v", err)
		return
	}

	if len(result.Buckets) == 0 {
		color.Yellow("No S3 buckets found.\n")
		log.Println("No S3 buckets found.")
		return
	}

	for _, bucket := range result.Buckets {
		color.Green("Bucket: %s", *bucket.Name)
		log.Printf("Bucket: %s", *bucket.Name)
	}
}

// GetBucketRegion retrieves the region of the specified S3 bucket
func GetBucketRegion(s3Client *s3.Client, bucketName string) (string, error) {
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
func IsBucketPublic(s3Client *s3.Client, bucketName string) (bool, error) {
	// Check Public Access Block configuration
	pabOutput, err := s3Client.GetPublicAccessBlock(context.Background(), &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		config := pabOutput.PublicAccessBlockConfiguration
		if config != nil && *config.BlockPublicAcls && *config.BlockPublicPolicy && *config.IgnorePublicAcls && *config.RestrictPublicBuckets {
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
