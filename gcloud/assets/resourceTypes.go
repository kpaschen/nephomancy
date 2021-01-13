package assets

type RegionZone struct {
	Region string
	Zone string
}

type MachineType struct {
	Name string
	CpuCount int64
	MemoryMb int64
	IsSharedCpu bool
}

type DiskType struct {
	Name string
	Region string  // if this is set, it's a regional disk
	DefaultSizeGb int64
}
