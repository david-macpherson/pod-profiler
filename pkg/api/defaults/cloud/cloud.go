package cloud

type CloudPlatformType string

const (
	CloudPlatformType_AWS       CloudPlatformType = "aws"
	CloudPlatformType_CoreWeave CloudPlatformType = "coreweave"
	CloudPlatformType_Azure     CloudPlatformType = "azure"
	CloudPlatformType_GCP       CloudPlatformType = "gcp"
	CloudPlatformType_Local     CloudPlatformType = "local"
)
