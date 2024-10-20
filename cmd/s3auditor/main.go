package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
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
	// Setup logging to a file and the console
	logFile, err := os.OpenFile("s3_audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Welcome screen with ASCII art
	figure.NewFigure("S3 Auditor", "slant", true).Print()
	color.Cyan("\nWelcome to the AWS S3 Bucket Auditor!\n")

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		color.Red("Error: Unable to load AWS SDK config: %v", err)
		log.Printf("Error: Unable to load AWS SDK config: %v", err)
		return
	}

	// Create S3 and Macie clients
	s3Client := s3.NewFromConfig(cfg)
	macieClient := macie2.NewFromConfig(cfg)

	// Main menu
	for {
		actions := []string{"List S3 Buckets", "Audit a Bucket", "Exit"}
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: actions,
		}

		_, result, err := prompt.Run()

		if err != nil {
			color.Red("Error: %v", err)
			log.Printf("Error: %v", err)
			return
		}

		switch result {
		case "List S3 Buckets":
			listBuckets(s3Client)
		case "Audit a Bucket":
			auditBucket(cfg, s3Client, macieClient)
		case "Exit":
			color.Green("Goodbye! Stay secure.")
			os.Exit(0)
		}
	}
}

// listBuckets retrieves a list of all S3 buckets
func listBuckets(s3Client *s3.Client) {
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

// auditBucket allows the user to select a bucket to audit
func auditBucket(cfg aws.Config, s3Client *s3.Client, macieClient *macie2.Client) {
	color.Cyan("\nSelect an S3 bucket to audit...\n")
	log.Println("Select an S3 bucket to audit...")

	// Retrieve the list of buckets
	result, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		color.Red("Error: Unable to list S3 buckets: %v", err)
		log.Printf("Error: Unable to list S3 buckets: %v", err)
		return
	}

	// Check if there are no buckets
	if len(result.Buckets) == 0 {
		color.Yellow("No S3 buckets available to audit.\n")
		log.Println("No S3 buckets available to audit.")
		return
	}

	// Convert bucket names to strings for prompt
	bucketNames := []string{}
	for _, bucket := range result.Buckets {
		bucketNames = append(bucketNames, *bucket.Name)
	}

	// Prompt user to select a bucket
	prompt := promptui.Select{
		Label: "Choose a bucket to audit",
		Items: bucketNames,
	}

	_, selectedBucket, err := prompt.Run()
	if err != nil {
		color.Red("Error: %v", err)
		log.Printf("Error: %v", err)
		return
	}

	// Confirm choice and begin auditing
	color.Yellow("Auditing bucket: %s", selectedBucket)
	color.Cyan("\nPerforming security checks...")
	log.Printf("Auditing bucket: %s", selectedBucket)

	// Here we reuse the audit logic
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(bucketName string) {
		defer wg.Done()
		bucketInfo := BucketInfo{Name: bucketName}

		color.Cyan("Auditing bucket: %s", bucketName)
		log.Printf("Auditing bucket: %s", bucketName)

		// Get bucket region
		region, err := getBucketRegion(s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to get region for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to get region for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.Region = region

		// Check if bucket is public
		public, err := isBucketPublic(s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to check public access for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to check public access for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.IsPublic = public

		// Check encryption status
		encryption, err := getBucketEncryption(s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to get encryption for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to get encryption for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.Encryption = encryption

		// Check versioning status
		versioningStatus, err := getBucketVersioning(s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to get versioning status for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to get versioning status for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.VersioningStatus = versioningStatus

		// Check for sensitive data using Macie
		sensitiveData, err := checkSensitiveData(cfg, macieClient, bucketName)
		if err != nil {
			color.Red("Error: Unable to check sensitive data for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to check sensitive data for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.SensitiveData = sensitiveData

		// Print the report for this bucket
		printBucketReport(bucketInfo)
	}(selectedBucket)

	wg.Wait()
}

// The rest of your functions like `getBucketRegion`, `isBucketPublic`, `getBucketEncryption`, etc., remain the same.
// Just add logging similarly as we did above for error tracking and auditing.

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
		log.Printf("Error: failed to retrieve account ID: %v", err)
		return false, fmt.Errorf("Error: failed to retrieve account ID: %w", err)
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
		log.Printf("Error: failed to create Macie classification job: %v", err)
		return false, fmt.Errorf("Error: failed to create Macie classification job: %w", err)
	}

	jobID = *createJobOutput.JobId
	color.Yellow("🔍 Macie classification job created with Job ID: %s\n", jobID)
	log.Printf("Macie classification job created with Job ID: %s", jobID)

	// Set a timeout for the polling loop (e.g., 40 minutes)
	timeout := time.After(40 * time.Minute)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Start the progress bar
	bar := progressbar.NewOptions(100,
		progressbar.OptionSetDescription("Performing Macie Classification..."),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	startTime := time.Now()
	jobCompleted := false
	attempts := 0

	for {
		select {
		case <-timeout:
			log.Printf("Error: Macie job timed out after 40 minutes. Job ID: %s", jobID)
			return false, fmt.Errorf("Error: Macie job timed out after 40 minutes. Job ID: %s", jobID)
		case <-ticker.C:
			attempts++
			describeJobInput := &macie2.DescribeClassificationJobInput{
				JobId: aws.String(jobID),
			}

			describeJobOutput, err := macieClient.DescribeClassificationJob(context.Background(), describeJobInput)
			if err != nil {
				log.Printf("Error: failed to describe Macie classification job: %v", err)
				return false, fmt.Errorf("Error: failed to describe Macie classification job: %w", err)
			}

			// Estimate progress based on elapsed time
			elapsedTime := time.Since(startTime)
			progress := (elapsedTime.Minutes() / 40) * 100 // Rough estimate assuming job will finish in 40 minutes
			bar.Set(int(progress))

			// If job is completed, set the progress bar to 100% and break
			if describeJobOutput.JobStatus == types.JobStatusComplete {
				bar.Set(100)
				color.Green("✅ Macie job completed for bucket: %s\n", bucketName)
				log.Printf("Macie job completed for bucket: %s", bucketName)
				jobCompleted = true
				break // Exit loop when job is complete
			} else if describeJobOutput.JobStatus == types.JobStatusUserPaused ||
				describeJobOutput.JobStatus == types.JobStatusCancelled ||
				describeJobOutput.JobStatus == types.JobStatusPaused {
				color.Red("Error: Macie job did not complete successfully. Status: %s", describeJobOutput.JobStatus)
				log.Printf("Error: Macie job did not complete successfully. Status: %s", describeJobOutput.JobStatus)
				return false, nil
			}
		}

		if jobCompleted {
			break
		}
	}

	// Exit loop if job is completed
	if !jobCompleted {
		log.Printf("Error: Macie job did not complete within the allowed time. Job ID: %s", jobID)
		return false, fmt.Errorf("Error: Macie job did not complete within the allowed time")
	}

	// List findings after the job completes using FindingCriteria
	findingsInput := &macie2.ListFindingsInput{
		FindingCriteria: &types.FindingCriteria{
			Criterion: map[string]types.CriterionAdditionalProperties{
				"classificationDetails.jobId": {
					Eq: []string{jobID},
				},
			},
		},
	}

	findingsOutput, err := macieClient.ListFindings(context.Background(), findingsInput)
	if err != nil {
		log.Printf("Error: failed to list Macie findings: %v", err)
		return false, fmt.Errorf("Error: failed to list Macie findings: %w", err)
	}

	if len(findingsOutput.FindingIds) == 0 {
		color.Green("✅ No sensitive data found.")
		log.Println("No sensitive data found.")
		return false, nil
	}

	// Get detailed information about the findings using GetFindings
	getFindingsInput := &macie2.GetFindingsInput{
		FindingIds: findingsOutput.FindingIds,
	}

	getFindingsOutput, err := macieClient.GetFindings(context.Background(), getFindingsInput)
	if err != nil {
		log.Printf("Error: failed to get findings details: %v", err)
		return false, fmt.Errorf("Error: failed to get findings details: %w", err)
	}

	// Output details of each finding
	for _, finding := range getFindingsOutput.Findings {
		color.Magenta("🛑 Finding ID: %s\nDetails: %v\n", *finding.Id, finding)
		log.Printf("Finding ID: %s, Details: %v", *finding.Id, finding)
	}

	// Return to the main menu after the report is complete
	color.Cyan("\nReturning to the main menu...\n")
	log.Println("Returning to the main menu...")

	return true, nil
}

// printBucketReport outputs the audit results for one bucket with colors
func printBucketReport(info BucketInfo) {
	color.Cyan("\nS3 Bucket Security Audit Report:")
	color.Cyan("=====================================================================")
	color.Green("Bucket Name      : %s", info.Name)
	color.Cyan("Region           : %s", info.Region)
	color.Yellow("Public Access    : %t", info.IsPublic)
	color.Cyan("Encryption       : %s", info.Encryption)
	color.Cyan("Versioning       : %s", info.VersioningStatus)
	if info.SensitiveData {
		color.Red("Sensitive Data   : %t", info.SensitiveData)
	} else {
		color.Green("Sensitive Data   : %t", info.SensitiveData)
	}
	color.Cyan("---------------------------------------------------------------------")
}
