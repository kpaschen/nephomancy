package resources

func MakeSampleProject() Project {
	ret := Project{
		Name:         "Nephomancy sample project",
		InstanceSets: []*InstanceSet{makeSampleInstanceSet()},
		DiskSets:     []*DiskSet{makeSampleDiskSet()},
		Images:       []*Image{makeSampleImage()},
		Networks:     []*Network{makeSampleNetwork()},
	}
	return ret
}

func sampleLocation() *Location {
	// This just happens to be the most common location.
	return &Location{
		GlobalRegion: "NAM",
		Continent:    "NorthAmerica",
		CountryCode:  "US",
	}
}

func makeSampleInstanceSet() *InstanceSet {
	mt := MachineType{
		CpuCount: 2,
		MemoryGb: 16,
	}
	ret := &InstanceSet{
		Name: "Sample InstanceSet",
		Template: &Instance{
			Location: sampleLocation(),
			Type:     &mt,
		},
		Count:              1,
		UsageHoursPerMonth: 720,
	}
	return ret
}

func makeSampleDiskSet() *DiskSet {
	ret := &DiskSet{
		Name: "Sample Disk Set",
		Template: &Disk{
			Location: sampleLocation(),
			Type: &DiskType{
				SizeGb:   100,
				DiskTech: "SSD",
			},
			ActualSizeGb: 100,
		},
		Count:          1,
		PercentUsedAvg: 70,
	}
	return ret
}

func makeSampleImage() *Image {
	ret := &Image{
		Name: "Sample Image",
		Location: sampleLocation(),
		SizeGb:   10,
	}
	return ret
}

func makeSampleNetwork() *Network {
	snw := &Subnetwork{
		Name:                        "default subnetwork",
		Location:                    sampleLocation(),
		IngressGbitsPerMonth:        1,
		ExternalEgressGbitsPerMonth: 1,
		InternalEgressGbitsPerMonth: 3,
	}
	ret := &Network{
		Name:        "default network",
		Subnetworks: []*Subnetwork{snw},
	}
	return ret
}
