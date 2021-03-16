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
		theMap, ok := rm["data"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected resource[data] to be another map but it is a %T",
				rm["data"])
		}
		a.resourceMap = theMap
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
	disks, ok := a.resourceMap["disks"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected disks to be a list of objects but it is a %T",
			a.resourceMap["disks"])
	}
	licenses := make([]string, 0)
	for _, d := range disks {
		diskmap, ok := d.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("disk list entry is a %T", d)
		}
		ls, ok := diskmap["licenses"].([]interface{})
		if !ok {
			continue // non-boot disks may not have a license
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
	disks, ok := a.resourceMap["disks"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected disks to be a list of objects but it is a %T",
			a.resourceMap["disks"])
	}
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
			// ubuntu-1604-xenial
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
	machineType, ok := a.resourceMap["machineType"].(string)
	if !ok {
		return "", fmt.Errorf("expected machine type to be a string but it is a %T",
			a.resourceMap["machineType"])
	}
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
	diskType, ok := a.resourceMap["type"].(string)
	if !ok {
		return "", fmt.Errorf("expected disk type to be a string but it is a %T",
			a.resourceMap["type"])
	}
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
		zone, ok := a.resourceMap["zone"].(string)
		if !ok {
			return nil, fmt.Errorf("expected zone to be a string but it is a %T", a.resourceMap["zone"])
		}
		path := strings.Split(zone, "/")
		z := path[len(path)-1]
		// Now z is of the form <continent>-<direction><integer>-<char> like
		// europe-west1-b. The region we need is z without the trailing char, so
		// europe-west1. The other region value is 'global' but there is no zone for that
		// in the resources afaik.
		regions = append(regions, z[:len(z)-2])
	} else if a.resourceMap["storageLocations"] != nil {
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
	return regions, nil
}

func (a *SmallAsset) networkName() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["network"] == nil {
		return "None", nil
	}
	nw, ok := a.resourceMap["network"].(string)
	if !ok {
		return "None", fmt.Errorf("network entry was a %T not a string",
			a.resourceMap["network"])
	}
	parts := strings.Split(nw, "/")
	return parts[len(parts)-1], nil
}

func (a *SmallAsset) serviceAccountName() (string, error) {
	if err := a.ensureResourceMap(); err != nil {
		return "", err
	}
	if a.resourceMap["name"] == nil {
		return "None", nil
	}
	n, ok := a.resourceMap["name"].(string)
	if !ok {
		return "None", fmt.Errorf("name was a %T not a string",
			a.resourceMap["network"])
	}
	parts := strings.Split(n, "/")
	// service account names have the form projects/<proj name>/serviceAccounts/<email>
	// keys look like an account name with "keys/<some uuid>" appended
	if len(parts) < 4 {
		return "None", fmt.Errorf("unexpected name format for service account or key: %s", n)
	}
	return parts[3], nil
}
