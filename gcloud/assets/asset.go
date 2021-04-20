package assets

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// This is a reduced version of the asset proto
type SmallAsset struct {
	Name           string
	AssetType      string
	ResourceAsJson string
	resourceMap    map[string]interface{} // parsed version of ResourceAsJson
}

func (a *SmallAsset) storageSize() (int64, error) {
	if err := a.ensureResourceMap(); err != nil {
		return 0, err
	}
	var diskSize int64
	abytes, ok := a.resourceMap["archiveSizeBytes"].(string)
	if ok {
		diskSize, _ = strconv.ParseInt(abytes, 10, 64)
		// The archive size gets reported as 4419062592 for a 4.12 GB image.
		diskSize = diskSize / (1024 * 1024 * 1024) // Should this be 1000?
	} else {
		gbytes, ok := a.resourceMap["sizeGb"].(string)
		if ok {
			diskSize, _ = strconv.ParseInt(gbytes, 10, 64)
		} else {
			return 0, fmt.Errorf("unable to determine storage size for asset %+v", a)
		}
	}
	return diskSize, nil
}

func (a *SmallAsset) BaseType() (string, error) {
	parts := strings.Split(a.AssetType, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected service/resource format for asset type but got %s", a.AssetType)
	}
	return parts[1], nil
}

func (a *SmallAsset) ensureResourceMap() error {
	if a.resourceMap == nil {
		rBytes := []byte(a.ResourceAsJson)
		var rm map[string]interface{}
		json.Unmarshal(rBytes, &rm)
		if rm["data"] != nil {
			theMap, _ := rm["data"].(map[string]interface{})
			a.resourceMap = theMap
		} else {
			return fmt.Errorf("asset %+v has nil resource map", a)
		}
	}
	return nil
}

func (a *SmallAsset) scheduling() (string, error) {
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
				}
				return "OnDemand", nil
			}
		}
	}
	return "", nil
}

// a is assumed to be an instance that has at least a boot disk.
func (a *SmallAsset) licenses() ([]string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	if a.resourceMap["disks"] == nil {
		return nil, nil
	}
	disks, _ := a.resourceMap["disks"].([]interface{})
	licenses := make([]string, 0)
	for _, d := range disks {
		diskmap, _ := d.(map[string]interface{})
		ls, ok := diskmap["licenses"].([]interface{})
		if !ok {
			continue // non-boot disks need not have a license
		}
		for _, l := range ls {
			lic, _ := l.(string)
			licenses = append(licenses, lic)
		}
	}
	return licenses, nil
}

func (a *SmallAsset) localDisks() ([]interface{}, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	if a.resourceMap["disks"] == nil {
		return nil, nil
	}
	disks, _ := a.resourceMap["disks"].([]interface{})
	ret := make([]interface{}, 0)
	for _, d := range disks {
		dk, _ := d.(map[string]interface{})
		if dk["type"] != "PERSISTENT" {
			ret = append(ret, dk)
		}
	}
	return ret, nil
}

func (a *SmallAsset) os() (string, error) {
	licenses, err := a.licenses()
	if err != nil {
		return "", err
	}
	for _, license := range licenses {
		parts := strings.Split(license, "/")
		if len(parts) > 0 {
			return OsFromLicenseName(parts[len(parts)-1]), nil
		}
	}
	return "", fmt.Errorf("no os found")
}

func (a *SmallAsset) machineType() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["machineType"] == nil {
		return "", nil
	}
	machineType, _ := a.resourceMap["machineType"].(string)
	path := strings.Split(machineType, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) diskType() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["type"] == nil {
		return "", nil
	}
	diskType, _ := a.resourceMap["type"].(string)
	path := strings.Split(diskType, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) zone() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["zone"] == nil {
		return "None", nil
	}
	zone, _ := a.resourceMap["zone"].(string)
	path := strings.Split(zone, "/")
	return path[len(path)-1], nil
}

func (a *SmallAsset) regions() ([]string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	regions := make([]string, 0)
	if a.resourceMap["region"] != nil {
		region, _ := a.resourceMap["region"].(string)
		path := strings.Split(region, "/")
		r := path[len(path)-1]
		regions = append(regions, r)
		// A regional disk will have 'region' and 'replicaZones' set, but the
		// replica zones will all be in the same region.
	} else if a.resourceMap["zone"] != nil {
		zone, _ := a.resourceMap["zone"].(string)
		path := strings.Split(zone, "/")
		z := path[len(path)-1]
		// Now z is of the form <continent>-<direction><integer>-<char> like
		// europe-west1-b. The region we need is z without the trailing char, so
		// europe-west1. The other region value is 'global' but there is no zone for that
		// in the resources afaik.
		regions = append(regions, z[:len(z)-2])
	} else if a.resourceMap["storageLocations"] != nil {
		loc, _ := a.resourceMap["storageLocations"].([]interface{})
		for _, l := range loc {
			r, _ := l.(string)
			regions = append(regions, r)
		}
	}
	return regions, nil
}

func (a *SmallAsset) networkName() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["network"] != nil {
		nw, _ := a.resourceMap["network"].(string)
		parts := strings.Split(nw, "/")
		return parts[len(parts)-1], nil
	} else if a.resourceMap["networkInterfaces"] != nil {
		nwif, _ := a.resourceMap["networkInterfaces"].([]interface{})
		for _, interf := range nwif {
			nwInterface, _ := interf.(map[string]interface{})
			nwname := nwInterface["network"].(string)
			parts := strings.Split(nwname, "/")
			return parts[len(parts)-1], nil
		}
	}
	return "None", nil
}

// This is for ip addresses attached to a VM.
// If this is a permanent address, there will also be a separate asset entry
// for it. The list generated here is just for ensuring that ephemeral IP addresses
// are also counted.
func (a *SmallAsset) ipAddr() ([]string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return nil, err
	}
	if a.resourceMap["networkInterfaces"] == nil {
		return nil, nil
	}
	nwif, _ := a.resourceMap["networkInterfaces"].([]interface{})
	ret := make([]string, 0)
	for _, interf := range nwif {
		nwInterface, _ := interf.(map[string]interface{})
		accessConfigs, _ := nwInterface["accessConfigs"].([]interface{})
		for _, ac := range accessConfigs {
			config, _ := ac.(map[string](interface{}))
			tp, _ := config["type"].(string)
			// Also verify that name == "External NAT"?
			// I think this type is the only one that signals an actual external IP, but maybe I'm wrong.
			if tp == "ONE_TO_ONE_NAT" {
				ipAddr, _ := config["natIP"].(string)
				if ipAddr != "" {
					ret = append(ret, ipAddr)
				}
			}
		}
	}
	return ret, nil
}

func (a *SmallAsset) serviceAccountName() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["name"] == nil {
		return "None", nil
	}
	n, _ := a.resourceMap["name"].(string)
	parts := strings.Split(n, "/")
	// service account names have the form projects/<proj name>/serviceAccounts/<email>
	// keys look like an account name with "keys/<some uuid>" appended
	if len(parts) < 4 {
		return "None", fmt.Errorf("unexpected name format for service account or key: %s", n)
	}
	return parts[3], nil
}
