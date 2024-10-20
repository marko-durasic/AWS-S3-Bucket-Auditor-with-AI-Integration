#!/bin/bash

# Variables
BUCKET_NAME_1="non-sensitive-data-bucket"
BUCKET_NAME_2="sens-data-bucket"
REGION="us-east-1"
TEMP_DIR=temp
# Function to check if a bucket exists
bucket_exists() {
    local BUCKET_NAME=$1
    aws s3 ls "s3://$BUCKET_NAME" --region $REGION > /dev/null 2>&1
    return $?
}

# Function to delete bucket contents and handle access errors
delete_bucket_contents() {
    local BUCKET_NAME=$1
    echo "Checking if bucket $BUCKET_NAME exists..."

    if bucket_exists $BUCKET_NAME; then
        echo "Removing objects from bucket: $BUCKET_NAME..."
        aws s3 rm s3://$BUCKET_NAME --recursive --region $REGION
        if [ $? -ne 0 ]; then
            echo "Error: Failed to delete objects from $BUCKET_NAME. Skipping object removal."
        fi

        echo "Deleting bucket: $BUCKET_NAME..."
        aws s3 rb s3://$BUCKET_NAME --force --region $REGION
        if [ $? -ne 0 ]; then
            echo "Error: Failed to delete bucket $BUCKET_NAME. You may need to manually resolve this."
        fi
    else
        echo "Bucket $BUCKET_NAME does not exist. Skipping deletion."
    fi
}

# Step 1: Cleanup both buckets
delete_bucket_contents $BUCKET_NAME_1
delete_bucket_contents $BUCKET_NAME_2

# Step 2: Remove local test files if they still exist
# clean up in the same folder as the script, not where the script is executed from
cd $(dirname $0)

echo "remove temp folder..."
rm -rf $TEMP_DIR

echo "Cleanup complete!"
