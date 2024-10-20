package models

// BucketInfo holds information about an S3 bucket
type BucketInfo struct {
	Name             string
	Region           string
	IsPublic         bool
	Encryption       string
	VersioningStatus string
	SensitiveData    bool
}
