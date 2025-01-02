package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/models"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/ui"
)

func DisplayBucketsList(s3Client *s3.Client, buckets []models.BucketBasicInfo) {
	if len(buckets) == 0 {
		color.Yellow("\nNo S3 buckets found.")
		return
	}

	// Add Exit option to the list
	items := append(buckets, models.BucketBasicInfo{Name: "[ Exit ]", Region: ""})

	for {
		prompt := &promptui.Select{
			Label: "S3 Buckets List (Enter for details, Ctrl+C or Exit option to return)",
			Items: items,
			Size:  10,
			Templates: &promptui.SelectTemplates{
				Label: "{{ . }}",
				Active: func() string {
					return `{{ if eq .Name "[ Exit ]" }}` +
						`▸ {{ .Name | red }}` +
						`{{ else }}` +
						`▸ {{ .Name | cyan }} ({{ .Region | cyan }})` +
						`{{ end }}`
				}(),
				Inactive: func() string {
					return `{{ if eq .Name "[ Exit ]" }}` +
						`  {{ .Name | red }}` +
						`{{ else }}` +
						`  {{ .Name }} ({{ .Region }})` +
						`{{ end }}`
				}(),
				Selected: `{{ if eq .Name "[ Exit ]" }}` +
					`{{ .Name | red }}` +
					`{{ else }}` +
					`Selected bucket: {{ .Name | green }} ({{ .Region | blue }})` +
					`{{ end }}`,
			},
			Searcher: func(input string, index int) bool {
				bucket := items[index]
				if bucket.Name == "[ Exit ]" {
					return true
				}
				input = strings.ToLower(input)
				return strings.Contains(strings.ToLower(bucket.Name), input) ||
					strings.Contains(strings.ToLower(bucket.Region), input)
			},
			HideSelected: true,
		}

		idx, _, err := prompt.Run()
		if err == promptui.ErrInterrupt {
			return
		}
		if err != nil {
			ui.ShowError("Error displaying buckets: %v", err)
			log.Printf("Error displaying buckets: %v", err)
			return
		}

		// If Exit was selected, return
		if idx == len(buckets) {
			return
		}

		// Display details for selected bucket
		displayBucketDetails(s3Client, buckets[idx])
	}
}

func displayBucketDetails(s3Client *s3.Client, bucket models.BucketBasicInfo) {
	color.Cyan("\nBucket Details:")
	color.Cyan("=====================================================================")
	color.Green("Name              : %s", bucket.Name)
	color.Cyan("Region            : %s", bucket.Region)

	// Get encryption status
	encryption, err := awsutils.GetBucketEncryption(s3Client, bucket.Name)
	if err != nil {
		encryption = "Not Enabled"
	}
	color.Cyan("Encryption        : %s", encryption)

	// Get versioning status
	versioning, err := awsutils.GetBucketVersioning(s3Client, bucket.Name)
	if err != nil {
		versioning = "Unknown"
	}
	color.Cyan("Versioning        : %s", versioning)

	// Check if bucket is public
	isPublic, err := awsutils.IsBucketPublic(s3Client, bucket.Name)
	if err != nil {
		color.Yellow("Public Access     : Unknown")
	} else if isPublic {
		color.Red("Public Access     : Yes")
	} else {
		color.Green("Public Access     : No")
	}
	color.Cyan("---------------------------------------------------------------------")

	// Wait for user input before returning to list
	fmt.Print("\nPress Enter to return to bucket list...")
	fmt.Scanln()
}
