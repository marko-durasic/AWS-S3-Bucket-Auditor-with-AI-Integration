package models

import "time"

type BucketBasicInfo struct {
	Name   string
	Region string
}

type BucketInfo struct {
	Name             string
	Region           string
	IsPublic         bool
	Encryption       string
	VersioningStatus string
	SensitiveData    bool
	AuditDuration    time.Duration
}
