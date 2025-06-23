package audit

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/config"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/models"
	"github.com/schollz/progressbar/v3"
)

type Scanner struct {
	cfg         aws.Config
	s3Client    awsutils.S3ClientAPI
	macieClient awsutils.MacieClientAPI
	stsClient   awsutils.STSClientAPI
}

func NewScanner(cfg aws.Config, s3Client awsutils.S3ClientAPI, macieClient awsutils.MacieClientAPI, stsClient awsutils.STSClientAPI) *Scanner {
	return &Scanner{
		cfg:         cfg,
		s3Client:    s3Client,
		macieClient: macieClient,
		stsClient:   stsClient,
	}
}

func (s *Scanner) AuditBucket(bucketName string) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	startTime := time.Now()
	go func(bucketName string) {
		defer wg.Done()
		bucketInfo := models.BucketInfo{Name: bucketName}

		color.Cyan("Auditing bucket: %s", bucketName)
		log.Printf("Auditing bucket: %s", bucketName)

		// Get bucket region
		region, err := awsutils.GetBucketRegion(s.s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to get region for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to get region for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.Region = region

		// Check if bucket is public
		public, err := awsutils.IsBucketPublic(s.s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to check public access for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to check public access for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.IsPublic = public

		// Check encryption status
		encryption, err := awsutils.GetBucketEncryption(s.s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to get encryption for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to get encryption for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.Encryption = encryption

		// Check versioning status
		versioningStatus, err := awsutils.GetBucketVersioning(s.s3Client, bucketName)
		if err != nil {
			color.Red("Error: Unable to get versioning status for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to get versioning status for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.VersioningStatus = versioningStatus
		// Check for sensitive data using Macie
		sensitiveData, err := s.checkSensitiveData(bucketName)
		if err != nil {
			color.Red("Error: Unable to check sensitive data for bucket %s: %v", bucketName, err)
			log.Printf("Error: Unable to check sensitive data for bucket %s: %v", bucketName, err)
			return
		}
		bucketInfo.SensitiveData = sensitiveData

		bucketInfo.AuditDuration = time.Since(startTime)
		// Print the report for this bucket
		PrintBucketReport(bucketInfo)
	}(bucketName)

	wg.Wait()
	return nil
}

func (s *Scanner) checkSensitiveData(bucketName string) (bool, error) {
	// Retrieve AWS Account ID
	identity, err := s.stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
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
	createJobOutput, err := s.macieClient.CreateClassificationJob(context.Background(), input)
	if err != nil {
		log.Printf("Error: failed to create Macie classification job: %v", err)
		return false, fmt.Errorf("Error: failed to create Macie classification job: %w", err)
	}

	jobID = *createJobOutput.JobId
	color.Yellow("🔍 Macie classification job created with Job ID: %s\n", jobID)
	log.Printf("Macie classification job created with Job ID: %s", jobID)

	// Set a timeout for the polling loop
	timeout := time.After(config.GetMacieTimeout())
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
	)

	// Poll for job completion
	jobDone := false
	for !jobDone {
		select {
		case <-timeout:
			return false, fmt.Errorf("timeout waiting for Macie classification job completion")
		case <-ticker.C:
			// Get job status
			describeJobInput := &macie2.DescribeClassificationJobInput{
				JobId: aws.String(jobID),
			}

			describeJobOutput, err := s.macieClient.DescribeClassificationJob(context.Background(), describeJobInput)
			if err != nil {
				log.Printf("Error: failed to get job status: %v", err)
				return false, fmt.Errorf("Error: failed to get job status: %w", err)
			}

			// Update progress bar
			_ = bar.Add(1)

			// Check if job is complete
			if describeJobOutput.JobStatus == types.JobStatusComplete {
				_ = bar.Finish()
				jobDone = true
			} else if describeJobOutput.JobStatus == types.JobStatusUserPaused ||
				describeJobOutput.JobStatus == types.JobStatusCancelled ||
				describeJobOutput.JobStatus == types.JobStatusPaused {
				return false, fmt.Errorf("Macie classification job failed")
			}
		}
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

	findingsOutput, err := s.macieClient.ListFindings(context.Background(), findingsInput)
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

	getFindingsOutput, err := s.macieClient.GetFindings(context.Background(), getFindingsInput)
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
