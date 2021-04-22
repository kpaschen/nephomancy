package resources

import ()

type InstanceType struct {
	Name      string
	MemoryMiB uint64
	// spec string can be "25 Gigabit", "Very Low", "Up to 10 Gigabit", "Moderate",
	// "Low to Moderate"
	// I think "Up to" can be burst capability, it won't be sustained.
	// Based on very random research, maybe "Low" ~ 1Gigabit, "Moderate" ~ 5
	// and "High" ~ 10.
	// For "up to" you probably want to divide at least 5.
	NetworkPerformanceGbit   uint32
	SupportedUsageClasses    []string // "on-demand", "spot"
	DefaultCpuCount          uint32
	InstanceStorageSupported bool
	// This is only set when InstanceStorageSupported is true.
	InstanceStorageMaxSizeGb uint64
	// This tells you whether the instance storage is ssd or hdd.
	InstanceStorageType string

	ValidCores []uint32
	GpuCount   uint32
}

// Note that for charging, have to look at both the disk space and
// the iops or throughput charges. Those charges use the same
// Volume API Name (e.g. "io1") as the disk space but the product
// family for the sku is something like "Provisioned Throughput"
// or "System Operation" and the "Group" is something like "EBS
// Throughput".
type VolumeType struct {
	Name                       string
	VolumeApiType              string
	Media                      string
	MaxVolumeSizeGiB           uint32
	MaxIOPSPerVolumeKiB        uint32
	MaxThroughputPerVolumeMiBs uint32
	MultiAttach                bool
}

func StandardVolumeTypes() []VolumeType {
	return []VolumeType{
		VolumeType{
			Name:                       "Provisioned IOPS",
			VolumeApiType:              "io1",
			Media:                      "SSD-backed",
			MaxVolumeSizeGiB:           uint32(16384),
			MaxIOPSPerVolumeKiB:        uint32(64000),
			MaxThroughputPerVolumeMiBs: uint32(1000),
			MultiAttach:                true,
		},
		VolumeType{
			Name:                       "Provisioned IOPS",
			VolumeApiType:              "io2",
			Media:                      "SSD-backed",
			MaxVolumeSizeGiB:           uint32(16384),
			MaxIOPSPerVolumeKiB:        uint32(64000),
			MaxThroughputPerVolumeMiBs: uint32(1000),
			MultiAttach:                true,
		},
		VolumeType{
			Name:                       "General Purpose",
			VolumeApiType:              "io2",
			Media:                      "SSD-backed",
			MaxVolumeSizeGiB:           uint32(16384),
			MaxIOPSPerVolumeKiB:        uint32(16000),
			MaxThroughputPerVolumeMiBs: uint32(250),
			MultiAttach:                false,
		},
		VolumeType{
			Name:                       "Throughput Optimized HDD",
			VolumeApiType:              "st1",
			Media:                      "HDD-backed",
			MaxVolumeSizeGiB:           uint32(16384),
			MaxIOPSPerVolumeKiB:        uint32(500),
			MaxThroughputPerVolumeMiBs: uint32(500),
			MultiAttach:                false,
		},
		VolumeType{
			Name:                       "Cold HDD",
			VolumeApiType:              "sc1",
			Media:                      "HDD-backed",
			MaxVolumeSizeGiB:           uint32(16384),
			MaxIOPSPerVolumeKiB:        uint32(250),
			MaxThroughputPerVolumeMiBs: uint32(250),
			MultiAttach:                false,
		},
	}
}
