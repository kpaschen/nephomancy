package cache

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	common "nephomancy/common/resources"
)

// Resolve the provider-specific resource types into generic machine and
// disk types. Fails if resource types are not found or if the settings
// are inconsistent.
func AddResourceTypesToProject(db *sql.DB, p *common.Project) error {
	for _, vmset := range p.VmSets {
		var gvm assets.GCloudVM
		err := ptypes.UnmarshalAny(vmset.Template.ProviderDetails[assets.GcloudProvider], &gvm)
		if err != nil {
			return err
		}
		mt, err := getMachineType(db, gvm.MachineType)
		if err != nil {
			return err
		}
		// Should probably check if this was already set and
		// verify that it matches if so?
		vmset.Template.Type = &common.MachineType{
			CpuCount: mt.CpuCount,
			MemoryGb: uint32(mt.MemoryMb / 1000),
			GpuCount: 0,  // FIXME
		}
	}
	for _, dset := range p.DiskSets {
		var dsk assets.GCloudDisk
		err := ptypes.UnmarshalAny(dset.Template.ProviderDetails[assets.GcloudProvider], &dsk)
		if err != nil {
			return err
		}
		region := ""
		if dsk.IsRegional {
			region = dsk.Region
		}
		dt, err := getDiskType(db, dsk.DiskType, region)
		dset.Template.Type = &common.DiskType{
			SizeGb: uint32(dt.DefaultSizeGb),
		}
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
			CpuCount: uint32(cpuCount),
			MemoryMb: uint64(memoryMb),
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
