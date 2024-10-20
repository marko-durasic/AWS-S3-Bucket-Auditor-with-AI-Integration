# TODO

## AWS S3 Bucket Auditor with Macie Integration


### Steps to Continue
- Refactor main.go, so it is not so messy, more modular
- can just have one command and then you enter into a cli "dashboard" where you can select the different options. make it look user friendly
- move stuff that are kind of hardcoded to environment variables like 40 minutes timeout and name of the bucket we use for macie findings
- when choosing audit a bucket, we should have option to audit few of them, not just one, and to search by letter, what if there is many buckets
- add check if the bucket logging is enabled
- acl and or policy is public.. see if you are already doing that
- so I have few checks, I should be able to select which ones I want to do (especially because Marcie o)
- dockerize it
- not everything in on main.go, split it into different files
- also besides saying it has sensitive data, we should give more info, maybe able to enter into each result
1. **Verify the S3 Bucket Auditor**:
   - Add more test cases for the auditor to handle different bucket configurations (e.g., public vs. private buckets, buckets with different encryption types).
   - Refactor the Go tool to improve error handling and logging, especially for longer-running Macie jobs.

2. **Explore Macie Scoping and Criteria**:
   - Investigate adding scoping criteria to focus on specific files or data types for the Macie job (e.g., specific prefixes, file extensions).
   - Experiment with `managedDataIdentifierSelector` to fine-tune sensitive data detection.

3. **Additional Features**:
   - Add command-line options to the tool for specifying regions, bucket names, and scan parameters.
   - Implement scanning for bucket policies and access logs for a more comprehensive audit.
   - Add support for concurrent bucket processing with better timeout management to avoid infinite loops on long-running jobs.

4. **Cloud Cost Management**:
   - Review AWS services used (Macie, S3, etc.) and set up billing alerts in the AWS console to monitor costs.
   - Set up automatic cleanup mechanisms in your Go tool to avoid orphaned resources (e.g., deleting test buckets after the audit).

5. **Testing**:
   - Develop a set of unit tests to validate the different functions in the Go tool, including mocked Macie and S3 responses.
   - Create a test bucket with a mixture of sensitive and non-sensitive files for end-to-end testing.

6. **Next Steps for AWS Usage**:
   - **Access Management**: Create IAM roles and policies tailored for your S3 and Macie usage, minimizing permissions to only what's necessary.
   - **Further Exploration**: Explore additional AWS services that can be integrated into the tool, such as AWS Config for configuration compliance.

7. **Documentation**:
   - Improve `README.md` to include detailed usage instructions, prerequisites, and AWS setup guidelines.
   - Add an FAQ section to the documentation to address common issues (e.g., long-running Macie jobs, insufficient permissions).

8. **Security Enhancements**:
   - Implement secure handling of AWS credentials using environment variables or the AWS Secrets Manager.
   - Log sensitive actions and generate an audit trail for review.

### Notes
- **Current Job Status**: Confirmed that the Macie jobs were successfully canceled, and test buckets were removed.
- **Bucket Management**: Test buckets used (`non-sensitive-data-bucket`) were successfully deleted.

### Commands for Reference
- **List Macie Jobs**: `aws macie2 list-classification-jobs --query 'items[*].[jobId, jobStatus]' --output table --region <your-region>`
- **Cancel Macie Job**: `aws macie2 update-classification-job --job-id <job-id> --job-status CANCELLED --region <your-region>`
- **Delete S3 Bucket**: `aws s3 rb s3://<bucket-name> --force --region <your-region>`
- **Create S3 Bucket**: `aws s3 mb s3://<bucket-name> --region <your-region>`

---

_Last updated on: 2024-09-15_
