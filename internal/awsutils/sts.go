package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// STSClientAPI defines the interface for STS operations we use
type STSClientAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}
