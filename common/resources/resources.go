package resources

import (
	"fmt"
)

func PrintLocation(l Location) string {
	if l.CountryCode != "" {
		if l.Continent != "" {
			return fmt.Sprintf("%s, %s", l.CountryCode, l.Continent)
		} else if l.GlobalRegion != "" {
			return fmt.Sprintf("%s, %s", l.CountryCode, l.GlobalRegion)
		} else {
			return l.CountryCode
		}
	}
	if l.Continent != "" {
		return l.Continent
	}
	if l.GlobalRegion != "" {
		return l.GlobalRegion
	}
	return "Probably somewhere on Earth"
}

func PrintMachineType(m MachineType) string {
	if m.GpuCount > 0 {
		return fmt.Sprintf("%d gpus, %d cpus, % gb memory",
		m.GpuCount, m.CpuCount, m.MemoryGb)
	} else {
		return fmt.Sprintf("%d cpus, %d gb memory",
		m.CpuCount, m.MemoryGb)
	}
}

func PrintDiskType(d DiskType) string {
	if d.DiskTech == "SSD" {
		return fmt.Sprintf("%d GB of SSD", d.SizeGb)
	} else {
		return fmt.Sprintf("%d GB of non-SSD", d.SizeGb)
	}
}
