package cache

import (
	"nephomancy/aws/resources"
	common "nephomancy/common/resources"
	"testing"
)

func TestCheckInstanceTypeAgainstVmSpec(t *testing.T) {
	var it resources.InstanceType
	var spec common.Instance

	it = resources.InstanceType{
		MemoryMiB:       1000,
		GpuCount:        0,
		DefaultCpuCount: 1,
	}
	spec = common.Instance{
		Type: &common.MachineType{
			MemoryGb: 1,
			GpuCount: 0,
			CpuCount: 1,
		},
		LocalStorage: []*common.Disk{},
	}

	if err := checkInstanceTypeAgainstVmSpec(it, spec); err != nil {
		t.Errorf("instance check failed: %v", err)
	}

	it = resources.InstanceType{
		MemoryMiB:       1000,
		GpuCount:        0,
		DefaultCpuCount: 1,
	}
	spec = common.Instance{
		Type: &common.MachineType{
			MemoryGb: 2,
			GpuCount: 0,
			CpuCount: 1,
		},
		LocalStorage: []*common.Disk{},
	}

	if err := checkInstanceTypeAgainstVmSpec(it, spec); err == nil {
		t.Errorf("instance check should have failed because memory too low")
	}

	it = resources.InstanceType{
		MemoryMiB:       1000,
		GpuCount:        0,
		DefaultCpuCount: 1,
	}
	spec = common.Instance{
		Type: &common.MachineType{
			MemoryGb: 1,
			GpuCount: 0,
			CpuCount: 2,
		},
		LocalStorage: []*common.Disk{},
	}
	if err := checkInstanceTypeAgainstVmSpec(it, spec); err == nil {
		t.Errorf("instance check should have failed because cpu count too low")
	}

	it = resources.InstanceType{
		MemoryMiB:       1000,
		GpuCount:        0,
		DefaultCpuCount: 1,
		ValidCores:      []uint32{1, 2},
	}
	spec = common.Instance{
		Type: &common.MachineType{
			MemoryGb: 1,
			GpuCount: 0,
			CpuCount: 2,
		},
		LocalStorage: []*common.Disk{},
	}
	if err := checkInstanceTypeAgainstVmSpec(it, spec); err != nil {
		t.Errorf("instance check failed: %v", err)
	}

	it = resources.InstanceType{
		MemoryMiB:       1000,
		GpuCount:        0,
		DefaultCpuCount: 1,
	}
	ls := make([]*common.Disk, 2)
	ls[0] = &common.Disk{
		Type: &common.DiskType{
			SizeGb:   1,
			DiskTech: "SSD",
		},
	}
	ls[1] = &common.Disk{
		Type: &common.DiskType{
			SizeGb:   2,
			DiskTech: "Standard",
		},
	}
	spec = common.Instance{
		Type: &common.MachineType{
			MemoryGb: 1,
			GpuCount: 0,
			CpuCount: 1,
		},
		LocalStorage: ls,
	}

	if err := checkInstanceTypeAgainstVmSpec(it, spec); err == nil {
		t.Errorf("instance check should have failed because of local storage")
	}

	it = resources.InstanceType{
		MemoryMiB:                1000,
		GpuCount:                 0,
		DefaultCpuCount:          1,
		InstanceStorageSupported: true,
		InstanceStorageMaxSizeGb: 3,
		InstanceStorageType:      "ssd",
	}
	if err := checkInstanceTypeAgainstVmSpec(it, spec); err != nil {
		t.Errorf("instance check failed: %v", err)
	}
}
