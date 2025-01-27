package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/macie2"
)

// MacieClientAPI defines the interface for Macie operations we use
type MacieClientAPI interface {
	CreateClassificationJob(ctx context.Context, params *macie2.CreateClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.CreateClassificationJobOutput, error)
	DescribeClassificationJob(ctx context.Context, params *macie2.DescribeClassificationJobInput, optFns ...func(*macie2.Options)) (*macie2.DescribeClassificationJobOutput, error)
	ListFindings(ctx context.Context, params *macie2.ListFindingsInput, optFns ...func(*macie2.Options)) (*macie2.ListFindingsOutput, error)
	GetFindings(ctx context.Context, params *macie2.GetFindingsInput, optFns ...func(*macie2.Options)) (*macie2.GetFindingsOutput, error)
}
