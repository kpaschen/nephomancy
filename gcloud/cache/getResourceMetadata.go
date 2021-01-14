package cache

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	_ "github.com/mattn/go-sqlite3"
)

func GetResourceMetadata(db *sql.DB, asset *assets.SmallAsset) (*assets.ResourceMetadata, error) {
	rf, _ := asset.ResourceFamily()
	switch rf {
	case "Compute":
		mt, err := asset.MachineType()
		if err != nil {
			return nil, err
		}
		mtd, err := getMachineType(db, mt)
		if err != nil {
			return nil, err
		}
		return &assets.ResourceMetadata{
			Mt: mtd,
		}, nil
	case "Storage":
		dt, err := asset.BaseType()
		if err != nil {
			return nil, err
		}
		region := ""
		if dt == "RegionDisk" {
			regions, err := asset.Regions()
			if err != nil {
				return nil, err
			}
			if len(regions) > 0 {
				region = regions[0]
			}
		} else if dt != "Disk" {
			return nil, nil  // probably an Image. TODO: other storage assets?
		}
		tp, err := asset.DiskType()
		if err != nil {
			return nil, err
		}
		mtd, err := getDiskType(db, tp, region)
		if err != nil {
			return nil, err
		}
		return &assets.ResourceMetadata{
			Dt: mtd,
		}, nil
	default:
		return nil, nil
	}
	return nil, nil
}

func getMachineType(db *sql.DB, mt string) (assets.MachineType, error) {
	queryMachineType := fmt.Sprintf(`SELECT CpuCount, MemoryMb, IsSharedCpu
	FROM MachineTypes where MachineType='%s';`, mt)
	fmt.Printf("mt query: %s\n", queryMachineType)
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
	fmt.Printf("dt query: %s\n", queryDiskType)
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
