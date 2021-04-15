package resources

import (
	"log"
	"nephomancy/common/geo"
)

func GetProviderNames(p Project) []string {
	providers := make(map[string]bool)
	for _, is := range p.InstanceSets {
		if is.Template != nil {
			if is.Template.ProviderDetails != nil {
				for pname, _ := range is.Template.ProviderDetails {
					providers[pname] = true
				}
			}
		}
	}
	ret := make([]string, len(providers))
	idx := 0
	for p, _ := range providers {
		ret[idx] = p
		idx++
	}
	return ret
}

func ResolveLocation(where string) *Location {
	region := geo.RegionFromString(where)
	if region != geo.UnknownG {
		return &Location{
			GlobalRegion: where,
		}
	}
	continent := geo.ContinentFromString(where)
	if continent != geo.UnknownC {
		return &Location{
			Continent: where,
		}
	}
	continent, region = geo.GetContinent(where)
	if continent != geo.UnknownC && region != geo.UnknownG {
		return &Location{
			GlobalRegion: region.String(),
			Continent:    continent.String(),
			CountryCode:  where,
		}
	}
	return nil
}

func MakeSampleProject(where string) Project {
	var loc *Location
	if where == "" {
		loc = sampleLocation()
	} else {
		loc = ResolveLocation(where)
		if loc == nil {
			log.Printf("Failed to resolve %s into a location, will be using default location.\n", where)
			loc = sampleLocation()
		}
	}
	ret := Project{
		Name:         "Nephomancy sample project",
		InstanceSets: []*InstanceSet{makeSampleInstanceSet(loc)},
		DiskSets:     []*DiskSet{makeSampleDiskSet(loc)},
		Networks:     []*Network{makeSampleNetwork(loc)},
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

func makeSampleInstanceSet(loc *Location) *InstanceSet {
	mt := MachineType{
		CpuCount: 2,
		MemoryGb: 16,
	}
	ret := &InstanceSet{
		Name: "Sample InstanceSet",
		Template: &Instance{
			Location: loc,
			Type:     &mt,
			Os:       "linux",
		},
		Count:              1,
		UsageHoursPerMonth: 720,
	}
	return ret
}

func makeSampleDiskSet(loc *Location) *DiskSet {
	ret := &DiskSet{
		Name: "Sample Disk Set",
		Template: &Disk{
			Location: loc,
			Type: &DiskType{
				SizeGb:   100,
				DiskTech: "SSD",
			},
		},
		Count:              1,
		UsageHoursPerMonth: 720,
	}
	return ret
}

func makeSampleNetwork(loc *Location) *Network {
	snw := &Subnetwork{
		Name: "default subnetwork",
		Location: loc,
		IpAddresses:    1,
		BandwidthMbits: 150,
		Gateways:       []*Gateway{&Gateway{}},
		IngressGbitsPerMonth:        1,
		ExternalEgressGbitsPerMonth: 1,
		InternalEgressGbitsPerMonth: 3,
	}
	ret := &Network{
		Name:           "default network",
		Subnetworks: []*Subnetwork{snw},
	}
	return ret
}
