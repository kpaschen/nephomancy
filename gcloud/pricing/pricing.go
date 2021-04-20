package pricing

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"math"
	common "nephomancy/common/resources"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
)

func GetCost(db *sql.DB, p *common.Project) ([][]string, error) {
	costs := make([][]string, 0)
	for _, vmset := range p.InstanceSets {
		var gvm assets.GCloudVM
		if err := ptypes.UnmarshalAny(
			vmset.Template.ProviderDetails[assets.GcloudProvider], &gvm); err != nil {
			return nil, err
		}
		skus, _ := cache.GetSkusForInstance(db, gvm)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return nil, err
		}
		vmcosts, err := vmCostRange(db, *vmset, gvm, pi)
		if err != nil {
			return nil, err
		}
		for idx, vc := range vmcosts {
			vmcosts[idx] = append([]string{
				p.Name, assets.GcloudProvider, vmset.Name}, vc...)
		}
		costs = append(costs, vmcosts...)

		skus, _ = cache.GetSkusForLicense(db, gvm)
		pi, err = cache.GetPricingInfo(db, skus)
		if err != nil {
			return nil, err
		}
		licenseCosts, err := licenseCost(db, *vmset, gvm, pi)
		if err != nil {
			return nil, err
		}
		for idx, lc := range licenseCosts {
			licenseCosts[idx] = append([]string{
				p.Name, assets.GcloudProvider, vmset.Name},
				lc...)
		}
		costs = append(costs, licenseCosts...)

		if vmset.Template.LocalStorage != nil && len(vmset.Template.LocalStorage) > 0 {
			skus, _ := cache.GetSkusForLocalDisk(db, gvm)
			pi, err := cache.GetPricingInfo(db, skus)
			if err != nil {
				return nil, err
			}
			localDiskCosts, err := localDiskCost(db, *vmset, gvm, pi)
			if err != nil {
				return nil, err
			}
			localDiskCosts = append([]string{
				p.Name, assets.GcloudProvider, vmset.Name}, localDiskCosts...)
			costs = append(costs, localDiskCosts)
		}
	}
	for _, dset := range p.DiskSets {
		var gdsk assets.GCloudDisk
		if err := ptypes.UnmarshalAny(
			dset.Template.ProviderDetails[assets.GcloudProvider], &gdsk); err != nil {
			return nil, err
		}
		skus, _ := cache.GetSkusForDisk(db, gdsk)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return nil, err
		}
		dcosts, err := diskCostRange(db, *dset, gdsk, pi)
		if err != nil {
			return nil, err
		}
		dcosts = append([]string{
			p.Name, assets.GcloudProvider, dset.Name}, dcosts...)
		costs = append(costs, dcosts)

		if dset.Template.Image != nil {
			skus, _ := cache.GetSkusForImage(db, gdsk)
			pi, err := cache.GetPricingInfo(db, skus)
			if err != nil {
				return nil, err
			}
			icosts, err := imageCost(db, *dset.Template.Image, pi)
			if err != nil {
				return nil, err
			}
			icosts = append([]string{
				p.Name, assets.GcloudProvider, dset.Template.Image.Name},
				icosts...)
			costs = append(costs, icosts)
		}
	}
	for _, nw := range p.Networks {
		tier, _ := assets.NetworkTier(*nw)
		var gnw assets.GCloudNetwork
		if err := ptypes.UnmarshalAny(nw.ProviderDetails[assets.GcloudProvider], &gnw); err != nil {
			return nil, err
		}
		for _, addr := range gnw.Addresses {
			if addr.Type != "EXTERNAL" {
				continue
			}
			var region string
			if addr.Region == "global" {
				region = ""
			} else {
				region = addr.Region
			}
			var usageType string
			if addr.Status == "IN_USE" {
				usageType = "STANDARD"
			} else {
				usageType = ""
			}
			addrSkus, _ := cache.GetSkusForIpAddress(db, region, usageType)
			pi, _ := cache.GetPricingInfo(db, addrSkus)
			c, err := ipAddrCostRange(db, region, usageType, pi)
			if err != nil {
				return nil, err
			}
			c = append([]string{p.Name, assets.GcloudProvider, region}, c...)
			costs = append(costs, c)
		}
		for _, snw := range nw.Subnetworks {
			ncosts := make([][]string, 2)
			region, _ := assets.SubnetworkRegion(*snw)
			if region == "" {
				fmt.Printf("Missing region in subnetwork %s:%s\n",
					nw.Name, snw.Name)
				region = "us-central1"
			}
			externalEgressSkus, _ := cache.GetSkusForExternalEgress(
				db, region, tier)
			pi, _ := cache.GetPricingInfo(db, externalEgressSkus)
			c1, err := subnetworkCostRange(db, *snw, true, pi)
			if err != nil {
				return nil, err
			}
			ncosts[0] = append([]string{p.Name, assets.GcloudProvider, snw.Name}, c1...)
			internalEgressSkus, _ := cache.GetSkusForInternalEgress(
				db, region)
			pi, _ = cache.GetPricingInfo(db, internalEgressSkus)
			c2, err := subnetworkCostRange(db, *snw, false, pi)
			if err != nil {
				return nil, err
			}
			ncosts[1] = append([]string{p.Name, assets.GcloudProvider, snw.Name}, c2...)
			costs = append(costs, ncosts...)
		}
	}
	return costs, nil
}

func getTotalsForRate(
	price cache.PricingInfo, maxUsage uint64, expectedUsage uint64) (
	maxCost float64, expectedCost float64, err error) {
	conversionRate := 1.0
	if price.CurrencyConversionRate != 1.0 {
		conversionRate = float64(price.CurrencyConversionRate)
	}
	pe := price.PricingExpression
	maxTotal := 0.0
	maxRemaining := maxUsage
	expectedTotal := 0.0
	expectedRemaining := expectedUsage
	length := len(pe.TieredRates)
	for i, r := range pe.TieredRates {
		if r.Nanos == 0 {
			continue // freebies have no nanos
		}
		unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
		if i < length-1 {
			interval := uint64(pe.TieredRates[i+1].StartUsageAmount - r.StartUsageAmount)
			if maxRemaining > 0 {
				if maxRemaining < interval {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				} else {
					maxTotal += float64(interval) * unitPrice
					maxRemaining -= interval
				}
			}
			if expectedRemaining > 0 {
				if expectedRemaining < interval {
					expectedTotal += float64(expectedRemaining) * unitPrice
					expectedRemaining = 0
				} else {
					expectedTotal += float64(interval) * unitPrice
					expectedRemaining -= interval
				}
			}
		} else {
			if maxRemaining > 0 {
				maxTotal += float64(maxRemaining) * unitPrice
				maxRemaining = 0
			}
			if expectedRemaining > 0 {
				expectedTotal += float64(expectedRemaining) * unitPrice
				expectedRemaining = 0
			}
		}
		if maxRemaining == 0 && expectedRemaining == 0 {
			break
		}
	}
	if conversionRate != 1.0 {
		expectedTotal /= conversionRate
		maxTotal /= conversionRate
	}
	return maxTotal, expectedTotal, nil
}

func ipAddrCostRange(db *sql.DB, region string, usageType string,
	pricing map[string](cache.PricingInfo)) ([]string, error) {
	maxUsage := uint64(30 * 24)
	for _, price := range pricing {
		max, exp, err := getTotalsForRate(price, maxUsage, maxUsage)
		if err != nil {
			return nil, err
		}
		var spec string
		if usageType == "RESERVED" {
			spec = fmt.Sprintf("in %s, not attached to a VM", region)
		} else {
			spec = fmt.Sprintf("attached to a %s VM in %s", usageType, region)
		}
		// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
		return []string{
			"IP Address",
			"1",
			spec,
			fmt.Sprintf("%d h", maxUsage),
			fmt.Sprintf("%.2f USD", max),
			fmt.Sprintf("%d h", maxUsage),
			fmt.Sprintf("%.2f USD", exp),
		}, nil
	}
	return nil, nil
}

func subnetworkCostRange(db *sql.DB, subnetwork common.Subnetwork, external bool,
	pricing map[string](cache.PricingInfo)) ([]string, error) {
	var usage uint64
	resourceName := ""
	region, _ := assets.SubnetworkRegion(subnetwork)
	if external {
		usage = subnetwork.ExternalEgressGbitsPerMonth
		resourceName = fmt.Sprintf("external egress traffic from %s",
			region)
	} else {
		usage = subnetwork.InternalEgressGbitsPerMonth
		resourceName = fmt.Sprintf("internal egress traffic from %s",
			region)
	}
	// There can be several different prices depending on the regions involved,
	// just use the highest.
	var highestTotal float64
	var ncost []string
	for skuId, price := range pricing {
		_ = skuId
		max, _, err := getTotalsForRate(price, usage, 0)
		if err != nil {
			return nil, err
		}
		if max > highestTotal {
			highestTotal = max
			// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
			ncost = []string{
				"Network",
				"1",
				resourceName,
				"unknown",
				"unknown",
				fmt.Sprintf("%d Gb", usage),
				fmt.Sprintf("%.2f USD", max),
			}
		}
	}
	return ncost, nil
}

func imageCost(db *sql.DB, image common.Image, pricing map[string](cache.PricingInfo)) ([]string, error) {
	if len(pricing) != 1 {
		return nil, fmt.Errorf(
			"expected exactly one price for image but got %d",
			len(pricing))
	}
	sizeGb := uint64(image.SizeGb)
	for skuId, price := range pricing {
		_ = skuId
		max, _, err := getTotalsForRate(price, sizeGb, 0)
		if err != nil {
			return nil, err
		}
		spec := fmt.Sprintf("%d GB", image.SizeGb)
		// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
		return []string{
			"Image",
			"1",
			spec,
			fmt.Sprintf("%d GB", sizeGb),
			fmt.Sprintf("%.2f USD", max),
			fmt.Sprintf("%d GB", sizeGb),
			fmt.Sprintf("%.2f USD", max),
		}, nil
	}
	return nil, fmt.Errorf("no price found for image")
}

func diskCostRange(db *sql.DB, disk common.DiskSet,
	gdsk assets.GCloudDisk, pricing map[string](cache.PricingInfo)) ([]string, error) {
	if len(pricing) != 1 {
		return nil, fmt.Errorf(
			"expected exactly one price for disk space but got %d",
			len(pricing))
	}
	diskCount := disk.Count
	sizeGb := gdsk.ActualSizeGb
	var maxUsage uint64
	maxUsage = sizeGb * uint64(diskCount)
	var projectedUsage uint64
	projectedUsage = uint64(diskCount) * maxUsage * uint64(disk.UsageHoursPerMonth) / (24 * 8)
	for skuId, price := range pricing {
		_ = skuId
		max, exp, err := getTotalsForRate(price, maxUsage, projectedUsage)
		if err != nil {
			return nil, err
		}
		spec := fmt.Sprintf("%s in %s",
			common.PrintDiskType(*disk.Template.Type),
			common.PrintLocation(*disk.Template.Location))
		// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
		// Assume there is only one price
		return []string{
			"Disk",
			fmt.Sprintf("%d", diskCount),
			spec,
			fmt.Sprintf("%d GiBy/mo", maxUsage),
			fmt.Sprintf("%.2f USD", max),
			fmt.Sprintf("%d GiBy/mo", projectedUsage),
			fmt.Sprintf("%.2f USD", exp),
		}, nil
	}
	// Should not get here.
	return nil, fmt.Errorf("no price found for disk")
}

func localDiskCost(db *sql.DB, vm common.InstanceSet, gvm assets.GCloudVM,
	pricing map[string](cache.PricingInfo)) ([]string, error) {
	if vm.Template.LocalStorage == nil {
		return nil, nil
	}
	vmCount := vm.Count
	diskCount := uint32(len(vm.Template.LocalStorage))
	if diskCount == 0 {
		return nil, nil
	}
	var totalSizeGb uint32
	for _, disk := range vm.Template.LocalStorage {
		totalSizeGb += disk.Type.SizeGb
	}
	// you only pay for local ssd when the instance is running
	usage := vm.UsageHoursPerMonth
	costs := make([]string, 0)
	var maxUsage uint64
	var projectedUsage uint64
	for skuId, price := range pricing {
		pe := price.PricingExpression
		fmt.Printf("sku %s pricing: %+v pe: %+v\n", skuId, price, pe)
		maxUsage = uint64(30 * 24 * vmCount * totalSizeGb)
		projectedUsage = uint64(usage * vmCount *
			totalSizeGb)
		max, exp, err := getTotalsForRate(price, maxUsage, projectedUsage)
		if err != nil {
			return nil, err
		}
		spec := fmt.Sprintf("Local SSD for VMset %s", vm.Name)
		// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
		costs = []string{
			fmt.Sprintf("SSD"),
			fmt.Sprintf("%d GB local storage on each of %d VMs", totalSizeGb, vmCount),
			spec,
			fmt.Sprintf("%d GB for a month", maxUsage),
			fmt.Sprintf("%.2f USD", max),
			fmt.Sprintf("%d GB for a month", projectedUsage),
			fmt.Sprintf("%.2f USD", exp),
		}
	}
	return costs, nil
}

func licenseCost(db *sql.DB, vm common.InstanceSet, gvm assets.GCloudVM,
	pricing map[string](cache.PricingInfo)) ([][]string, error) {
	vmCount := vm.Count
	usage := vm.UsageHoursPerMonth
	costs := make([][]string, 0)
	mt, err := cache.GetMachineType(db, gvm.MachineType, gvm.Region)
	if err != nil {
		return nil, err
	}
	// Some Licenses have different charges depending on number of CPUs,
	// but that information is needed when selecting the SKU, not here.
	// cpuCount := mt.CpuCount
	memoryGbFraction := float64(mt.MemoryMb) / 1024.0
	_, div := math.Modf(memoryGbFraction)
	var memoryGb uint32
	if div >= 0.5 {
		memoryGb = uint32(math.Ceil(memoryGbFraction))
	} else {
		memoryGb = uint32(math.Floor(memoryGbFraction))
	}
	var maxUsage uint64
	var projectedUsage uint64
	var resourceName string
	for skuId, price := range pricing {
		pe := price.PricingExpression
		fmt.Printf("sku %s pricing: %+v pe: %+v\n", skuId, price, pe)
		// TODO: handle licenses with a gpu price
		if pe.UsageUnit == "h" { // cpu or gpu hours
			maxUsage = uint64(30 * 24 * vmCount)
			projectedUsage = uint64(usage * vmCount)
			resourceName = "license (cpu)"
		} else if pe.UsageUnit == "GiBy.h" {
			// I think the RAM pricing is by GiBy.h, need to double check.
			maxUsage = uint64(30 * 24 * vmCount * memoryGb)
			projectedUsage = uint64(usage * vmCount * memoryGb)
			resourceName = "license (memory)"
		} else {
			return nil, fmt.Errorf("sku %s has unknown usage unit %s",
				skuId, pe.UsageUnit)
		}
		max, exp, err := getTotalsForRate(price, maxUsage, projectedUsage)
		if err != nil {
			return nil, err
		}
		spec := fmt.Sprintf("License for OS %s", gvm.OsChoice)
		// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
		costs = append(costs, []string{
			fmt.Sprintf("OS %s", resourceName),
			fmt.Sprintf("%d", vmCount),
			spec,
			fmt.Sprintf("%d %s per month", maxUsage, pe.UsageUnit),
			fmt.Sprintf("%.2f USD", max),
			fmt.Sprintf("%d %s per month", projectedUsage, pe.UsageUnit),
			fmt.Sprintf("%.2f USD", exp),
		})
	}
	return costs, nil
}

func vmCostRange(db *sql.DB, vm common.InstanceSet, gvm assets.GCloudVM,
	pricing map[string](cache.PricingInfo)) ([][]string, error) {
	mt, err := cache.GetMachineType(db, gvm.MachineType, gvm.Region)
	if err != nil {
		return nil, err
	}
	cpuCount := mt.CpuCount
	memoryGbFraction := float64(mt.MemoryMb) / 1024.0
	_, div := math.Modf(memoryGbFraction)
	var memoryGb uint32
	if div >= 0.5 {
		memoryGb = uint32(math.Ceil(memoryGbFraction))
	} else {
		memoryGb = uint32(math.Floor(memoryGbFraction))
	}
	usage := vm.UsageHoursPerMonth
	vmCount := vm.Count
	var maxUsage uint64
	var projectedUsage uint64
	costs := make([][]string, 0)
	for skuId, price := range pricing {
		pe := price.PricingExpression
		resourceName := ""
		if pe.UsageUnit == "h" { // cpu hours
			maxUsage = uint64(30 * 24 * cpuCount * vmCount)
			projectedUsage = uint64(usage * cpuCount * vmCount)
			resourceName = "cpu"
		} else if pe.UsageUnit == "GiBy.h" {
			maxUsage = uint64(30 * 24 * memoryGb * vmCount)
			projectedUsage = uint64(usage * memoryGb * vmCount)
			resourceName = "memory"
		} else {
			return nil, fmt.Errorf("sku %s has unknown usage unit %s",
				skuId, pe.UsageUnit)
		}
		max, exp, err := getTotalsForRate(price, maxUsage, projectedUsage)
		if err != nil {
			return nil, err
		}
		spec := fmt.Sprintf("%s in %s",
			common.PrintMachineType(*vm.Template.Type),
			common.PrintLocation(*vm.Template.Location))
		// resource type | count | spec | max usage | max cost | exp. usage | exp. cost
		costs = append(costs, []string{
			fmt.Sprintf("VM %s", resourceName),
			fmt.Sprintf("%d", vmCount),
			spec,
			fmt.Sprintf("%d %s per month", maxUsage, pe.UsageUnit),
			fmt.Sprintf("%.2f USD", max),
			fmt.Sprintf("%d %s per month", projectedUsage, pe.UsageUnit),
			fmt.Sprintf("%.2f USD", exp),
		})
	}
	return costs, nil
}
