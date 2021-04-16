package cache

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"nephomancy/common/geo"
	common "nephomancy/common/resources"
	"nephomancy/gcloud/assets"
	"strings"
)

// Returns nil if the gcloud vm meets the spec, an error otherwise.
func checkVmSpec(db *sql.DB, gvm assets.GCloudVM, spec common.Instance) error {
	if l := spec.Location; l != nil {
		if err := checkLocation(gvm.Region, *l); err != nil {
			return err
		}
	}
	mt, err := GetMachineType(db, gvm.MachineType, gvm.Region)
	if err != nil {
		return err
	}
	if t := spec.Type; t != nil {
		if t.CpuCount > mt.CpuCount ||
			t.MemoryGb > uint32(mt.MemoryMb/1000) {
			return fmt.Errorf("%s vm provider details (%+v) do not match spec (%+v)",
				assets.GcloudProvider, mt, t)
		}
	}
	if os := spec.Os; os != "" {
		gOs := gvm.OsChoice
		if gOs == "" {
			return fmt.Errorf("%s vm provider details do not contain os. Please choose one that matches the spec os %s.",
				assets.GcloudProvider, os)
		}
		if err := assets.DoesOsMatch(os, gOs); err != nil {
			return err
		}
	}
	// TODO: also check local disk specs. If spec.LocalStorage is not nil, need to see
	// whether mt supports local disk.
	// E2, memory-optimized ultramem, e2 shared-core and n1 shared-core do not support local ssd.
	// n2, n2d, n1, compute-optimized, memory-optimized megamem, accelerator-optimized high-gpu,
	// accelerator-optimized mega-gpu support local ssd.
	return nil
}

// Returns nil if the gcloud disk meets the spec, an error otherwise.
func checkDiskSpec(db *sql.DB, dsk assets.GCloudDisk, spec common.Disk) error {
	region := ""
	if dsk.IsRegional {
		region = dsk.Region
	}
	if l := spec.Location; l != nil {
		if err := checkLocation(dsk.Region, *l); err != nil {
			return err
		}
	}
	dt, err := getDiskType(db, dsk.DiskType, region)
	if err != nil {
		return err
	}
	if t := spec.Type; t != nil {
		if t.SizeGb > uint32(dt.DefaultSizeGb) {
			return fmt.Errorf("%s disk provider details (%+v) do not match spec (%+v)",
				assets.GcloudProvider, dt, t)
		}
	}
	// Assume for now that it doesn't matter whether there is an image attached
	// to the disk.
	return nil
}

func getOsBySpec(spec string) string {
	maybe := assets.OsChoiceByName(spec)
	if maybe != assets.UnspecifiedOs {
		return maybe.String()
	}
	s := strings.ToLower(spec)
	if s == "linux" {
		return assets.Ubuntu.String()
	}
	if s == "windows" {
		return assets.WindowsServer.String()
	}
	return ""
}

// Populates provider-specific details from spec if they are empty.
// If they are not empty, they will be left as they were, but the tool
// will log a warning message.
func FillInProviderDetails(db *sql.DB, p *common.Project) error {
	// These are locations that have been resolved into zones or regions.
	// This is so that if an instance set has the same location spec as
	// a disk set, they both end up in the same region.
	// It might be nice for this to be configurable in the spec, and it might
	// also be nice for it to use zones where possible, I just haven't done it yet.
	locations := make(map[string]string)
	for _, vmset := range p.InstanceSets {
		if vmset.Template.Location == nil {
			return fmt.Errorf("missing vmset location information")
		}
		if vmset.Template.Type == nil {
			return fmt.Errorf("missing vmset type information")
		}
		if vmset.Template.Os == "" {
			return fmt.Errorf("missing vmset os information")
		}
		if vmset.Template.ProviderDetails == nil {
			vmset.Template.ProviderDetails = make(map[string](*anypb.Any))
		}
		// Special case: there already are provider details.
		// Make sure they are consistent, otherwise print a warning.
		if vmset.Template.ProviderDetails[assets.GcloudProvider] != nil {
			var gvm assets.GCloudVM
			err := ptypes.UnmarshalAny(vmset.Template.ProviderDetails[assets.GcloudProvider], &gvm)
			if err != nil {
				return err
			}
			if err = checkVmSpec(db, gvm, *vmset.Template); err != nil {
				return err
			}
			// If checkVmSpec hasn't errored, gvm.Region exists.
			locstring := common.PrintLocation(*vmset.Template.Location)
			if locations[locstring] == "" {
				locations[locstring] = gvm.Region
			}
			log.Printf("Instance Set %s already has details for provider %s, leaving them as they are.\n",
				vmset.Name, assets.GcloudProvider)
		} else { // There are no provider details
			regions := resolveSpecLocation(*vmset.Template.Location, "")
			if len(regions) == 0 {
				return fmt.Errorf(
					"provider %s does not support regions matching location %v",
					assets.GcloudProvider, vmset.Template.Location)
			}
			mt, r, err := getMachineTypeBySpec(db, *vmset.Template.Type, regions)
			if err != nil {
				return err
			}
			// TODO: are all OS available for all machine types?
			os := getOsBySpec(vmset.Template.Os)
			if os == "" {
				return fmt.Errorf("provider %s does not support an os matching %s",
					assets.GcloudProvider, vmset.Template.Os)
			}
			details, err := ptypes.MarshalAny(&assets.GCloudVM{
				MachineType: mt,
				Region:      r[0], // only using first region
				OsChoice:    os,
			})
			if err != nil {
				return err
			}
			locstring := common.PrintLocation(*vmset.Template.Location)
			if locations[locstring] == "" {
				locations[locstring] = r[0]
			}
			vmset.Template.ProviderDetails[assets.GcloudProvider] = details
		}
	}
	for _, nw := range p.Networks {
		if nw.ProviderDetails == nil {
			nw.ProviderDetails = make(map[string](*anypb.Any))
		}
		var gnw assets.GCloudNetwork
		if nw.ProviderDetails[assets.GcloudProvider] != nil {
			if err := ptypes.UnmarshalAny(
				nw.ProviderDetails[assets.GcloudProvider], &gnw); err != nil {
				return err
			}
			log.Printf("network %s already has details\n", nw.Name)
		} else {
			gnw := &assets.GCloudNetwork{
				Addresses: make([]*assets.GCloudIpAddress, 0),
			}
			for i := int32(0); i < nw.IpAddresses; i++ {
				addr := &assets.GCloudIpAddress{
					Type:      "EXTERNAL",
					Network:   nw.Name,
					Region:    "",
					Purpose:   "NAT_AUTO",
					Status:    "IN_USE",
					Ephemeral: false,
				}
				gnw.Addresses = append(gnw.Addresses, addr)
			}
			details, err := ptypes.MarshalAny(gnw)
			if err != nil {
				return err
			}
			nw.ProviderDetails[assets.GcloudProvider] = details
		}
		for _, snw := range nw.Subnetworks {
			if snw.Location == nil {
				return fmt.Errorf("missing subnetwork location")
			}
			if snw.ProviderDetails == nil {
				snw.ProviderDetails = make(map[string](*anypb.Any))
			}
			if snw.ProviderDetails[assets.GcloudProvider] != nil {
				var gsw assets.GCloudSubnetwork
				if err := ptypes.UnmarshalAny(
					snw.ProviderDetails[assets.GcloudProvider],
					&gsw); err != nil {
					return err
				}
				log.Printf("subnetwork %s already has details\n",
					snw.Name)
			} else {
				locstring := common.PrintLocation(*snw.Location)
				regions := resolveSpecLocation(*snw.Location,
					locations[locstring])
				if len(regions) == 0 {
					return fmt.Errorf(
						"provider %s does not support regions matching location %v",
						assets.GcloudProvider, snw.Location)
				}
				details, err := ptypes.MarshalAny(&assets.GCloudSubnetwork{
					Region: regions[0], // only using first region
				})
				if err != nil {
					return err
				}
				snw.ProviderDetails[assets.GcloudProvider] = details
			}
		}
	}
	for _, dset := range p.DiskSets {
		if dset.Template.Location == nil {
			return fmt.Errorf("missing diskset location information")
		}
		if dset.Template.Type == nil {
			return fmt.Errorf("missing diskset template information")
		}
		if dset.Template.ProviderDetails == nil {
			dset.Template.ProviderDetails = make(map[string](*anypb.Any))
		}
		if dset.Template.ProviderDetails[assets.GcloudProvider] != nil {
			// If there are provider details, just check consistency.
			var dsk assets.GCloudDisk
			if err := ptypes.UnmarshalAny(
				dset.Template.ProviderDetails[assets.GcloudProvider],
				&dsk); err != nil {
				return err
			}
			if err := checkDiskSpec(db, dsk, *dset.Template); err != nil {
				return err
			}
			log.Printf("Disk Set %s already has details for provider %s, leaving them as they are.\n",
				dset.Name, assets.GcloudProvider)
		} else { // There are no provider details yet.
			// Get regions for spec location.
			locstring := common.PrintLocation(*dset.Template.Location)
			regions := resolveSpecLocation(*dset.Template.Location, locations[locstring])
			if len(regions) == 0 {
				return fmt.Errorf("provider %s does not support regions matching location %v", assets.GcloudProvider, dset.Template.Location)
			}
			dt, r, err := getDiskTypeBySpec(db, *dset.Template.Type, regions)
			if err != nil {
				return err
			}
			details, err := ptypes.MarshalAny(&assets.GCloudDisk{
				DiskType: dt,
				Region:   r[0], // only using first region
			})
			if err != nil {
				return err
			}
			dset.Template.ProviderDetails[assets.GcloudProvider] = details
		}
	}
	return nil
}

// Populates spec from provider details if the spec is empty.
func FillInSpec(db *sql.DB, p *common.Project) error {
	for _, vmset := range p.InstanceSets {
		if vmset.Template.ProviderDetails == nil {
			return fmt.Errorf("Missing provider details for instance set %s\n",
				vmset.Name)
		}
		var gvm assets.GCloudVM
		err := ptypes.UnmarshalAny(
			vmset.Template.ProviderDetails[assets.GcloudProvider], &gvm)
		if err != nil {
			return err
		}
		if err = checkVmSpec(db, gvm, *vmset.Template); err != nil {
			return err
		}
		mt, err := GetMachineType(db, gvm.MachineType, gvm.Region)
		if err != nil {
			return err
		}
		if vmset.Template.Location == nil {
			loc, err := resolveLocation(gvm.Region)
			if err != nil {
				return err
			}
			vmset.Template.Location = &loc
		}
		if vmset.Template.Type == nil {
			vmset.Template.Type = &common.MachineType{
				CpuCount: mt.CpuCount,
				MemoryGb: uint32(mt.MemoryMb / 1000),
				GpuCount: mt.GpuCount,
			}
		}
		if vmset.UsageHoursPerMonth == 0 {
			vmset.UsageHoursPerMonth = 24 * 30 // full usage
		}
		if vmset.Template.Os == "" {
			if gvm.OsChoice != "" {
				choice := assets.OsChoiceByName(gvm.OsChoice)
				if assets.IsLinux(choice) {
					vmset.Template.Os = "linux"
				} else if assets.IsWindows(choice) {
					vmset.Template.Os = "windows"
				} else {
					fmt.Errorf("unsupported os choice %s", gvm.OsChoice)
				}
			}
		}
	}
	for _, dset := range p.DiskSets {
		if dset.Template.ProviderDetails == nil {
			return fmt.Errorf("Missing provider details for disk set %s\n",
				dset.Name)
		}
		var dsk assets.GCloudDisk
		err := ptypes.UnmarshalAny(
			dset.Template.ProviderDetails[assets.GcloudProvider], &dsk)
		if err != nil {
			return err
		}
		dt, err := getDiskType(db, dsk.DiskType, dsk.Region)
		if err != nil {
			return err
		}
		if dset.Template.Location == nil {
			loc, err := resolveLocation(dsk.Region)
			if err != nil {
				return err
			}
			dset.Template.Location = &loc
		}
		if dset.Template.Type == nil {
			tech := "SSD"
			if dsk.DiskType == "pd-standard" {
				tech = "Standard"
			}
			dset.Template.Type = &common.DiskType{
				DiskTech: tech,
				SizeGb:   uint32(dt.DefaultSizeGb),
			}
		}
		if dset.UsageHoursPerMonth == 0 {
			dset.UsageHoursPerMonth = 30 * 24 // full usage
		}
	}
	for _, nw := range p.Networks {
		for _, snw := range nw.Subnetworks {
			if snw.ProviderDetails == nil {
				return fmt.Errorf("Missing provider details for subnetwork %s\n",
					snw.Name)
			}
			var sw assets.GCloudSubnetwork
			err := ptypes.UnmarshalAny(snw.ProviderDetails[assets.GcloudProvider], &sw)
			if err != nil {
				return err
			}
			if snw.Location == nil {
				loc, err := resolveLocation(sw.Region)
				if err != nil {
					return err
				}
				snw.Location = &loc
			}
		}
	}
	return nil
}

// Returns all regions consistent with loc.
// Assume loc is internally consistent.
// If preferred is not empty, and is contained in the possible regions
// for the location, return only preferred.
func resolveSpecLocation(loc common.Location, preferred string) []string {
	var regions []string
	if loc.CountryCode != "" {
		regions = RegionsByCountry(loc.CountryCode)
	} else if loc.Continent != "" {
		regions = RegionsByContinent(geo.ContinentFromString(loc.Continent))
	} else if loc.GlobalRegion != "" {
		regions = make([]string, 0)
		continents := geo.GetContinents(geo.RegionFromString(loc.GlobalRegion))
		for _, c := range continents {
			regions = append(regions, RegionsByContinent(c)...)
		}
	}
	if len(regions) == 0 {
		if preferred == "" {
			return RegionsByCountry("US") // default to US if nothing specified.
		} else {
			regions = RegionsByCountry("US")
		}
	}
	if preferred != "" {
		for _, r := range regions {
			if r == preferred {
				return []string{r}
			}
		}
	}
	return regions
}

func resolveLocation(region string) (common.Location, error) {
	cc := RegionCountry(region)
	if cc == "Unknown" {
		return common.Location{}, fmt.Errorf("unknown region %s", region)
	}
	continent, gr := geo.GetContinent(cc)
	if continent == geo.UnknownC {
		return common.Location{},
			fmt.Errorf("no continent known for region %s.", region)
	}
	return common.Location{
		GlobalRegion: gr.String(),
		Continent:    continent.String(),
		CountryCode:  cc,
	}, nil
}

// Returns nil if region is consistent with the spec location,
// an error otherwise.
func checkLocation(region string, spec common.Location) error {
	loc, err := resolveLocation(region)
	if err != nil {
		return err
	}
	if spec.GlobalRegion != "" && spec.GlobalRegion != loc.GlobalRegion {
		return fmt.Errorf("spec global region %s does not match provider details (%s): %s",
			spec.GlobalRegion, assets.GcloudProvider, loc.GlobalRegion)
	}
	if spec.Continent != "" && spec.Continent != loc.Continent {
		return fmt.Errorf("spec continent %s does not match provider details (%s): %s",
			spec.Continent, assets.GcloudProvider, loc.Continent)
	}
	if spec.CountryCode != "" && spec.CountryCode != loc.CountryCode {
		return fmt.Errorf("spec country %s does not match provider details (%s): %s",
			spec.CountryCode, assets.GcloudProvider, loc.CountryCode)
	}
	return nil
}

// Retrieves a disk type satisfying the spec and available in at least
// one of the regions provided.
func getDiskTypeBySpec(db *sql.DB, dt common.DiskType, r []string) (
	string, []string, error) {
	var regionsClause strings.Builder
	fmt.Fprintf(&regionsClause, "(")
	for idx, region := range r {
		fmt.Fprintf(&regionsClause, "'%s'", region)
		if idx < len(r)-1 {
			fmt.Fprintf(&regionsClause, ",")
		}
	}
	fmt.Fprintf(&regionsClause, ")")
	var typeClause string
	if dt.DiskTech == "Standard" {
		typeClause = "DiskType='pd-standard'"
	} else {
		typeClause = "DiskType IN ('pd-ssd', 'pd-balanced')"
	}
	queryDiskType := fmt.Sprintf(`SELECT DISTINCT DiskType, Region from DiskTypes WHERE DefaultSizeGb >= %d AND Region in %s AND %s`,
		dt.SizeGb, regionsClause.String(), typeClause)
	res, err := db.Query(queryDiskType)
	if err != nil {
		return "", []string{}, err
	}
	defer res.Close()
	var dtype string
	var reg string
	presence := make(map[string][]string)
	for res.Next() {
		err = res.Scan(&dtype, &reg)
		if err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}
		fmt.Printf("found one: %s in %s\n", dtype, reg)
		if presence[dtype] == nil {
			presence[dtype] = []string{reg}
		} else {
			presence[dtype] = append(presence[dtype], reg)
		}
	}
	if len(presence) == 0 {
		return "", nil, fmt.Errorf("Failed to find a suitable disk type for %+v in %v", dt, r)
	}
	best_typename := "pd-ssd"
	for typename, regions := range presence {
		if typename == "pd-standard" {
			return typename, regions, nil
		}
		if typename == "pd-balanced" {
			return typename, regions, nil
		}
	}
	return best_typename, presence[best_typename], nil
}

// Retrieves a machine type satisfying the spec and available in at least one
// of the regions provided. Returns the machine type and the list of regions
// where it is available. If several machine types match the spec, the smallest
// one is returned. If there are several smallest types, the order of preference
// is E2 > N2 > N2D > N1 (based on a generic "TCO" consideration).
// TODO: probably want to take preemptible status, sole tenancy, commitments into account here.
// TODO: also query by gpu count (also, if gpu requested, only look at a2)
func getMachineTypeBySpec(db *sql.DB, st common.MachineType, r []string) (
	string, []string, error) {
	var regionsClause strings.Builder
	fmt.Fprintf(&regionsClause, "(")
	for idx, region := range r {
		fmt.Fprintf(&regionsClause, "'%s'", region)
		if idx < len(r)-1 {
			fmt.Fprintf(&regionsClause, ",")
		}
	}
	fmt.Fprintf(&regionsClause, ")")
	queryMachineType := fmt.Sprintf(`SELECT DISTINCT mt.MachineType, rz.Region from MachineTypes  mt join MachineTypesByZone mtbz on mt.MachineType=mtbz.MachineType JOIN REGIONZONE rz on mtbz.Zone=rz.Zone WHERE rz.Region in %s AND mt.CpuCount >= %d AND mt.CpuCount <= %d AND mt.MemoryMb >= %d AND mt.MemoryMb <= %d ORDER BY mt.CpuCount ASC, mt.MemoryMb asc LIMIT 10;`,
		regionsClause.String(), st.CpuCount, st.CpuCount*2,
		st.MemoryGb*1000, st.MemoryGb*2000)

	res, err := db.Query(queryMachineType)
	if err != nil {
		return "", []string{}, err
	}
	defer res.Close()
	var mt string
	var reg string
	presence := make(map[string][]string)
	for res.Next() {
		err = res.Scan(&mt, &reg)
		if err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}
		fmt.Printf("found one: %s in %s\n", mt, reg)
		if presence[mt] == nil {
			presence[mt] = []string{reg}
		} else {
			presence[mt] = append(presence[mt], reg)
		}
	}
	best_typename := ""
	for typename, regions := range presence {
		if strings.HasPrefix(typename, "e2-") {
			return typename, regions, nil
		}
		if strings.HasPrefix(typename, "n2-") {
			best_typename = typename
			continue
		}
		if strings.HasPrefix(typename, "n2d-") && strings.HasPrefix(best_typename, "n1-") {
			best_typename = typename
			continue
		}
		if strings.HasPrefix(typename, "n1-") && best_typename == "" {
			best_typename = typename
			continue
		}
		if best_typename == "" {
			best_typename = typename
		}
	}
	if best_typename != "" {
		return best_typename, presence[best_typename], nil
	}
	return "", nil, fmt.Errorf("Failed to find a suitable machine type for %v in %v", st, r)
}

// Retrieves a machine type by type name and region.
func GetMachineType(db *sql.DB, mt string, region string) (
	assets.MachineType, error) {
	queryMachineType := ""
	if region == "" {
		queryMachineType = fmt.Sprintf(`SELECT CpuCount, MemoryMb, IsSharedCpu
		FROM MachineTypes where MachineType='%s';`, mt)
	} else {
		queryMachineType = fmt.Sprintf(`SELECT mt.CpuCount, mt.MemoryMb, mt.IsSharedCpu FROM MachineTypes mt JOIN MachineTypesByZone mtbz on mt.MachineType=mtbz.MachineType JOIN RegionZone rz on mtbz.Zone=rz.Zone WHERE rz.Region='%s' AND mt.MachineType='%s';`, region, mt)
	}
	res, err := db.Query(queryMachineType)
	if err != nil {
		return assets.MachineType{}, err
	}
	defer res.Close()
	var cpuCount int64
	var memoryMb int64
	var isSharedCpu int32
	for res.Next() {
		err = res.Scan(&cpuCount, &memoryMb, &isSharedCpu)
		if err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}
		shared := isSharedCpu != 0
		return assets.MachineType{
			Name:        mt,
			CpuCount:    uint32(cpuCount),
			MemoryMb:    uint64(memoryMb),
			IsSharedCpu: shared,
		}, nil
	}
	return assets.MachineType{}, fmt.Errorf("Failed to find machine type %s\n", mt)
}

func getDiskType(db *sql.DB, dt string, region string) (assets.DiskType, error) {
	queryDiskType := ""
	if region == "" {
		queryDiskType = fmt.Sprintf(`SELECT DefaultSizeGb, Region 
		FROM DiskTypes where DiskType='%s' and Region='None';`, dt)
	} else {
		queryDiskType = fmt.Sprintf(`SELECT DefaultSizeGb, Region 
		FROM DiskTypes where DiskType='%s' and Region='%s';`, dt, region)
	}
	res, err := db.Query(queryDiskType)
	if err != nil {
		return assets.DiskType{}, err
	}
	defer res.Close()
	var r string
	var defaultSizeGb int64
	for res.Next() {
		err = res.Scan(&defaultSizeGb, &r)
		if err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}
		return assets.DiskType{
			Name:          dt,
			DefaultSizeGb: defaultSizeGb,
			Region:        region, // is this right?

		}, nil
	}
	return assets.DiskType{}, fmt.Errorf("Failed to find disk type %s in region %s\n", dt, region)
}
