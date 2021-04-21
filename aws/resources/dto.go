package resources

import (
)

type InstanceType struct {
	Name string
	MemoryMiB uint64
	// spec string can be "25 Gigabit", "Very Low", "Up to 10 Gigabit", "Moderate",
	// "Low to Moderate"
	// I think "Up to" can be burst capability, it won't be sustained.
	// Based on very random research, maybe "Low" ~ 1Gigabit, "Moderate" ~ 5
	// and "High" ~ 10.
	// For "up to" you probably want to divide at least 5.
	NetworkPerformanceGbit uint32
	SupportedUsageClasses []string  // "on-demand", "spot"
	DefaultCpuCount uint32
	InstanceStorageSupported bool
	// This is only set when InstanceStorageSupported is true.
	InstanceStorageMaxSizeGb uint64

	ValidCores []uint32
	GpuCount uint32
}
