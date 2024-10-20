package awsutils

import (
	"context"
	"log"

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