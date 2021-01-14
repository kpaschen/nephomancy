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
	resourceMap map[string]interface{}  // parsed version of ResourceAsJson
}

func (a *SmallAsset) StorageSize() (int64, error) {
	if err := a.ensureResourceMap(); err != nil {
                return 0, err
        }
	var diskSize int64
	abytes, ok := a.resourceMap["archiveSizeBytes"].(string)
	if ok {
		diskSize, _ = strconv.ParseInt(abytes, 10, 64)
		fmt.Printf("disk size parsed: %d\n", diskSize)
		// The archive size gets reported as 4419062592 for a 4.12 GB image.
		diskSize = diskSize / (1024 * 1024 * 1024)  // Should this be 1000?
		fmt.Printf("disk size adjusted to gb: %d\n", diskSize)
		// should probably multiply this by number of storage locations?
	} else {
		gbytes, ok := a.resourceMap["sizeGb"].(string)
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
	if a.resourceMap == nil {
		rBytes := []byte(a.ResourceAsJson)
		var rm map[string]interface{}
		json.Unmarshal(rBytes, &rm)
		theMap, ok := rm["data"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected resource[data] to be another map but it is a %T\n",
			rm["data"])
		}
		a.resourceMap = theMap
	}
	return nil
}

func (a *SmallAsset) Scheduling() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["scheduling"] != nil {
                scheduling, ok := a.resourceMap["scheduling"].(map[string]interface{})
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
	if a.resourceMap["machineType"] == nil {
		return "", nil
	}
	machineType, ok := a.resourceMap["machineType"].(string)
        if !ok {
	        return "", fmt.Errorf("expected machine type to be a string but it is a %T\n",
		a.resourceMap["machineType"])
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
	if a.resourceMap["type"] == nil {
		return "", nil
	}
	diskType, ok := a.resourceMap["type"].(string)
	if !ok {
		return "", fmt.Errorf("Expected disk type to be a string but it is a %T\n",
		a.resourceMap["type"])
	}
	u, err := url.Parse(diskType)
	if err != nil {
		return "", err
	}
	path := strings.Split(u.Path, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) Regions() ([]string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	regions := make([]string, 0)
	if a.resourceMap["zone"] != nil {
		zone, ok := a.resourceMap["zone"].(string)
		if !ok {
			return nil, fmt.Errorf("expected zone to be a string but it is a %T\n", a.resourceMap["zone"])
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
                if a.resourceMap["storageLocations"] != nil {
                        loc, ok := a.resourceMap["storageLocations"].([]interface{})
                        if !ok {
                                fmt.Printf("expected sl to be a string array but it is a %T\n",
                                a.resourceMap["storageLocations"])
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
