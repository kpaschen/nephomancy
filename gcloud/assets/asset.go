package assets

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

// This is a reduced version of the asset proto
type SmallAsset struct {
        Name string
        AssetType string
        ResourceAsJson string
	ResourceMap map[string]interface{}  // parsed version of ResourceAsJson
}

func (a *SmallAsset) StorageSize() (int64, error) {
	if err := a.ensureResourceMap(); err != nil {
                return 0, err
        }
	var diskSize int64
	abytes, ok := a.ResourceMap["archiveSizeBytes"].(string)
	if ok {
		diskSize, _ = strconv.ParseInt(abytes, 10, 64)
		fmt.Printf("disk size parsed: %d\n", diskSize)
		// The archive size gets reported as 4419062592 for a 4.12 GB image.
		diskSize = diskSize / (1024 * 1024 * 1024)  // Should this be 1000?
		fmt.Printf("disk size adjusted to gb: %d\n", diskSize)
		// should probably multiply this by number of storage locations?
	} else {
		gbytes, ok := a.ResourceMap["sizeGb"].(string)
		if ok {
			diskSize, _ =  strconv.ParseInt(gbytes, 10, 64)
		} else {
			return 0, fmt.Errorf("Unable to determine storage size for asset %+v\n", a)
		}
	}
	return diskSize, nil
}

func (a *SmallAsset) BillingService() (string, error) {
	parts := strings.Split(a.AssetType, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected service/resource format for asset type but got %s\n", a.AssetType)
	}

	switch parts[0] {
	case "compute.googleapis.com":
		return "6F81-5844-456A", nil
	case "container.googleapis.com":
		return "CCD8-9BF1-090E", nil
	case "monitoring.googleapis.com":
		return "58CD-E7C3-72CA", nil
	case "cloudresourcemanager.googleapis.com":
		return "", nil  // This is the project resource, not sure what else?
        case "iam.googleapis.com":
		return "", nil  // ServiceAccountKey, ServiceAccount, what else?
	case "serviceusage.googleapis.com":
		return "", nil  // TODO: services, some of them get a charge
	default:
		log.Printf("No billing service configured for asset type %s, api %s?\n", a.AssetType, parts[0])
		return "", nil
	}
	return "", fmt.Errorf("Reached part after switch statement unexpectedly.\n")
}

func (a *SmallAsset) BaseType() (string, error) {
	parts := strings.Split(a.AssetType, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected service/resource format for asset type but got %s\n", a.AssetType)
	}
	return parts[1], nil
}

func (a *SmallAsset) ResourceFamily() (string, error) {
	tp, err := a.BaseType()
	if err != nil {
		return "", err
	}
	switch tp {
	case "Route":
		return "Network", nil
	case "Network":
		return "Network", nil
	case "Subnetwork":
		return "Network", nil
	case "Firewall":
		return "Network", nil
	case "Instance":
		return "Compute", nil  // also return License here (and/or for image)
	case "Image":
		return "Storage", nil
	case "Disk":
		return "Storage", nil
	case "RegionDisk":
		return "Storage", nil
	case "Project":
		return "", nil
	default:
		log.Printf("No resource family known for %s\n", tp)
		return "", nil
	}
}

func (a *SmallAsset) ensureResourceMap() error {
	if a.ResourceMap == nil {
		rBytes := []byte(a.ResourceAsJson)
		var rm map[string]interface{}
		json.Unmarshal(rBytes, &rm)
		theMap, ok := rm["data"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected resource[data] to be another map but it is a %T\n",
			rm["data"])
		}
		a.ResourceMap = theMap
	}
	return nil
}

func (a *SmallAsset) Scheduling() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.ResourceMap["scheduling"] != nil {
                scheduling, ok := a.ResourceMap["scheduling"].(map[string]interface{})
                if ok {
                        preempt, ok := scheduling["preemptible"].(bool)
                        // TODO: also support Commit1Yr etc
                        if ok {
                                if preempt {
					return "Preemptible", nil
                                } else {
					return "OnDemand", nil
                                }
                        }
                }
        }
	return "", nil
}

func (a *SmallAsset) MachineType() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.ResourceMap["machineType"] == nil {
		return "", nil
	}
	machineType, ok := a.ResourceMap["machineType"].(string)
        if !ok {
	        return "", fmt.Errorf("expected machine type to be a string but it is a %T\n",
		a.ResourceMap["machineType"])
        }
        u, err := url.Parse(machineType)
        if err != nil {
                return "", err
        }
        path := strings.Split(u.Path, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) DiskType() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.ResourceMap["type"] == nil {
		return "", nil
	}
	diskType, ok := a.ResourceMap["type"].(string)
	if !ok {
		return "", fmt.Errorf("Expected disk type to be a string but it is a %T\n",
		a.ResourceMap["type"])
	}
	u, err := url.Parse(diskType)
	if err != nil {
		return "", err
	}
	path := strings.Split(u.Path, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) Zone() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.ResourceMap["zone"] == nil {
		return "None", nil
	}
	zone, _ := a.ResourceMap["zone"].(string)
	path := strings.Split(zone, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) Regions() ([]string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	regions := make([]string, 0)
	if a.ResourceMap["zone"] != nil {
		zone, ok := a.ResourceMap["zone"].(string)
		if !ok {
			return nil, fmt.Errorf("expected zone to be a string but it is a %T\n", a.ResourceMap["zone"])
		}
		u, err := url.Parse(zone)
		if err != nil {
			return nil, err
		}
		path := strings.Split(u.Path, "/")
		z := path[len(path)-1]
		// Now z is of the form <continent>-<direction><integer>-<char> like
                // europe-west1-b. The region we need is z without the trailing char, so
                // europe-west1. The other region value is 'global' but there is no zone for that
                // in the resources afaik.
                regions = append(regions, z[:len(z)-2])
	} else {
                if a.ResourceMap["storageLocations"] != nil {
                        loc, ok := a.ResourceMap["storageLocations"].([]interface{})
                        if !ok {
                                fmt.Printf("expected sl to be a string array but it is a %T\n",
                                a.ResourceMap["storageLocations"])
                                return nil, nil
                        }
                        for _, l := range loc {
                                r, ok := l.(string)
                                if !ok {
                                        fmt.Printf("expected l to be a string but it is a %T\n", l)
                                        return nil, nil
                                }
				regions = append(regions, r)
                        }
                }

	}
	return regions, nil
}

func (a *SmallAsset) Networks() ([]string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	if a.ResourceMap["networkInterfaces"] == nil {
		return nil, nil
	}
	nwis, _ := a.ResourceMap["networkInterfaces"].([]interface{})
	ret := make([]string, len(nwis))
	for idx, nwi := range nwis {
		nwix, _ := nwi.(map[string]interface{})
		nic, _ := nwix["name"].(string)
		x, _ := nwix["network"].(string)
		parts := strings.Split(x, "/")
		nwName := parts[len(parts)-1]
		x, _ = nwix["subnetwork"].(string)
		parts = strings.Split(x, "/")
		snwName := parts[len(parts)-1]
		ret[idx] = fmt.Sprintf("%s:%s:%s", nic, nwName, snwName)
	}
	return ret, nil
}

func (a *SmallAsset) NetworkName() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.ResourceMap["network"] == nil {
		return "None", nil
	}
	nw, ok := a.ResourceMap["network"].(string)
	if !ok {
		return "None", fmt.Errorf("network entry was a %T not a string\n",
		a.ResourceMap["network"])
	}
	parts := strings.Split(nw, "/")
	return parts[len(parts)-1], nil
}

func (a *SmallAsset) ServiceAccountName() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.ResourceMap["name"] == nil {
		return "None", nil
	}
	n, ok := a.ResourceMap["name"].(string)
	if !ok {
		return "None", fmt.Errorf("name was a %T not a string\n",
		a.ResourceMap["network"])
	}
	parts := strings.Split(n, "/")
	// service account names have the form projects/<proj name>/serviceAccounts/<email>
	// keys look like an account name with "keys/<some uuid>" appended
	if len(parts) < 4 {
		return "None", fmt.Errorf("unexpected name format for service account or key: %s\n", n)
	}
	return parts[3], nil
}

func (a *SmallAsset) PrettyPrint() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	tp, err := a.BaseType()
        if err != nil {
                return "", err
        }
        switch tp {
        case "Route":
		return a.prettyPrintRoute()
        case "Network":
                return a.prettyPrintNetwork()
        case "Subnetwork":
                return a.prettyPrintSubnetwork()
        case "Firewall":
                return a.prettyPrintFirewall()
        case "Instance":
                return a.prettyPrintInstance()
        case "Image":
                return a.prettyPrintImage()
        case "Disk":
                return a.prettyPrintDisk()
        case "RegionDisk":
                return a.prettyPrintRegionDisk()
        case "Project":
                return "Project", nil
	case "ServiceAccountKey":
		return a.prettyPrintServiceAccountKey()
	case "ServiceAccount":
		return a.prettyPrintServiceAccount()
	case "Service":
		return a.prettyPrintService()
        default:
                log.Printf("No idea how to pretty-print %+v\n", a)
                return tp, nil
        }
}

func (a *SmallAsset) prettyPrintImage() (string, error) {
	idParts := strings.Split(a.Name, "/")
	name := idParts[len(idParts)-1]
	status, _ := a.ResourceMap["status"].(string)
	regions, _ := a.Regions()
	size, _ := a.StorageSize()
	lcs := make([]string, 0)
	if a.ResourceMap["licenses"] != nil {
		licenses, _ := a.ResourceMap["licenses"].([]interface{})
		for _, l := range licenses {
			lic, _ := l.(string)
			lcs = append(lcs, lic)
		}
	}
	licenses := fmt.Sprintf("%v", lcs)
	return fmt.Sprintf(`%s is an Image with size %d Gb in region(s) %s.
Its current status is %s and the applicable licenses are %s`,
	name, size, regions, status, licenses), nil
}

func (a *SmallAsset) prettyPrintDisk() (string, error) {
	idParts := strings.Split(a.Name, "/")
	name := idParts[len(idParts)-1]
	diskType, _ := a.DiskType()
	zone, _ := a.Zone()
	status, _ := a.ResourceMap["status"].(string)
	size, _ := a.StorageSize()

	return fmt.Sprintf(`%s is a zonal disk of type %s in zone %s.
Its current status is %s and its size is %d Gb.`,
	name, diskType, zone, status, size), nil
}

func (a *SmallAsset) prettyPrintRegionDisk() (string, error) {
	idParts := strings.Split(a.Name, "/")
	name := idParts[len(idParts)-1]
	diskType, _ := a.DiskType()
	regions, _ := a.Regions()
	status, _ := a.ResourceMap["status"].(string)
	size, _ := a.StorageSize()

	return fmt.Sprintf(`%s is a region disk of type %s in region(s) %s.
Its current status is %s and its size is %d Gb.`,
	name, diskType, regions, status, size), nil
}

func (a *SmallAsset) prettyPrintInstance() (string, error) {
	idParts := strings.Split(a.Name, "/")
	name := idParts[len(idParts)-1]
	scheduling, _ := a.Scheduling()
	status, _ := a.ResourceMap["status"].(string)
	machineType, _ := a.MachineType()
	regions, _ := a.Regions()
	networks, _ := a.Networks()

	return fmt.Sprintf(`%s is an instance of type %s in region(s) %s.
Its current status is %s. It uses %s scheduling.
It is on network %+v`,
	name, machineType, regions, status, scheduling, networks), nil
}

func (a *SmallAsset) prettyPrintRoute() (string, error) {
	name, _ := a.ResourceMap["name"].(string)
	description, _ := a.ResourceMap["description"].(string)
	return fmt.Sprintf("route %s: %s\n", name, description), nil
}

func (a *SmallAsset) prettyPrintNetwork() (string, error) {
	name, _ := a.ResourceMap["name"].(string)
	routingConfig, _ := a.ResourceMap["routingConfig"].(map[string]interface{})
	routingMode, _ := routingConfig["routingMode"].(string)
	subnetworks, _ := a.ResourceMap["subnetworks"].([]interface{})
	subnw := make([]string, len(subnetworks))
	for idx, snw := range subnetworks {
		s, _ := snw.(string)
		parts := strings.Split(s, "/")
		snwName := parts[len(parts)-1]
		region := parts[len(parts)-3]
		subnw[idx] = fmt.Sprintf("%s:%s", region, snwName)
	}
	return fmt.Sprintf("network %s with routing mode %s and subnetworks %v\n",
	name, routingMode, subnw), nil
}

func (a *SmallAsset) prettyPrintSubnetwork() (string, error) {
	name, _ := a.ResourceMap["name"].(string)
	purpose, _ := a.ResourceMap["purpose"].(string)
	r, _ := a.ResourceMap["region"].(string)
	p := strings.Split(r, "/")
	region := p[len(p)-1]
	ipCidrRange := a.ResourceMap["ipCidrRange"].(string)
	return fmt.Sprintf("Subnetwork %s: %s in region %s with ip range %s\n",
	name, purpose, region, ipCidrRange), nil
}

func (a *SmallAsset) prettyPrintFirewall() (string, error) {
	name, _ := a.ResourceMap["name"].(string)
	description, _ := a.ResourceMap["description"].(string)
	network, _ := a.ResourceMap["network"].(string)
	nwParts := strings.Split(network, "/")
	networkName := nwParts[len(nwParts)-1]
	return fmt.Sprintf("Firewall %s: %s on network %s\n", name, description, networkName), nil
}

func (a *SmallAsset) prettyPrintService() (string, error) {
	name, _ := a.ResourceMap["name"].(string)
	state, _ := a.ResourceMap["state"].(string)
	return fmt.Sprintf("Service %s in state %s\n", name, state), nil
}

func (a *SmallAsset) prettyPrintServiceAccount() (string, error) {
	displayname, _ := a.ResourceMap["displayName"].(string)
	description, _ := a.ResourceMap["description"].(string)
	email, _ := a.ResourceMap["email"].(string)
	return fmt.Sprintf("ServiceAccount %s: %s (%s)\n", email, displayname, description), nil
}

func (a *SmallAsset) prettyPrintServiceAccountKey() (string, error) {
	keytype, _ := a.ResourceMap["keyType"].(string)
	name, _ := a.ResourceMap["name"].(string)
	parts := strings.Split(name, "/")
	serviceAccountEmail := parts[len(parts)-3]
	return fmt.Sprintf("ServiceAccountKey for account %s has type %s\n", serviceAccountEmail, keytype), nil
}
