#!/bin/bash

# Load AWS credentials from ~/.aws/credentials
export AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id --profile ${AWS_PROFILE:-default})
export AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key --profile ${AWS_PROFILE:-default})
export AWS_REGION=$(aws configure get region --profile ${AWS_PROFILE:-default})

# Run the integration tests with increased timeout
go test ./test/integration/... -v -timeout 30m 