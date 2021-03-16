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
			return nil, fmt.Errorf("missing %s provider details for instance set %s",
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
		nwcosts, err := networkCostRange(db, sla, *nw)
		if err != nil {
			return nil, err
		}
		costs = append(costs, nwcosts...)
	}
	return costs, nil
}

func networkCostRange(db *sql.DB, sla string, network common.Network) (
	[][]string, error) {
	ipAddrCount := network.IpAddresses
	bandwidthMBits := network.BandwidthMbits
	costs := make([][]string, 0)
	// the max number of IP Addresses is: 2^(32 - Cidr) - 5
	cidr := uint32(32 - math.Log2(float64(ipAddrCount+5)))
	fmt.Printf("cidr for %d ip addresses is %d\n", ipAddrCount, cidr)
	priceIPAddr, err := executePriceQuery(db, "IpAddrCosts", sla,
		fmt.Sprintf(` AND Cidr >= %d `, cidr))
	if err != nil {
		return nil, err
	}
	// This is the price for 10 MBit/s
	pricePer10MBits, err := executePriceQuery(db, "BandwidthCosts", sla, "")
	if err != nil {
		return nil, err
	}
	priceBandwidth := float64(pricePer10MBits) * math.Ceil(float64(bandwidthMBits)/10.0) / math.Pow(10, 9)
	hoursPerMonth := uint32(24 * 30)
	costs = append(costs, []string{
		"IP Addresses",
		fmt.Sprintf("%d", ipAddrCount),
		fmt.Sprintf("ip addresses: /%d cidr", cidr),
		fmt.Sprintf("%d addresses for %d h per month", ipAddrCount, hoursPerMonth),
		fmt.Sprintf("%.2f CHF", float64(priceIPAddr*hoursPerMonth)/math.Pow(10, 9)),
		fmt.Sprintf("%d addresses for %d h per month", ipAddrCount, hoursPerMonth),
		fmt.Sprintf("%.2f CHF", float64(priceIPAddr*hoursPerMonth)/math.Pow(10, 9)),
	})
	costs = append(costs, []string{
		"Bandwidth",
		fmt.Sprintf("%d", bandwidthMBits),
		fmt.Sprintf("Bandwidth in MBit/s"),
		fmt.Sprintf("%d MBit/s for %d h per month", bandwidthMBits, hoursPerMonth),
		fmt.Sprintf("%.2f CHF", priceBandwidth*float64(hoursPerMonth)),
		fmt.Sprintf("%d MBit/s for %d h per month", bandwidthMBits, hoursPerMonth),
		fmt.Sprintf("%.2f CHF", priceBandwidth*float64(hoursPerMonth)),
	})

	for _, gw := range network.Gateways {
		var dcsGw resources.DcsGateway
		err := ptypes.UnmarshalAny(gw.ProviderDetails[resources.DcsProvider], &dcsGw)
		if err != nil {
			return nil, err
		}
		gwType := dcsGw.Type
		if gwType == "" {
			gwType = "Eco" // Default, free.
		}
		priceGateway, err := executePriceQuery(db, "GatewayCosts", sla,
			fmt.Sprintf(` AND Type="%s" `, gwType))
		if err != nil {
			return nil, err
		}
		costs = append(costs, []string{
			"Gateway",
			"1",
			fmt.Sprintf("Gateway of type %s", gwType),
			fmt.Sprintf("for %d h per month", hoursPerMonth),
			fmt.Sprintf("%.2f CHF", float64(priceGateway*hoursPerMonth)/math.Pow(10, 9)),
			fmt.Sprintf("for %d h per month", hoursPerMonth),
			fmt.Sprintf("%.2f CHF", float64(priceGateway*hoursPerMonth)/math.Pow(10, 9)),
		})
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
		return nil, fmt.Errorf("missing disk type information for disk set %s",
			disk.Name)
	}
	diskCount := disk.Count
	sizeGb := uint64(disk.Template.Type.SizeGb)
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
	expectedHours := disk.UsageHoursPerMonth
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
	lic := dcsvm.OsChoice
	if lic == "" {
		return nil, fmt.Errorf("missing os license information for instance set %s",
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
