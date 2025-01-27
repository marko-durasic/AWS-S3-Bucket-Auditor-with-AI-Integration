package config

import "os"

func GetTestBucketPrefix() string {
	if prefix := os.Getenv("TEST_BUCKET_PREFIX"); prefix != "" {
		return prefix
	}
	return "s3auditor-test-"
}

func GetAWSRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	return "us-east-1"
}

func GetAWSProfile() string {
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		return profile
	}
	return "default"
}
