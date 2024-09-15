for job_id in $(aws macie2 list-classification-jobs --query 'items[?jobStatus==`RUNNING`].jobId' --output text --region us-east-1); do
    aws macie2 update-classification-job --job-id $job_id --job-status CANCELLED --region us-east-1
    echo "Cancelled job: $job_id"
done
