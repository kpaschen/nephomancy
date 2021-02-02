package pricing

import (
	"database/sql"
	"fmt"
	"nephomancy/gcloud/cache"
	common "nephomancy/common/resources"
)


func GetCost(db *sql.DB, p *common.Project) ([][]string, error) {
	costs := make([][]string, 0)
	for _, vmset := range p.VmSets {
		skus, _ := cache.GetSkusForInstance(db, *vmset.Template)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
		}
		vmcosts, err := vmCostRange(db, *vmset, pi)
		if err != nil {
			return nil, err
		}
		costs = append(costs, vmcosts...)
	}
	for _, dset := range p.DiskSets {
		skus, _ := cache.GetSkusForDisk(db, *dset.Template)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return nil, err
		}
		dcosts, err := diskCostRange(db, *dset, pi)
		if err != nil {
			return nil, err
		}
		costs = append(costs, dcosts)
	}
	for _, img := range p.Images {
		skus, _ := cache.GetSkusForImage(db, *img)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return nil, err
		}
		icosts, err := imageCost(db, *img, pi)
		if err != nil {
			return nil, err
		}
		costs = append(costs, icosts)
	}
	for _, nw := range p.Networks {
		ncosts := make([][]string, 3)
		for _, snw := range nw.Subnetworks {
			ingressSkus, _ := cache.GetSkusForIngress(
				db, snw.Region, "PREMIUM")
			pi, _ := cache.GetPricingInfo(db, ingressSkus)
			// TODO: this needs to be done per load balancer
			c1, err := subnetworkCostRange(db, *snw, true, true, pi)
			if err != nil {
				return nil, err
			}
			ncosts[0] = c1
			externalEgressSkus, _ := cache.GetSkusForExternalEgress(
				db, snw.Region, "PREMIUM") // todo: tiers
			pi, _ = cache.GetPricingInfo(db, externalEgressSkus)
			c2, err := subnetworkCostRange(db, *snw, false, true, pi)
			if err != nil {
				return nil, err
			}
			ncosts[1] = c2
			internalEgressSkus, _ := cache.GetSkusForInternalEgress(
				db, snw.Region)
			pi, _ = cache.GetPricingInfo(db, internalEgressSkus)
			c3, err := subnetworkCostRange(db, *snw, false, false, pi)
			if err != nil {
				return nil, err
			}
			ncosts[2] = c3
		}
		costs = append(costs, ncosts...)
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
			continue  // freebies have no nanos
		}
		unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
		if i < length - 1 {
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

func subnetworkCostRange(db *sql.DB, subnetwork common.Subnetwork,
                         ingress bool, external bool,
			 pricing map[string](cache.PricingInfo)) (
				 []string, error) {
	var usage uint64
	resourceName := ""
	if ingress {
		usage = subnetwork.IngressGbitsPerMonth
		resourceName = fmt.Sprintf("ingress traffic into %s", subnetwork.Region)
	} else if external {
		usage = subnetwork.ExternalEgressGbitsPerMonth
		resourceName = fmt.Sprintf("external egress traffic from %s",
		subnetwork.Region)
	} else {
		usage = subnetwork.InternalEgressGbitsPerMonth
		resourceName = fmt.Sprintf("internal egress traffic from %s",
		subnetwork.Region)
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
		spec := fmt.Sprintf("%d GB in %s", image.SizeGb, image.Region)
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

func diskCostRange(db *sql.DB, disk common.DiskSet, pricing map[string](cache.PricingInfo)) ([]string, error) {
	if len(pricing) != 1 {
		return nil, fmt.Errorf(
			"expected exactly one price for disk space but got %d",
			len(pricing))
	}
	diskCount := disk.Count
	sizeGb := disk.Template.ActualSizeGb
	var maxUsage uint64
	maxUsage = sizeGb * 30 * 24 * uint64(diskCount)
	var projectedUsage uint64
	projectedUsage = uint64(diskCount) * maxUsage * uint64(disk.PercentUsedAvg) / 100
	for skuId, price := range pricing {
		_ = skuId
		max, exp, err := getTotalsForRate(price, maxUsage, projectedUsage)
		if err != nil {
			return nil, err
		}
		spec := fmt.Sprintf("%d GB in %s", sizeGb, disk.Template.Region)
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

func vmCostRange(db *sql.DB, vm common.VMSet, pricing map[string](cache.PricingInfo)) ([][]string, error) {
	cpuCount := vm.Template.Type.CpuCount
	memoryGb := vm.Template.Type.MemoryGb
	usage := vm.UsageHoursPerMonth
	vmCount := vm.Count
	var maxUsage uint64
	var projectedUsage uint64
	costs := make([][]string, 0)
	for skuId, price := range pricing {
		pe := price.PricingExpression
		resourceName := ""
		if pe.UsageUnit == "h" {  // cpu hours
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
		spec := fmt.Sprintf("%v in %s", vm.Template.Type, vm.Template.Region)
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
