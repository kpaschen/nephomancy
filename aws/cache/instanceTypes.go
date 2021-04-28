package cache

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"nephomancy/aws/resources"
	common "nephomancy/common/resources"
	"strings"
)

func getInstanceTypeByName(db *sql.DB, itName string) (resources.InstanceType, error) {
	query := fmt.Sprintf(`SELECT it.InstanceType, it.Memory, it.NetworkPerformance,
	it.SupportsSpot, it.SupportsOnDemand, it.StorageAmount,
	it.StorageType, it.GPU, it.CPU FROM InstanceTypes it WHERE
	it.InstanceType='%s'`, itName)

	cur, err := db.Query(query)
	if err != nil {
		return resources.InstanceType{}, err
	}
	defer cur.Close()
	var name string
	var memory int
	var networkperf int
	var supportsSpot int
	var supportsOnDemand int
	var instanceStorageMaxSizeGb int
	var instanceStorageType string
	var gpuCount int
	var cpuCount int
	for cur.Next() {
		err = cur.Scan(&name, &memory, &networkperf, &supportsSpot, &supportsOnDemand,
			&instanceStorageMaxSizeGb, &instanceStorageType, &gpuCount,
			&cpuCount)
		if err != nil {
			return resources.InstanceType{}, err
		}
	}
	ret := resources.InstanceType{
		Name:                     name,
		MemoryMiB:                uint64(memory),
		NetworkPerformanceGbit:   uint32(networkperf),
		DefaultCpuCount:          uint32(cpuCount),
		GpuCount:                 uint32(gpuCount),
		InstanceStorageType:      instanceStorageType,
		InstanceStorageMaxSizeGb: uint64(instanceStorageMaxSizeGb),
	}
	if supportsSpot > 0 {
		if supportsOnDemand > 0 {
			ret.SupportedUsageClasses = []string{"on-demand", "spot"}
		} else {
			ret.SupportedUsageClasses = []string{"spot"}
		}
	} else {
		if supportsOnDemand > 0 {
			ret.SupportedUsageClasses = []string{"on-demand"}
		} else {
			ret.SupportedUsageClasses = []string{}
		}
	}
	if instanceStorageMaxSizeGb > 0 {
		ret.InstanceStorageSupported = true
	}

	queryValidCores := fmt.Sprintf(`SELECT c.CoreCount from CoreCount c WHERE
	c.InstanceType='%s`, itName)
	cur2, err := db.Query(queryValidCores)
	if err != nil {
		return resources.InstanceType{}, err
	}
	defer cur2.Close()
	var it string
	var coreCount int
	ret.ValidCores = make([]uint32, 0)
	for cur2.Next() {
		err = cur2.Scan(&it, &coreCount)
		if err != nil {
			return resources.InstanceType{}, err
		}
		ret.ValidCores = append(ret.ValidCores, uint32(coreCount))
	}

	return ret, nil
}

// Retrieves an instance type satisfying the spec and available in at
// least one of the regions provided. Returns the instance type and
// the list of matching regions where it is available.
// If several instance types match the spec, the smallest one is returned.
// If there are several smallest types, you get one of them.
// TODO: take more features into account, like scheduling.
// TODO: also look at gpu count and local storage support.
func getInstanceTypeForSpec(db *sql.DB, mt common.MachineType, r []string) (
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
	queryMachineType := fmt.Sprintf(`SELECT DISTINCT it.InstanceType, r.Region
	FROM InstanceTypes it join InstanceTypeByRegion r ON
	it.InstanceType=r.InstanceType
	JOIN CoreCount c on it.InstanceType=c.InstanceType
	WHERE r.Region in %s AND c.CoreCount >= %d AND c.CoreCount <= %d
	AND it.Memory >= %d AND it.Memory <= %d
	ORDER BY c.CoreCount ASC, it.Memory ASC, it.InstanceType ASC LIMIT 1;`,
		regionsClause.String(), mt.CpuCount, mt.CpuCount*2,
		mt.MemoryGb*1000, mt.MemoryGb*2000)

	res, err := db.Query(queryMachineType)
	if err != nil {
		return "", []string{}, err
	}
	defer res.Close()
	var it string
	var reg string
	for res.Next() {
		err = res.Scan(&it, &reg)
		if err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}
	}
	if it != "" {
		return it, []string{reg}, nil
	}
	return "", nil, fmt.Errorf("Failed to find a suitable machine type for %v in %v", mt, r)
}
