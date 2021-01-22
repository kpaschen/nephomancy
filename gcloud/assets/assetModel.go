package assets

import (
	"fmt"
	"strings"
)

type ResourceUsage struct {
	UsageUnit string // same unit as in PricingExpression
	MaxUsage int64  // Min is 0 for inactive resources, max depends on things like num cpus
}

type BaseAsset interface {
	BillingService() string
	ResourceFamily() string
	Regions() []string
	// Returns a map of resource (e.g. cpu, egress) to usage struct.
	MaxResourceUsage() (map[string]ResourceUsage, error)
}

const BS_COMPUTE = "6F81-5844-456A"
const BS_CONTAINER = "CCD8-9BF1-090E"
const BS_MONITORING = "58CD-E7C3-72CA"
const BS_TODO = "TODO"
// TODO: others

type Firewall struct {
	Name string
	networkName string
}

func (Firewall) ResourceFamily() string { return "Network" }
func (Firewall) BillingService() string { return BS_COMPUTE }
func (Firewall) Regions() []string { return nil }
func (Firewall) MaxResourceUsage() (map[string]ResourceUsage, error) { return nil, nil }

type Subnetwork struct {
	Name string
	IpRange string
	Region string
	networkName string
	MaxTier string
}
func (Subnetwork) ResourceFamily() string { return "Network" }
func (Subnetwork) BillingService() string { return BS_COMPUTE }
func (s Subnetwork) Regions() []string { return []string{s.Region} }
func (Subnetwork) MaxResourceUsage() (map[string]ResourceUsage, error) {
	return map[string]ResourceUsage{
		"egress": ResourceUsage{
			UsageUnit: "GiBy",
			MaxUsage: -1, // is this unlimited?
		},
		"ingress": ResourceUsage{
			UsageUnit: "GiBy",
			MaxUsage: -1,
		},
	}, nil
}

type Route struct {
	IpRange string
	networkName string
}
func (Route) ResourceFamily() string { return "Network" }
func (Route) BillingService() string { return BS_COMPUTE }
func (Route) Regions() []string { return nil }
func (Route) MaxResourceUsage() (map[string]ResourceUsage, error) { return nil, nil }

type Network struct {
	Name string
        Routes []Route
	Subnetworks []*Subnetwork
	Firewalls []Firewall
}
func (Network) ResourceFamily() string { return "Network" }
func (Network) BillingService() string { return BS_COMPUTE }
func (Network) Regions() []string { return nil }
func (Network) MaxResourceUsage() (map[string]ResourceUsage, error) {
	return map[string]ResourceUsage{
		"egress": ResourceUsage{
			UsageUnit: "GiBy",
			MaxUsage: -1, // is this unlimited?
		},
		"ingress": ResourceUsage{
			UsageUnit: "GiBy",
			MaxUsage: -1,
		},
	}, nil
}

type ServiceAccountKey struct {
	serviceAccountName string
}
func (ServiceAccountKey) ResourceFamily() string { return "IAM" }
func (ServiceAccountKey) BillingService() string { return BS_TODO }
func (ServiceAccountKey) Regions() []string { return nil }
func (ServiceAccountKey) MaxResourceUsage() (map[string]ResourceUsage, error) { return nil, nil }

type ServiceAccount struct {
	Name string
        keys []ServiceAccountKey
}
func (ServiceAccount) ResourceFamily() string { return "IAM" }
func (ServiceAccount) BillingService() string { return BS_TODO }
func (ServiceAccount) Regions() []string { return nil }
func (ServiceAccount) MaxResourceUsage() (map[string]ResourceUsage, error) { return nil, nil }

type Service struct {
	Name string
	State string
}
func (Service) ResourceFamily() string { return "Service" }
func (Service) BillingService() string { return BS_TODO }
func (Service) Regions() []string { return nil }
func (Service) MaxResourceUsage() (map[string]ResourceUsage, error) {
	return map[string]ResourceUsage{
		"calls": ResourceUsage{
			UsageUnit: "count",
			MaxUsage: -1,
		},
	}, nil
}

type Image struct {
	Name string
	Licenses []string
	sourceDiskName string
	StorageSize int64
	StorageLocations []string
}
func (Image) ResourceFamily() string { return "Storage" }
func (Image) BillingService() string { return BS_COMPUTE }
func (i Image) Regions() []string { return i.StorageLocations }
func (i Image) MaxResourceUsage() (map[string]ResourceUsage, error) {
	return map[string]ResourceUsage{
		"diskspace": ResourceUsage{
			UsageUnit: "GiBy.mo",
			MaxUsage: i.StorageSize,
		},
	}, nil
}

type Disk struct {
	Name string
	IsRegional bool
	ZoneOrRegion string
	SourceImage *Image
	Size int64
	DiskTypeName string
	DiskType DiskType
}
func (Disk) ResourceFamily() string { return "Storage" }
func (Disk) BillingService() string { return BS_COMPUTE }
func (d Disk) Regions() []string {
	if d.IsRegional {
		return []string{d.ZoneOrRegion}
	} else {
		return nil
	}
}
func (d Disk) MaxResourceUsage() (map[string]ResourceUsage, error) {
	return map[string]ResourceUsage{
		"diskspace": ResourceUsage{
			UsageUnit: "GiBy.mo",
			MaxUsage: d.Size,
		},
	}, nil
}

type Nic struct {
	Name string
	NetworkName string
	SubnetworkName string
	NetworkTier string
}

type Instance struct {
	Name string
	// instances with the same zone, machine type and scheduling settings
	// should have the same costs and can be combined for accounting.
	// the fingerprint is a concatenation of zone, machine type and scheduling.
	fingerprint string
	Zone string
	MachineTypeName string
	MachineType MachineType
	Scheduling string
	Nics []Nic
}
func (Instance) ResourceFamily() string { return "Compute" }
func (Instance) BillingService() string { return BS_COMPUTE }
func (i Instance) Regions() []string {
	if i.Zone != "" {
		parts := strings.Split(i.Zone, "-")
		return []string{fmt.Sprintf("%s-%s", parts[0], parts[1])}
	}
	return nil
}
func (i Instance) MaxResourceUsage() (map[string]ResourceUsage, error) {
	if i.MachineType.CpuCount == 0 {
		return nil, fmt.Errorf("Missing compute metadata for instance %+v\n", i)
	}
	cpuCount := i.MachineType.CpuCount
	memoryGb := i.MachineType.MemoryMb / 1024
	// TODO: gpus, shared cpu
	return map[string]ResourceUsage{
		"cpu": ResourceUsage{
			UsageUnit: "h",
			MaxUsage: 30 * 24 * cpuCount,
		},
		"memory": ResourceUsage{
			UsageUnit: "GiBy.h",
			MaxUsage: 30 * 24 * memoryGb,
		},
	}, nil
}

// This is basically a project.
type AssetStructure struct {
	Instances []*Instance
	// Assume all subnetworks, routes, firewalls etc. belong to a network.
	// But not all networks are used by an instance necessarily.
	Networks []*Network
	// Assume all images and licenses belong to a disk. But not all disks belong
	// to an instance, hence they have their own top level entry.
	Disks []*Disk
	ServiceAccounts []*ServiceAccount
	Services []*Service
	DefaultNetworkTier string

	danglingRoutes []Route
	danglingSubnetworks []Subnetwork
	danglingFirewalls []Firewall
	danglingKeys []ServiceAccountKey
	danglingImages []Image
}

func (as *AssetStructure) buildInstance(a SmallAsset) error {
	idParts := strings.Split(a.Name, "/")
	name := idParts[len(idParts)-1]
	nics := make([]Nic, 0)
	if a.resourceMap["networkInterfaces"] != nil {
		nw, _ := a.resourceMap["networkInterfaces"].([]interface{})
		for _, nwi := range nw {
			networkInterface, _ := nwi.(map[string]interface{})
			nicName, _ := networkInterface["name"].(string)
			x, _ := networkInterface["network"].(string)
			parts := strings.Split(x, "/")
			networkName := parts[len(parts)-1]
			x, _ = networkInterface["subnetwork"].(string)
			parts = strings.Split(x, "/")
			subnetworkName := parts[len(parts)-1]
			ac, _ := networkInterface["accessConfigs"].([]interface{})
			tier := ""  // if not set, the project's default tier will take effect.
			// Currently, there is only at most one config, and the type is always
			// ONE_TO_ONE_NAT. When there is no config, the instance has no
			// access to the internet.
			for _, config := range ac {
				cfg := config.(map[string]interface{})
				tier, _ = cfg["networkTier"].(string)
			}
			nic := Nic {
				Name: nicName,
				NetworkName: networkName,
				SubnetworkName: subnetworkName,
				NetworkTier: tier,
			}
			nics = append(nics, nic)
		}
	}
	zone, _ := a.zone()
	machineType, _ := a.machineType()
	scheduling, _ := a.scheduling()
	fingerprint := fmt.Sprintf("%s:%s:%s", machineType, zone, scheduling)

	inst := Instance{
		Name: name,
		Zone: zone,
		MachineTypeName: machineType,
		Scheduling: scheduling,
		fingerprint: fingerprint,
		Nics: nics,
	}

	as.Instances = append(as.Instances, &inst)
        return nil
}

func (as *AssetStructure) buildDisk(a SmallAsset, isRegional bool) error {
	diskName, _ := a.resourceMap["name"].(string)
	regions, err := a.regions()
	if err != nil || len(regions) == 0 {
		return fmt.Errorf("missing both zone and region for disk %+v\n", a)
	}
	if len(regions) > 1 {
		return fmt.Errorf("multiple regions for regional disk? %+v\n", a)
	}
	zoneOrRegion := regions[0]
	storageSize, _ := a.storageSize()
	diskType, _ := a.diskType()
	disk := Disk{
		Name: diskName,
		IsRegional: isRegional,
		ZoneOrRegion: zoneOrRegion,
		Size: storageSize,
		DiskTypeName: diskType,
	}
	di := make([]Image, 0)
	for _, img := range as.danglingImages {
		if img.sourceDiskName == diskName {
			disk.SourceImage = &img
		} else {
			di = append(di, img)
		}
	}
	as.danglingImages = di
        as.Disks = append(as.Disks, &disk)
	return nil
}

func (as *AssetStructure) buildImage(a SmallAsset) error {
	if a.resourceMap["sourceDisk"] == nil {
		return fmt.Errorf("missing sourceDisk resource for image %+v\n", a)
	}
	sdn, _ := a.resourceMap["sourceDisk"].(string)
	parts := strings.Split(sdn, "/")
	sourceDiskName := parts[len(parts)-1]
	licenses := make([]string, 0)
	if a.resourceMap["licenses"] != nil {
		lcs, _ := a.resourceMap["licenses"].([]interface{})
		for _, l := range lcs {
			lc, _ := l.(string)
			licenses = append(licenses, lc)
		}
	}
	idParts := strings.Split(a.Name, "/")
	storageSize, _ := a.storageSize()
	regions := make([]string, 0)
	storageLocations, _ := a.resourceMap["storageLocations"].([]interface{})
	for _, stl := range storageLocations {
		loc, _ := stl.(string)
		regions = append(regions, loc)
	}
	img := Image{
		Name: idParts[len(idParts)-1],
		Licenses: licenses,
		sourceDiskName: sourceDiskName,
		StorageSize: storageSize,
		StorageLocations: regions,
	}
	found := false
	for _, d := range as.Disks {
		if d.Name == sourceDiskName {
			d.SourceImage = &img
			found = true
			break
		}
	}
	if !found {
		as.danglingImages = append(as.danglingImages, img)
	}
	return nil
}

// only interested in the default network tier for now
func (as *AssetStructure) buildProject(a SmallAsset) error {
	as.DefaultNetworkTier = "PREMIUM"
	if a.resourceMap["defaultNetworkTier"] != nil {
		tier, _ := a.resourceMap["defaultNetworkTier"].(string)
		as.DefaultNetworkTier = tier
	}
	return nil
}

func (as *AssetStructure) buildService(a SmallAsset) error {
	name, _ := a.resourceMap["name"].(string)
	state, _ := a.resourceMap["state"].(string)
	sv := Service{
		Name: name,
		State: state,
	}
	as.Services = append(as.Services, &sv)
	return nil
}

func (as *AssetStructure) buildServiceAccountKey(a SmallAsset) error {
	name, err := a.serviceAccountName()
	if err != nil {
		return err
	}
	sa := ServiceAccountKey{
		serviceAccountName: name,
	}
	found := false
	for _, s := range as.ServiceAccounts {
		if s.Name == name {
			s.keys = append(s.keys, sa)
			found = true
			break
		}
	}
	if !found {
		as.danglingKeys = append(as.danglingKeys, sa)
	}
	return nil
}

func (as *AssetStructure) buildServiceAccount(a SmallAsset) error {
	name, err := a.serviceAccountName()
	if err != nil {
		return err
	}
	sa := ServiceAccount{
		Name: name,
	}
	dk := make([]ServiceAccountKey, 0)
	for _, d := range as.danglingKeys {
		if d.serviceAccountName == name {
			sa.keys = append(sa.keys, d)
		} else {
			dk = append(dk, d)
		}
	}
	as.danglingKeys = dk
	as.ServiceAccounts = append(as.ServiceAccounts, &sa)
	return nil
}

func (as *AssetStructure) buildNetwork(a SmallAsset) error {
	parts := strings.Split(a.Name, "/")
	networkName := parts[len(parts)-1]
	n := Network {
		Name: networkName,
	}
	dr := make([]Route, 0)
	for _, r := range as.danglingRoutes {
		if r.networkName != networkName {
			dr = append(dr, r)
		} else {
			n.Routes = append(n.Routes, r)
		}
	}
	as.danglingRoutes = dr

	ds := make([]Subnetwork, 0)
	for _, s := range as.danglingSubnetworks {
		if s.networkName != networkName {
			ds = append(ds, s)
		} else {
			n.Subnetworks = append(n.Subnetworks, &s)
		}
	}
	as.danglingSubnetworks = ds

	df := make([]Firewall, 0)
	for _, f := range as.danglingFirewalls {
		if f.networkName != networkName {
			df = append(df, f)
		} else {
			n.Firewalls = append(n.Firewalls, f)
		}
	}
	as.danglingFirewalls = df
	as.Networks = append(as.Networks, &n)
	return nil
}

func (as *AssetStructure) buildFirewall(a SmallAsset) error {
	networkName, _ := a.networkName()
	name, _  := a.resourceMap["name"].(string)
	f := Firewall{
		Name: name,
		networkName: networkName,
	}
	found := false
	for _, n := range as.Networks {
		if n.Name == networkName {
			// TODO: avoid adding duplicates.
			n.Firewalls = append(n.Firewalls, f)
			found = true
			break
		}
	}
	if !found {
		as.danglingFirewalls = append(as.danglingFirewalls, f)
	}
	return nil
}

func (as *AssetStructure) buildRoute(a SmallAsset) error {
	ipRange, _ := a.resourceMap["destRange"].(string)
	networkName, _ := a.networkName()
	r := Route{
		IpRange: ipRange,
		networkName: networkName,
	}
	// Do we already have the network?
	found := false
	for _, n := range as.Networks {
		if n.Name == networkName {
			// TODO: avoid adding duplicate routes.
			n.Routes = append(n.Routes, r)
			found = true
			break
		}
	}
	if !found {
		as.danglingRoutes = append(as.danglingRoutes, r)
	}
	return nil
}

func (as *AssetStructure) buildSubnetwork(a SmallAsset) error {
	ipRange, _ := a.resourceMap["ipCidrRange"].(string)
	name, _ := a.resourceMap["name"].(string)
	fullRegion, _ := a.resourceMap["region"].(string)
	parts := strings.Split(fullRegion, "/")
	region := parts[len(parts)-1]
	network, _ := a.networkName()
        s := Subnetwork{
		Name: name,
		IpRange: ipRange,
		Region: region,
		networkName: network,
	}
	found := false
	for _, n := range as.Networks {
		if n.Name == network {
			// TODO: avoid adding duplicates.
			n.Subnetworks = append(n.Subnetworks, &s)
			found = true
			break
		}
	}
	if !found {
		as.danglingSubnetworks = append(as.danglingSubnetworks, s)
	}
	return nil
}

func (as *AssetStructure) reconcile() error {
	if len(as.danglingRoutes) > 0 {
		return fmt.Errorf("orphaned routes: %+v\n", as.danglingRoutes)
	}
	if len(as.danglingSubnetworks) > 0 {
		return fmt.Errorf("orphaned Subnetworks: %+v\n", as.danglingSubnetworks)
	}
	if len(as.danglingFirewalls) > 0 {
		return fmt.Errorf("orphaned Firewalls: %+v\n", as.danglingFirewalls)
	}
	if len(as.danglingImages) > 0 {
		return fmt.Errorf("orphaned Images: %+v\n", as.danglingImages)
	}

	// See which subnetworks are in use by which instances, and record the maximum
	// network tier applicable per subnetwork.
	tier := as.DefaultNetworkTier
	for _, inst := range as.Instances {
		for _, nic := range inst.Nics {
			nwName := nic.NetworkName
			sName := nic.SubnetworkName
			found := false
			if nic.NetworkTier != "" {
				tier = nic.NetworkTier
			}
			for _, nw := range as.Networks {
				if nw.Name != nwName {
					continue
				}
				for _, sw := range nw.Subnetworks {
					if sw.Name != sName {
						continue
					}
					found = true
					// There are only two tiers, PREMIUM and STANDARD
					if sw.MaxTier != "PREMIUM" {
						sw.MaxTier = tier
					}
				}
			}
			if !found {
				return fmt.Errorf("Instance %+v refers to non-existent network/subnetwork %s/%s\n", inst, nwName, sName)
			}
		}
	}
	return nil
}

func BuildAssetStructure(ax []SmallAsset) (*AssetStructure, error) {
	ret := AssetStructure{}
	for _, as := range ax {
		err := as.ensureResourceMap()
		if err != nil {
			return nil, nil
		}
		bt, err := as.BaseType()
		if err != nil {
			return nil, nil
		}
		switch bt {
		case "Route": if err = ret.buildRoute(as); err != nil {
			return nil, err
		}
		case "Subnetwork": if err = ret.buildSubnetwork(as); err != nil {
			return nil, err
		}
		case "Firewall": if err = ret.buildFirewall(as); err != nil {
			return nil, err
		}
		case "Network": if err = ret.buildNetwork(as); err != nil {
			return nil, err
		}
		case "ServiceAccount": if err = ret.buildServiceAccount(as); err != nil {
			return nil, err
		}
		case "ServiceAccountKey": if err = ret.buildServiceAccountKey(as); err != nil {
			return nil, err
		}
		case "Service": if err = ret.buildService(as); err != nil {
			return nil, err
		}
		case "Project": if err = ret.buildProject(as); err != nil {
			return nil, err
		}
		case "Disk": if err = ret.buildDisk(as, false); err != nil {
			return nil, err
		}
		case "RegionDisk": if err = ret.buildDisk(as, true); err != nil {
			return nil, err
		}
		case "Image": if err = ret.buildImage(as); err != nil {
			return nil, err
		}
		case "Instance": if err = ret.buildInstance(as); err != nil {
			return nil, err
		}
	        default: fmt.Printf("not handling type %s yet\n", bt)
		continue
		}
	}
	if err := ret.reconcile(); err != nil {
		return nil, err
	}
	return &ret, nil
}
