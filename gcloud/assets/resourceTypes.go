package assets

import (
	"strconv"
	"strings"
)

type RegionZone struct {
	Region string
	Zone   string
}

type AcceleratorType struct {
	Name  string
	Count int64
}

type MachineType struct {
	Name         string
	CpuCount     uint32
	MemoryMb     uint64
	IsSharedCpu  bool
	Accelerators []AcceleratorType
}

type DiskType struct {
	Name   string
	Region string // if this is set, it's a regional disk
	// TODO: what about zone for zonal disks?
	DefaultSizeGb int64
}

type ResourceMetadata struct {
	Mt MachineType
	Dt DiskType
}

// Max bandwith for a given machine type, in Gb per second.
func MaxBandwidthGbps(m string, ingress bool, external bool) int32 {
	// external ingress is capped at 1.8 million packets/second or 20Gbps
	// internal ingress is not charged (the egress point gets charged).
	if ingress {
		return 20
	}
	// external egress is capped at 7Gbps per VM
	if external {
		return 7
	}
	// Internal egress bandwidth depends on the machine type. I haven't found
	// this information in the machine type list you can get from the compute
	// API though, so this is from the website. Should maybe put it into the cache db.
	parts := strings.Split(m, "-")
	cpuCount, _ := strconv.Atoi(parts[2])
	if parts[0] == "n1" {
		if cpuCount == 1 {
			return 2
		}
		if cpuCount <= 4 {
			return 10
		}
		if cpuCount <= 8 {
			return 16
		}
		return 32 // it's only 16 if using a cpu before skylake actually
	}
	if parts[0] == "n2" || parts[0] == "n2d" {
		if cpuCount <= 4 {
			return 10
		}
		if cpuCount <= 8 {
			return 16
		}
		return 32
	}
	max := 2 * cpuCount
	if max > 16 {
		return 16
	}
	return int32(max)
}
