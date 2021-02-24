package cache

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"math"
	common "nephomancy/common/resources"
	"nephomancy/dcs/resources"
)

func GetCost(db *sql.DB, p *common.Project) ([][]string, error) {
	costs := make([][]string, 0)
	sla := "Basic"
	if p.ProviderDetails != nil && p.ProviderDetails[resources.DcsProvider] != nil {
		var dcsProject resources.DcsProject
		err := ptypes.UnmarshalAny(p.ProviderDetails[resources.DcsProvider], &dcsProject)
		if err != nil {
			return nil, err
		}
		sla = dcsProject.Sla
	}
	for _, vmset := range p.InstanceSets {
		if vmset.Template.ProviderDetails == nil || vmset.Template.ProviderDetails[resources.DcsProvider] == nil {
			return nil, fmt.Errorf("Missing %s provider details for instance set %s\n",
				resources.DcsProvider, vmset.Name)
		}
		var dcsVm resources.DcsVM
		err := ptypes.UnmarshalAny(vmset.Template.ProviderDetails[resources.DcsProvider], &dcsVm)
		if err != nil {
			return nil, err
		}
		vmcosts, err := vmCostRange(db, sla, *vmset, dcsVm)
		if err != nil {
			return nil, err
		}
		for _, vc := range vmcosts {
			vc = append([]string{p.Name, resources.DcsProvider, vmset.Name}, vc...)
		}
		costs = append(costs, vmcosts...)
	}
	for _, dset := range p.DiskSets {
		if dset.Template.ProviderDetails == nil || dset.Template.ProviderDetails[resources.DcsProvider] == nil {
			return nil, fmt.Errorf("Missing %s provider details for disk set %s\n",
				resources.DcsProvider, dset.Name)
		}
		var dcsDisk resources.DcsDisk
		err := ptypes.UnmarshalAny(dset.Template.ProviderDetails[resources.DcsProvider],
			&dcsDisk)
		if err != nil {
			return nil, err
		}
		dcosts, err := diskCostRange(db, sla, *dset, dcsDisk)
		if err != nil {
			return nil, err
		}
		for _, dc := range dcosts {
			dc = append([]string{p.Name, resources.DcsProvider, dset.Name}, dc...)
		}
		costs = append(costs, dcosts...)
	}
	for _, nw := range p.Networks {
		for _, gw := range nw.Gateways {
			// Type
			_ = gw
		}
	}
	return costs, nil
}

func diskCostRange(db *sql.DB, sla string, disk common.DiskSet, dcsDisk resources.DcsDisk) (
	[][]string, error) {
	dtype := dcsDisk.DiskType
	backup := dcsDisk.WithBackup
	backupInt := 0
	if backup {
		backupInt = 1
	}
	if dtype == "" {
		return nil, fmt.Errorf("Missing disk type information for disk set %s\n",
			disk.Name)
	}
	diskCount := disk.Count
	sizeGb := disk.Template.ActualSizeGb
	if sizeGb == 0 {
		sizeGb = uint64(disk.Template.Type.SizeGb)
	}
	priceDisk, err := executePriceQuery(db, "DiskCosts", sla,
		fmt.Sprintf(` AND DiskType="%s" AND Backup=%d `, dtype, backupInt))
	if err != nil {
		return nil, err
	}
	price := float64(priceDisk) / math.Pow(10, 9)
	spec := fmt.Sprintf("%s in %s",
		common.PrintDiskType(*disk.Template.Type),
		common.PrintLocation(*disk.Template.Location))
	const hoursPerMonth = 24 * 30
	expectedHours := hoursPerMonth * disk.PercentUsedAvg / 100
	costs := make([][]string, 1)
	costs[0] = []string{
		"Disk",
		fmt.Sprintf("%d", diskCount),
		spec,
		fmt.Sprintf("%d GB for %d h per month", sizeGb, hoursPerMonth),
		fmt.Sprintf("%.2f CHF", price*float64(hoursPerMonth)),
		fmt.Sprintf("%d GB for %d h per month", sizeGb, expectedHours),
		fmt.Sprintf("%.2f CHF", price*float64(expectedHours)),
	}
	return costs, nil
}

func vmCostRange(db *sql.DB, sla string, vm common.InstanceSet, dcsvm resources.DcsVM) (
	[][]string, error) {
	lic := dcsvm.OsLicense
	if lic == "" {
		return nil, fmt.Errorf("Missing os license information for instance set %s\n",
			vm.Name)
	}
	cpuCount := vm.Template.Type.CpuCount
	memoryGb := vm.Template.Type.MemoryGb
	usage := vm.UsageHoursPerMonth
	vmCount := vm.Count
	maxCpuUsage := uint32(30 * 24 * cpuCount * vmCount)
	projectedCpuUsage := uint32(usage * cpuCount * vmCount)
	maxMemoryUsage := uint32(30 * 24 * vmCount * memoryGb)
	projectedMemoryUsage := uint32(usage * vmCount * memoryGb)

	spec := fmt.Sprintf("%s in %s",
		common.PrintMachineType(*vm.Template.Type),
		common.PrintLocation(*vm.Template.Location))

	priceCpu, err := executePriceQuery(db, "CpuCosts", sla, "")
	if err != nil {
		return nil, err
	}
	priceMem, err := executePriceQuery(db, "MemoryCosts", sla, "")
	if err != nil {
		return nil, err
	}
	priceOs, err := executePriceQuery(db, "OSCosts", sla, fmt.Sprintf(` AND Vendor="%s"`, lic))
	if err != nil {
		return nil, err
	}
	cpu := float64(priceCpu) / math.Pow(10, 9)
	maxCpu := cpu * float64(maxCpuUsage)
	expCpu := cpu * float64(projectedCpuUsage)
	mem := float64(priceMem) / math.Pow(10, 9)
	maxMem := mem * float64(maxMemoryUsage)
	expMem := mem * float64(projectedMemoryUsage)
	os := float64(priceOs) / math.Pow(10, 9)
	maxOs := os * float64(30*24*vmCount)
	expOs := os * float64(usage*vmCount)
	costs := make([][]string, 3)

	costs[0] = []string{
		"VM CPU",
		fmt.Sprintf("%d", vmCount),
		spec,
		fmt.Sprintf("%d h per month", maxCpuUsage),
		fmt.Sprintf("%.2f CHF", maxCpu),
		fmt.Sprintf("%d h per month", projectedCpuUsage),
		fmt.Sprintf("%.2f CHF", expCpu),
	}
	costs[1] = []string{
		"VM RAM",
		fmt.Sprintf("%d", memoryGb),
		spec,
		fmt.Sprintf("%d GB-hours per month", maxMemoryUsage),
		fmt.Sprintf("%.2f CHF", maxMem),
		fmt.Sprintf("%d GB-hours per month", projectedMemoryUsage),
		fmt.Sprintf("%.2f CHF", expMem),
	}
	costs[2] = []string{
		"VM OS",
		fmt.Sprintf("%s", lic),
		spec,
		fmt.Sprintf("%d h per month", 30*24*vmCount),
		fmt.Sprintf("%.2f CHF", maxOs),
		fmt.Sprintf("%d h per month", usage*vmCount),
		fmt.Sprintf("%.2f CHF", expOs),
	}

	return costs, nil
}

func executePriceQuery(db *sql.DB, table string, sla string, q string) (uint32, error) {
	query := fmt.Sprintf(`SELECT Nanos from %s WHERE SLA="%s" %s;`, table, sla, q)
	fmt.Printf("query: %s\n", query)
	res, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer res.Close()
	var nanos uint32
	for res.Next() {
		err = res.Scan(&nanos)
		if err != nil {
			return 0, fmt.Errorf("error scanning row: %v", err)
		}
		return nanos, nil
	}
	return 0, fmt.Errorf("no nanos retrieved for query %s", query)
}
