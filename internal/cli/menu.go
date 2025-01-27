package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/manifoldco/promptui"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/audit"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/models"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/ui"
)

func HandleBucketAudit(cfg aws.Config, s3Client *s3.Client, macieClient *macie2.Client) {
	bucketName, err := PromptForBucketSelection(s3Client)
	if err == promptui.ErrInterrupt {
		return
	}
	if err != nil {
		ui.ShowError("Error selecting bucket: %v", err)
		log.Printf("Error selecting bucket: %v", err)
		return
	}

	stsClient := sts.NewFromConfig(cfg)
	scanner := audit.NewScanner(cfg, s3Client, macieClient, stsClient)
	scanner.AuditBucket(bucketName)
}

func PromptForBucketSelection(s3Client *s3.Client) (string, error) {
	buckets, err := awsutils.ListBuckets(s3Client)
	if err != nil {
		return "", fmt.Errorf("unable to list buckets: %w", err)
	}

	items := append(buckets, models.BucketBasicInfo{Name: "[ Exit ]", Region: ""})

	prompt := &promptui.Select{
		Label: "Choose a bucket to audit (Ctrl+C or Exit option to return)",
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
		return "", promptui.ErrInterrupt
	}
	if err != nil {
		return "", err
	}

	if idx == len(buckets) {
		return "", promptui.ErrInterrupt
	}

	return buckets[idx].Name, nil
}

func PromptMainMenu() (string, error) {
	actions := []string{"List S3 Buckets", "Audit a Bucket", "Exit"}
	prompt := &promptui.Select{
		Label: "What would you like to do? (Ctrl+C or Exit option to exit)",
		Items: actions,
	}

	_, result, err := prompt.Run()
	if err == promptui.ErrInterrupt || err == promptui.ErrAbort || err == promptui.ErrEOF {
		return "Exit", nil
	}
	if err != nil {
		return "", fmt.Errorf("menu selection failed: %w", err)
	}
	return result, nil
}
