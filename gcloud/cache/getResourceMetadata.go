package cache

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	_ "github.com/mattn/go-sqlite3"
)

func AddResourceTypesToAssets(db *sql.DB, ax *assets.AssetStructure) error {
	for _, inst := range ax.Instances {
		mt, err := getMachineType(db, inst.MachineTypeName)
		if err != nil {
			return err
		}
		inst.MachineType = mt
	}
	for _, dsk := range ax.Disks {
		region := ""
		if dsk.IsRegional {
			region = dsk.ZoneOrRegion
		}
		dt, err := getDiskType(db, dsk.DiskTypeName, region)
		if err != nil {
			return err
		}
		dsk.DiskType = dt
	}
	return nil
}

func getMachineType(db *sql.DB, mt string) (assets.MachineType, error) {
	queryMachineType := fmt.Sprintf(`SELECT CpuCount, MemoryMb, IsSharedCpu
	FROM MachineTypes where MachineType='%s';`, mt)
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
			Name: mt,
			CpuCount: cpuCount,
			MemoryMb: memoryMb,
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
			Name: dt,
			DefaultSizeGb: defaultSizeGb,
			Region: region,  // is this right?

		}, nil
	}
	return assets.DiskType{}, fmt.Errorf("Failed to find disk type %s in region %s\n", dt, region)
}
