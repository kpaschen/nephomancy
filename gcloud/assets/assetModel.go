package assets

import (
	"fmt"
	"strings"
)

type Firewall struct {
	a *SmallAsset
	networkName string
	Name string
}

type Subnetwork struct {
	a *SmallAsset
	IpRange string
	Region string
	Name string
	networkName string
}

type Route struct {
	a *SmallAsset
	IpRange string
	networkName string
}

type Network struct {
	a *SmallAsset
        Routes []Route
	Subnetworks []Subnetwork
	Firewalls []Firewall
	NetworkName string
}

type ServiceAccountKey struct {
	a *SmallAsset
	serviceAccountName string
}

type ServiceAccount struct {
	a *SmallAsset
	ServiceAccountName string
        keys []ServiceAccountKey
}

type Service struct {
	a *SmallAsset
	Name string
	State string
}

type Image struct {
	a *SmallAsset
	Licenses []string
	sourceDiskName string
	Name string
	StorageSize int64
}

type Disk struct {
	a *SmallAsset
	DiskName string
	IsRegional bool
	ZoneOrRegion string
	SourceImage *Image
	Size int64
	DiskType string
}

type Nic struct {
	Name string
	NetworkName string
	SubnetworkName string
}

type Instance struct {
	a *SmallAsset
	// instances with the same zone, machine type and scheduling settings
	// should have the same costs and can be combined for accounting.
	// the fingerprint is a concatenation of zone, machine type and scheduling.
	fingerprint string
	Zone string
	MachineType string
	Scheduling string
	Nics []Nic
	Name string
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
	if a.ResourceMap["networkInterfaces"] != nil {
		nw, _ := a.ResourceMap["networkInterfaces"].([]interface{})
		for _, nwi := range nw {
			networkInterface, _ := nwi.(map[string]interface{})
			nicName, _ := networkInterface["name"].(string)
			x, _ := networkInterface["network"].(string)
			parts := strings.Split(x, "/")
			networkName := parts[len(parts)-1]
			x, _ = networkInterface["subnetwork"].(string)
			parts = strings.Split(x, "/")
			subnetworkName := parts[len(parts)-1]
			nic := Nic {
				Name: nicName,
				NetworkName: networkName,
				SubnetworkName: subnetworkName,
			}
			nics = append(nics, nic)
		}
	}
	zone, _ := a.Zone()
	machineType, _ := a.MachineType()
	scheduling, _ := a.Scheduling()
	fingerprint := fmt.Sprintf("%s:%s:%s", machineType, zone, scheduling)

	inst := Instance{
		a: &a,
		Zone: zone,
		MachineType: machineType,
		Scheduling: scheduling,
		fingerprint: fingerprint,
		Nics: nics,
		Name: name,
	}

	as.Instances = append(as.Instances, &inst)
        return nil
}

func (as *AssetStructure) buildDisk(a SmallAsset, isRegional bool) error {
	diskName, _ := a.ResourceMap["name"].(string)
	regions, err := a.Regions()
	if err != nil || len(regions) == 0 {
		return fmt.Errorf("missing both zone and region for disk %+v\n", a)
	}
	if len(regions) > 1 {
		return fmt.Errorf("multiple regions for regional disk? %+v\n", a)
	}
	zoneOrRegion := regions[0]
	storageSize, _ := a.StorageSize()
	diskType, _ := a.DiskType()
	disk := Disk{
		a: &a,
		DiskName: diskName,
		IsRegional: isRegional,
		ZoneOrRegion: zoneOrRegion,
		Size: storageSize,
		DiskType: diskType,
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
	if a.ResourceMap["sourceDisk"] == nil {
		return fmt.Errorf("missing sourceDisk resource for image %+v\n", a)
	}
	sdn, _ := a.ResourceMap["sourceDisk"].(string)
	parts := strings.Split(sdn, "/")
	sourceDiskName := parts[len(parts)-1]
	licenses := make([]string, 0)
	if a.ResourceMap["licenses"] != nil {
		lcs, _ := a.ResourceMap["licenses"].([]interface{})
		for _, l := range lcs {
			lc, _ := l.(string)
			licenses = append(licenses, lc)
		}
	}
	idParts := strings.Split(a.Name, "/")
	storageSize, _ := a.StorageSize()
	img := Image{
		a: &a,
		Licenses: licenses,
		sourceDiskName: sourceDiskName,
		StorageSize: storageSize,
		Name: idParts[len(idParts)-1],
	}
	found := false
	for _, d := range as.Disks {
		if d.DiskName == sourceDiskName {
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

func (as *AssetStructure) buildService(a SmallAsset) error {
	name, _ := a.ResourceMap["name"].(string)
	state, _ := a.ResourceMap["state"].(string)
	sv := Service{
		a: &a,
		Name: name,
		State: state,
	}
	as.Services = append(as.Services, &sv)
	return nil
}

func (as *AssetStructure) buildServiceAccountKey(a SmallAsset) error {
	name, err := a.ServiceAccountName()
	if err != nil {
		return err
	}
	sa := ServiceAccountKey{
		a: &a,
		serviceAccountName: name,
	}
	found := false
	for _, s := range as.ServiceAccounts {
		if s.ServiceAccountName == name {
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
	name, err := a.ServiceAccountName()
	if err != nil {
		return err
	}
	sa := ServiceAccount{
		a: &a,
		ServiceAccountName: name,
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
		a: &a,
		NetworkName: networkName,
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
			n.Subnetworks = append(n.Subnetworks, s)
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
	networkName, _ := a.NetworkName()
	name, _  := a.ResourceMap["name"].(string)
	f := Firewall{
		a: &a,
		networkName: networkName,
		Name: name,
	}
	found := false
	for _, n := range as.Networks {
		if n.NetworkName == networkName {
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
	ipRange, _ := a.ResourceMap["destRange"].(string)
	networkName, _ := a.NetworkName()
	r := Route{
		a: &a,
		IpRange: ipRange,
		networkName: networkName,
	}
	// Do we already have the network?
	found := false
	for _, n := range as.Networks {
		if n.NetworkName == networkName {
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
	ipRange, _ := a.ResourceMap["ipCidrRange"].(string)
	name, _ := a.ResourceMap["name"].(string)
	fullRegion, _ := a.ResourceMap["region"].(string)
	parts := strings.Split(fullRegion, "/")
	region := parts[len(parts)-1]
	network, _ := a.NetworkName()
        s := Subnetwork{
		a: &a,
		IpRange: ipRange,
		Region: region,
		Name: name,
		networkName: network,
	}
	found := false
	for _, n := range as.Networks {
		if n.NetworkName == network {
			// TODO: avoid adding duplicates.
			n.Subnetworks = append(n.Subnetworks, s)
			found = true
			break
		}
	}
	if !found {
		as.danglingSubnetworks = append(as.danglingSubnetworks, s)
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
		case "Project": continue
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
	return &ret, nil
}
