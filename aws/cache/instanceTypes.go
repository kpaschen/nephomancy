package cache

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	common "nephomancy/common/resources"
	"strings"
)

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

	fmt.Printf("query: %s\n", queryMachineType)

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
