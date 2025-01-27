package testutils

import (
	"fmt"
	"strings"
	"time"
)

// GetTestBucketName generates a valid S3 bucket name for testing
func GetTestBucketName(prefix string) string {
	// Sanitize prefix: lowercase, replace invalid chars with dashes
	sanitized := strings.ToLower(prefix)
	sanitized = strings.ReplaceAll(sanitized, " ", "-")

	// Remove any double dashes and trim dashes from ends
	sanitized = strings.ReplaceAll(sanitized, "--", "-")
	sanitized = strings.Trim(sanitized, "-")

	// Add timestamp for uniqueness
	timestamp := time.Now().Format("20060102150405")

	name := fmt.Sprintf("%s-%s", sanitized, timestamp)
	if len(name) > 63 {
		name = name[len(name)-63:]
	}
	return name
}
