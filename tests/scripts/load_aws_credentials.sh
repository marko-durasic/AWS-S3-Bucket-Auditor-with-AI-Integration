#!/bin/bash

# Load AWS credentials from ~/.aws/credentials if AWS_ACCESS_KEY_ID is not set
if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    # Get credentials from AWS CLI config
    export AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id --profile ${AWS_PROFILE:-default})
    export AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key --profile ${AWS_PROFILE:-default})
fi 