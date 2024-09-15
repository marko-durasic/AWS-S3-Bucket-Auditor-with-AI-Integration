package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// BucketInfo holds information about an S3 bucket
type BucketInfo struct {
	Name             string
	Region           string
	IsPublic         bool
	Encryption       string
	VersioningStatus string
	SensitiveData    bool
}

func main() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}

	// Create S3 and Macie clients
	s3Client := s3.NewFromConfig(cfg)
	macieClient := macie2.NewFromConfig(cfg)

	// List S3 buckets
	buckets, err := listBuckets(s3Client)
	if err != nil {
		log.Fatalf("Unable to list S3 buckets: %v", err)
	}

	// Audit buckets
	bucketInfos := auditBuckets(cfg, s3Client, macieClient, buckets)

	// Print report
	printReport(bucketInfos)
}

// listBuckets retrieves a list of all S3 buckets
func listBuckets(s3Client *s3.Client) ([]string, error) {
	result, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var buckets []string
	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}
	return buckets, nil
}

// auditBuckets performs security checks on each bucket
func auditBuckets(cfg aws.Config, s3Client *s3.Client, macieClient *macie2.Client, buckets []string) []BucketInfo {
	var wg sync.WaitGroup
	bucketInfos := make([]BucketInfo, len(buckets))

	for i, bucketName := range buckets {
		wg.Add(1)
		go func(i int, bucketName string) {
			defer wg.Done()
			bucketInfo := BucketInfo{Name: bucketName}

			// Get bucket region
			region, err := getBucketRegion(s3Client, bucketName)
			if err != nil {
				log.Printf("Unable to get region for bucket %s: %v", bucketName, err)
				return
			}
			bucketInfo.Region = region

			// Check if bucket is public
			public, err := isBucketPublic(s3Client, bucketName)
			if err != nil {
				log.Printf("Unable to check public access for bucket %s: %v", bucketName, err)
				return
			}
			bucketInfo.IsPublic = public

			// Check encryption status
			encryption, err := getBucketEncryption(s3Client, bucketName)
			if err != nil {
				log.Printf("Unable to get encryption for bucket %s: %v", bucketName, err)
				return
			}
			bucketInfo.Encryption = encryption

			// Check versioning status
			versioningStatus, err := getBucketVersioning(s3Client, bucketName)
			if err != nil {
				log.Printf("Unable to get versioning status for bucket %s: %v", bucketName, err)
				return
			}
			bucketInfo.VersioningStatus = versioningStatus

			// Check for sensitive data using Macie
			sensitiveData, err := checkSensitiveData(cfg, macieClient, bucketName)
			if err != nil {
				log.Printf("Unable to check sensitive data for bucket %s: %v", bucketName, err)
				return
			}
			bucketInfo.SensitiveData = sensitiveData

			bucketInfos[i] = bucketInfo
		}(i, bucketName)
	}

	wg.Wait()
	return bucketInfos
}

// getBucketRegion retrieves the region of the bucket
func getBucketRegion(s3Client *s3.Client, bucketName string) (string, error) {
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

// isBucketPublic checks if the bucket is publicly accessible
func isBucketPublic(s3Client *s3.Client, bucketName string) (bool, error) {
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

// getBucketEncryption checks if server-side encryption is enabled
func getBucketEncryption(s3Client *s3.Client, bucketName string) (string, error) {
	encryptionOutput, err := s3Client.GetBucketEncryption(context.Background(), &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "Not Enabled", nil
	}

	if len(encryptionOutput.ServerSideEncryptionConfiguration.Rules) > 0 {
		return string(encryptionOutput.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm), nil
	}

	return "Not Enabled", nil
}

// getBucketVersioning checks if versioning is enabled
func getBucketVersioning(s3Client *s3.Client, bucketName string) (string, error) {
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

// checkSensitiveData uses Macie to create a classification job to check for sensitive data in the bucket
func checkSensitiveData(cfg aws.Config, macieClient *macie2.Client, bucketName string) (bool, error) {
	// Retrieve AWS Account ID
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return false, fmt.Errorf("failed to retrieve account ID: %w", err)
	}

	// Define a unique job ID for the Macie classification job
	jobID := fmt.Sprintf("s3-audit-%s-%d", bucketName, time.Now().Unix())

	// Create the classification job for the specified S3 bucket
	input := &macie2.CreateClassificationJobInput{
		JobType: types.JobTypeOneTime,
		Name:    aws.String(jobID),
		S3JobDefinition: &types.S3JobDefinition{
			BucketDefinitions: []types.S3BucketDefinitionForJob{
				{
					AccountId: aws.String(*identity.Account),
					Buckets:   []string{bucketName},
				},
			},
		},
	}

	// Create the Macie classification job
	createJobOutput, err := macieClient.CreateClassificationJob(context.Background(), input)
	if err != nil {
		return false, fmt.Errorf("failed to create Macie classification job: %w", err)
	}

	jobID = *createJobOutput.JobId
	fmt.Printf("Macie classification job created with Job ID: %s\n", jobID)

	// Set a timeout for the polling loop (e.g., 15 minutes)
	timeout := time.After(15 * time.Minute)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	attempts := 0
	for {
		select {
		case <-timeout:
			return false, fmt.Errorf("Macie job timed out. Job ID: %s", jobID)
		case <-ticker.C:
			attempts++
			describeJobInput := &macie2.DescribeClassificationJobInput{
				JobId: aws.String(jobID),
			}

			describeJobOutput, err := macieClient.DescribeClassificationJob(context.Background(), describeJobInput)
			if err != nil {
				return false, fmt.Errorf("failed to describe Macie classification job: %w", err)
			}

			// Log current job status
			fmt.Printf("Attempt %d: Current job status: %s\n", attempts, describeJobOutput.JobStatus)

			// Check the job status using available status types
			if describeJobOutput.JobStatus == types.JobStatusComplete {
				fmt.Printf("Macie job completed.\n")
				break
			} else if describeJobOutput.JobStatus == types.JobStatusUserPaused ||
				describeJobOutput.JobStatus == types.JobStatusCancelled ||
				describeJobOutput.JobStatus == types.JobStatusPaused {
				return false, fmt.Errorf("Macie job did not complete successfully. Status: %s", describeJobOutput.JobStatus)
			}
		}
	}

	// List findings after the job completes using FindingCriteria
	findingsInput := &macie2.ListFindingsInput{
		FindingCriteria: &types.FindingCriteria{
			Criterion: map[string]types.CriterionAdditionalProperties{
				"jobId": {
					Eq: []string{jobID},
				},
			},
		},
	}

	findingsOutput, err := macieClient.ListFindings(context.Background(), findingsInput)
	if err != nil {
		return false, fmt.Errorf("failed to list Macie findings: %w", err)
	}

	if len(findingsOutput.FindingIds) == 0 {
		fmt.Println("No sensitive data found.")
		return false, nil
	}

	// Get detailed information about the findings using GetFindings
	getFindingsInput := &macie2.GetFindingsInput{
		FindingIds: findingsOutput.FindingIds,
	}

	getFindingsOutput, err := macieClient.GetFindings(context.Background(), getFindingsInput)
	if err != nil {
		return false, fmt.Errorf("failed to get findings details: %w", err)
	}

	// Output details of each finding
	for _, finding := range getFindingsOutput.Findings {
		fmt.Printf("Finding ID: %s\nDetails: %v\n", *finding.Id, finding)
	}

	return true, nil
}

// printReport outputs the audit results
func printReport(bucketInfos []BucketInfo) {
	fmt.Println("S3 Bucket Security Audit Report:")
	fmt.Println("=====================================================================")
	for _, info := range bucketInfos {
		if info.Name == "" {
			continue
		}
		fmt.Printf("Bucket Name      : %s\n", info.Name)
		fmt.Printf("Region           : %s\n", info.Region)
		fmt.Printf("Public Access    : %t\n", info.IsPublic)
		fmt.Printf("Encryption       : %s\n", info.Encryption)
		fmt.Printf("Versioning       : %s\n", info.VersioningStatus)
		fmt.Printf("Sensitive Data   : %t\n", info.SensitiveData)
		fmt.Println("---------------------------------------------------------------------")
	}
}
