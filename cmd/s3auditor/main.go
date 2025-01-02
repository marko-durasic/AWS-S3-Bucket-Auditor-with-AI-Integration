package main

import (
	"context"
	"fmt"
	"log"

	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/awsutils"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/cli"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/logger"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/ui"
)

func main() {
	if err := logger.InitLogger("s3_audit.log"); err != nil {
		fmt.Println(err)
		return
	}

	ui.ShowWelcomeScreen()

	clients, err := awsutils.NewAWSClients(context.Background())
	if err != nil {
		ui.ShowError("Unable to initialize AWS clients: %v", err)
		log.Printf("Error: Unable to initialize AWS clients: %v", err)
		return
	}

	for {
		result, err := cli.PromptMainMenu()
		if err != nil {
			ui.ShowError("Menu error: %v", err)
			log.Printf("Menu error: %v", err)
			continue
		}

		switch result {
		case "List S3 Buckets":
			buckets, err := awsutils.ListBuckets(clients.S3Client)
			if err != nil {
				ui.ShowError("Error listing buckets: %v", err)
				log.Printf("Error listing buckets: %v", err)
				continue
			}
			cli.DisplayBucketsList(clients.S3Client, buckets)
		case "Audit a Bucket":
			cli.HandleBucketAudit(clients.Config, clients.S3Client, clients.MacieClient)
		case "Exit":
			ui.ShowSuccess("Goodbye! Stay secure.")
			return
		}
	}
}
