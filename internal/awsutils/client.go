package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSClients struct {
	Config      aws.Config
	S3Client    *s3.Client
	MacieClient *macie2.Client
}

func NewAWSClients(ctx context.Context) (*AWSClients, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &AWSClients{
		Config:      cfg,
		S3Client:    s3.NewFromConfig(cfg),
		MacieClient: macie2.NewFromConfig(cfg),
	}, nil
}
