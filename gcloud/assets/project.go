package assets

import (
	"fmt"
	"strings"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"github.com/golang/protobuf/ptypes"
	common "nephomancy/common/resources"
)

const GcloudProvider = "gcloud"

// This can be in common
func UnmarshalProject(projectAsJsonBytes []byte) (*common.Project, error) {
	p := &common.Project{}
	if err := protojson.Unmarshal(projectAsJsonBytes, p); err != nil {
		return nil, err
	}
	return p, nil
}

// BuildProject takes a list of small assets and create a project proto
// containing lists of vm sets, disk sets, and images.
func BuildProject(ax []SmallAsset) (*common.Project, error) {
	p := &common.Project{
		VmSets: make([]*common.VMSet, 0),
		DiskSets: make([]*common.DiskSet, 0),
		Images: make([]*common.Image, 0),
	}
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
		case "Instance":
			vm, err := createVM(as)
			if err != nil {
				return nil, err
			}
			if err = addVMToProject(p, vm); err != nil {
				return nil, err
			}
		case "Disk": {
			d, err := createDisk(as, false)
			if err != nil {
				return nil, err
			}
			if err = addDiskToProject(p, d); err != nil {
				return nil, err
			}
		}
		case "RegionDisk": {
			d, err := createDisk(as, true)
			if err != nil {
				return nil, err
			}
			if err = addDiskToProject(p, d); err != nil {
				return nil, err
			}
		}
		case "Image": {
			img, err := createImage(as)
			if err != nil {
				return nil, err
			}
			if err = addImageToProject(p, img); err != nil {
				return nil, err
			}
		}
		case "Project": {
			// There are two project resources, one for the actual
			// project and one for its parent. They have the project
			// name in different fields.
			err := as.ensureResourceMap()
			if err != nil {
				return nil, err
			}
			name, _ := as.resourceMap["projectId"].(string)
			if name != "" {
				p.Name = name
			}
		}
	        case "Firewall": {}
	        case "Route": {}
		case "Network": {
			nw, err := createNetwork(as)
			if err != nil {
				return nil, err
			}
			if err = addNetworkToProject(p, nw); err != nil {
				return nil, err
			}
		}
		case "Subnetwork": {
			snw, err := createSubnetwork(as)
			if err != nil {
				return nil, err
			}
			nwName, _ := as.networkName()
			if err = addSubnetworkToProject(p, snw, nwName); err != nil {
				return nil, err
			}
		}
		case "Service": {}
		case "ServiceAccount": {}
		case "ServiceAccountKey": {}
		default: fmt.Printf("type %s not handled yet\n", bt)
		}
	}
	if err := pruneSubnetworks(p); err != nil {
		return nil, err
	}
	return p, nil
}

func VmRegionZone(vm common.VM) (region string, zone string, err error) {
	var gvm GCloudVM
	if err := ptypes.UnmarshalAny(
		vm.ProviderDetails[GcloudProvider], &gvm); err != nil {
		return "", "", err
	}
	return gvm.Region, gvm.Zone, nil
}

func DiskRegionZone(disk common.Disk) (region string, zone string, err error) {
	var gdsk GCloudDisk
	if err := ptypes.UnmarshalAny(
		disk.ProviderDetails[GcloudProvider], &gdsk); err != nil {
		return "", "", err
	}
	return gdsk.Region, gdsk.Zone, nil
}

func SubnetworkRegionTier(subnetwork common.Subnetwork) (region string, tier string, err error) {
	var gsnw GCloudSubnetwork
	if err := ptypes.UnmarshalAny(
		subnetwork.ProviderDetails[GcloudProvider], &gsnw); err != nil {
			return "", "", err
	}
	return gsnw.Region, gsnw.Tier, nil
}

func vmNetworkTier(vm common.VM) (string, error) {
	var gvm GCloudVM
	if err := ptypes.UnmarshalAny(
		vm.ProviderDetails[GcloudProvider], &gvm); err != nil {
		return "", err
	}
        if gvm.NetworkTier == "" {
		return "STANDARD", nil
	}
	return gvm.NetworkTier, nil
}

func pruneSubnetworks(p *common.Project) error {
	regions := make(map[string]string)
	for _, vms := range p.VmSets {
		region, _, _ := VmRegionZone(*vms.Template)
		tier, _ := vmNetworkTier(*vms.Template)
		regions[region] = tier
	}
	for _, nw := range p.Networks {
		pruned := make([]*common.Subnetwork, 0)
		for _, snw := range nw.Subnetworks {
			var gsnw GCloudSubnetwork
			err := ptypes.UnmarshalAny(
				snw.ProviderDetails[GcloudProvider], &gsnw)
			if err != nil {
				return err
			}
			if regions[gsnw.Region] != "" {
				pruned = append(pruned, snw)
				gsnw.Tier = regions[gsnw.Region]
				details, _ := ptypes.MarshalAny(&gsnw)
				snw.ProviderDetails[GcloudProvider] = details
			}
		}
		nw.Subnetworks = pruned
	}
	return nil
}

func addSubnetworkToProject(p *common.Project, snw *common.Subnetwork, networkName string) error {
	for _, nw := range p.Networks {
		if nw.Name == networkName {
			nw.Subnetworks = append(nw.Subnetworks, snw)
			return nil
		}
	}
	nw := &common.Network{
		Name: networkName,
		Subnetworks: make([]*common.Subnetwork, 1),
	}
	nw.Subnetworks[0] = snw
	return addNetworkToProject(p, nw)
}

func createSubnetwork(a SmallAsset) (*common.Subnetwork, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	name, _ := a.resourceMap["name"].(string)
	fullRegion, _ := a.resourceMap["region"].(string)
	parts := strings.Split(fullRegion, "/")
	region := parts[len(parts)-1]
	details, _ := ptypes.MarshalAny(&GCloudSubnetwork{
		Region: region,
	})
	return &common.Subnetwork{
		Name: name,
		// Fill in values for traffic estimate.
		// Gcloud has a limit of 20Gbits/s per external IP address.
		IngressGbitsPerMonth: 20,
		// This is a quota metric: compute.googleapis.com/vm_to_internet_egress_bandwidth
		// The default value is 75 Gb total per region _per month_.
		// There is also a cap based on the VMs you are using,
		// but it is way more than the 75 Gbps per region.
		ExternalEgressGbitsPerMonth: 75,
		// There is an internal limit per VM, depending on the
		// machine type. It is between 2 and 32 Gbit/s.
		InternalEgressGbitsPerMonth: 100,
		ProviderDetails: map[string]*anypb.Any{
			GcloudProvider: details,
		},
	}, nil
}

func createNetwork(a SmallAsset) (*common.Network, error) {
	parts := strings.Split(a.Name, "/")
	networkName := parts[len(parts)-1]
	return &common.Network{
		Name: networkName,
	}, nil
}

func addNetworkToProject(p *common.Project, n *common.Network) error {
	for _, nw := range p.Networks {
		if nw.Name == n.Name {
			return nil
		}
	}
	p.Networks = append(p.Networks, n)
	return nil
}

// The fingerprints are internal only. This just creates a basic
// grouping of VMs into sets that is probably useful. There is
// no need in general for VMSets to be uniquely distinguishable.
func fingerprintVM(vm common.VM) (string, error) {
	region, _, err := VmRegionZone(vm)
	if err != nil {
		return "", err
	}
	if region == "" {
		return "", fmt.Errorf("missing region for vm %+v", vm)
	}
	var fp strings.Builder
	fmt.Fprintf(&fp, "%s:", region)
	if vm.Type != nil {
		fmt.Fprintf(&fp, "%s:", vm.Type)
	} else {
		var gvm GCloudVM
		err := ptypes.UnmarshalAny(vm.ProviderDetails[GcloudProvider], &gvm)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&fp, "%s", gvm.MachineType)
	}
	return fp.String(), nil
}

func addVMToProject(p *common.Project, vm *common.VM) error {
	fp, err := fingerprintVM(*vm)
	if err != nil {
		return err
	}
	for _, vmSet := range p.VmSets {
		f, _ := fingerprintVM(*vmSet.Template)
		if f == fp {
			vmSet.Count++
			return nil
		}
	}
	// No vmset with the given fingerprint yet
	vset := common.VMSet{
		Template: vm,
		Count: 1,
	}
	p.VmSets = append(p.VmSets, &vset)
	return nil
}

func createVM(a SmallAsset) (*common.VM, error) {
	networkTier := ""
	if a.resourceMap["networkInterfaces"] != nil {
		nw, _ := a.resourceMap["networkInterfaces"].([]interface{})
		for _, nwi := range nw {
			networkInterface, _ := nwi.(map[string]interface{})
			ac, _ := networkInterface["accessConfigs"].([]interface{})
			for _, config := range ac {
				cfg := config.(map[string]interface{})
				networkTier, _ = cfg["networkTier"].(string)
			}
		}
	}
	zone, _ := a.zone()
	regions, _ := a.regions()
	region := regions[0]
	machineType, _ := a.machineType()
	scheduling, _ := a.scheduling()

	details, err := ptypes.MarshalAny(&GCloudVM{
		MachineType: machineType,
		Scheduling: scheduling,
		NetworkTier: networkTier,
		Region: region,
		Zone: zone,
	})
	if err != nil {
		return nil, err
	}

	ret := common.VM{
		ProviderDetails: map[string]*anypb.Any{
			GcloudProvider: details,
		},
	}

        return &ret, nil
}

func addDiskToProject(p *common.Project, dsk *common.Disk) error {
	fp, err := fingerprintDisk(*dsk)
	if err != nil {
		return err
	}
	for _, dskSet := range p.DiskSets {
		f, _ := fingerprintDisk(*dskSet.Template)
		if f == fp {
			dskSet.Count++
			return nil
		}
	}
	// No disk set with the given fingerprint yet
	dset := common.DiskSet{
		Template: dsk,
		Count: 1,
	}
	p.DiskSets = append(p.DiskSets, &dset)
	return nil
}

func fingerprintDisk(disk common.Disk) (string, error) {
	region, _, _ := DiskRegionZone(disk)
	if region == "" {
		return "", fmt.Errorf("missing region for disk %+v", disk)
	}
	var fp strings.Builder
	fmt.Fprintf(&fp, "%s:", region)
	fmt.Fprintf(&fp, "%d:", disk.ActualSizeGb)
	if disk.Type != nil {
		fmt.Fprintf(&fp, "%s:", disk.Type)
	} else {
		var gdsk GCloudDisk
		err := ptypes.UnmarshalAny(disk.ProviderDetails[GcloudProvider], &gdsk)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&fp, "%s", gdsk.DiskType)
	}
	return fp.String(), nil
}

func createDisk(a SmallAsset, isRegional bool) (*common.Disk, error) {
	regions, err := a.regions()
	if err != nil || len(regions) == 0 {
		return nil, fmt.Errorf("missing both zone and region for disk %+v", a)
	}
	if len(regions) > 1 {
		// This is probably not an error, i just need to decide how
		// to handle it.
		return nil, fmt.Errorf("multiple regions for a disk? %+v", a)
	}
	sizeGB, _ := a.storageSize()
	zone, _ := a.zone()
	diskType, _ := a.diskType()
	details, err := ptypes.MarshalAny(&GCloudDisk{
		DiskType: diskType,
		IsRegional: isRegional,
		Region: regions[0],
		Zone: zone,
	})
	if err != nil {
		return nil, err
	}
	ret := common.Disk{
		ActualSizeGb: uint64(sizeGB),
		ProviderDetails: map[string]*anypb.Any{
			GcloudProvider: details,
		},
	}
	return &ret, nil
}

func createImage(a SmallAsset) (*common.Image, error) {
	// Not handling licenses yet (or maybe ever).
	size, _ := a.storageSize()
	regions, _ := a.regions()
	if len(regions) != 1 {
		return nil, fmt.Errorf("unexpected number of regions in image: %v", regions)
	}
	details, _ := ptypes.MarshalAny(&GCloudImage{
		Region: regions[0],
	})
	return &common.Image{
		SizeGb: uint32(size),
		ProviderDetails: map[string]*anypb.Any{
			GcloudProvider: details,
		},
	}, nil
}

func addImageToProject(p *common.Project, img *common.Image) error {
	p.Images = append(p.Images, img)
	return nil
}

