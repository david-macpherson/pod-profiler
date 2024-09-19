//go:build !coreweave && !aws && !azure && !gcp
// +build !coreweave,!aws,!azure,!gcp

package cloud

const PLATFORM CloudPlatformType = CloudPlatformType_Local

var VALID_REGIONS = []string{"ORD1", "LGA1", "LAS1"}
var VALID_CPU_TYPES = []string{"intel-xeon-v3", "intel-xeon-v4", "intel-xeon-scalable", "amd-epyc-rome", "amd-epyc-milan"}
var VALID_GPU_TYPES = []string{"A40", "RTX_A6000", "RTX_A5000", "RTX_A4000", "Quadro_RTX_5000", "Quadro_RTX_4000"}
