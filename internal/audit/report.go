package audit

import (
	"time"

	"github.com/fatih/color"
	"github.com/marko-durasic/aws-s3-bucket-auditor/internal/models"
)

func PrintBucketReport(info models.BucketInfo) {
	color.Cyan("\nS3 Bucket Security Audit Report:")
	color.Cyan("=====================================================================")
	color.Green("Bucket Name      : %s", info.Name)
	color.Cyan("Region           : %s", info.Region)
	color.Yellow("Public Access    : %t", info.IsPublic)
	color.Cyan("Encryption       : %s", info.Encryption)
	color.Cyan("Versioning       : %s", info.VersioningStatus)
	if info.SensitiveData {
		color.Red("Sensitive Data   : %t", info.SensitiveData)
	} else {
		color.Green("Sensitive Data   : %t", info.SensitiveData)
	}
	color.Cyan("Audit Duration   : %s", info.AuditDuration.Round(time.Second))
	color.Cyan("---------------------------------------------------------------------")
}
