# Codebase Overview

This document provides a brief tour of the repository for newcomers. It explains the main directories and their purpose, along with tips on where to look next.

## General Structure

- **`cmd/s3auditor`** – Entry point containing `main.go`. It sets up logging, initializes AWS clients, displays the menu and routes user selections.
- **`internal/awsutils`** – Helper functions and interfaces for interacting with AWS services such as S3 and Macie.
- **`internal/audit`** – Houses the `Scanner` type, which orchestrates the audit workflow and formats the final report.
- **`internal/cli`** – Provides the interactive command-line interface.
- **`internal/ui`** – Presents the welcome banner and helpers for colored output.
- **`internal/config`** – Handles configuration like the Macie job timeout via environment variables.
- **`internal/logger`** – Initializes file-based logging.
- **`test` and `tests`** – Unit tests and integration tests, including helper scripts under `tests/scripts`.
- **`docs`** – Project documentation, screenshots and TODO items.

## Key Concepts

- The tool uses the AWS SDK for Go. Make sure your AWS credentials are configured.
- Macie classification jobs are asynchronous; the scanner polls until a job finishes.
- Set `MACIE_JOB_TIMEOUT_MINUTES` to adjust how long the scanner waits for Macie.

## Next Steps

- Check `docs/TODO.md` for planned features like Docker support and additional CLI options.
- Explore test utilities in `internal/testutils` to see how test buckets are created and cleaned up.
- Follow existing patterns when extending functionality: add AWS helpers under `internal/awsutils`, enhance `Scanner` logic, and expose new CLI options in `internal/cli`.

This overview should help you get up to speed quickly and identify areas for further exploration.
