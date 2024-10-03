#!/bin/bash

REGION="us-east-1"  # Specify your desired region here
BUCKET_NAME_1="non-sensitive-data-bucket"
BUCKET_NAME_2="sens-data-bucket"
TEST_FILE_1="non-sensitive.txt"
TEST_FILE_2="sensitive.txt"

# Function to create a bucket and upload data
create_bucket_and_upload_data() {
    local BUCKET_NAME=$1
    local TEST_FILE=$2
    local FILE_CONTENT=$3

    echo "Creating bucket $BUCKET_NAME in region $REGION..."
    if [ "$REGION" == "us-east-1" ]; then
        aws s3 mb s3://$BUCKET_NAME
    else
        aws s3 mb s3://$BUCKET_NAME --region $REGION
    fi

    echo "Uploading file to $BUCKET_NAME..."
    echo "$FILE_CONTENT" > $TEST_FILE
    aws s3 cp $TEST_FILE s3://$BUCKET_NAME/ --region $REGION
}

# Step 1: Create non-sensitive bucket and upload non-sensitive data
create_bucket_and_upload_data $BUCKET_NAME_1 $TEST_FILE_1 "This is a non-sensitive file."

# Step 2: Create sensitive bucket and upload sensitive data
create_bucket_and_upload_data $BUCKET_NAME_2 $TEST_FILE_2 "American Express
5135725008183484 09/26
CVE: 550

American Express
347965534580275 05/24
CCV: 4758

Mastercard
5105105105105100
Exp: 01/27
Security code: 912

"

echo "Running the S3 Bucket Auditor tool..."
go build -o s3auditor main.go || { echo "Go build failed!"; exit 1; }
./s3auditor || { echo "S3 Auditor tool failed!"; exit 1; }

# Step 4: Clean up the test S3 buckets and files
echo "Cleaning up test S3 buckets and files..."

aws s3 rm s3://$BUCKET_NAME_1 --recursive --region $REGION || echo "Failed to remove objects from $BUCKET_NAME_1."
aws s3 rm s3://$BUCKET_NAME_2 --recursive --region $REGION || echo "Failed to remove objects from $BUCKET_NAME_2."

aws s3 rb s3://$BUCKET_NAME_1 --force --region $REGION || echo "Failed to delete $BUCKET_NAME_1."
aws s3 rb s3://$BUCKET_NAME_2 --force --region $REGION || echo "Failed to delete $BUCKET_NAME_2."

# Step 5: Remove local test files
rm -f $TEST_FILE_1 $TEST_FILE_2

echo "Test complete!"