package cache

import (
	"database/sql"
	"fmt"
	"strings"
	"nephomancy/gcloud/cache"
	common "nephomancy/common/resources"
)


func GetCost(db *sql.DB, p *common.Project) error {
	for _, vmset := range p.VmSets {
		skus, _ := cache.GetSkusForInstance(db, *vmset.Template)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return err
		}
		if err = vmCostRange(db, *vmset, pi); err != nil {
			return err
		}
	}
	for _, dset := range p.DiskSets {
		skus, _ := cache.GetSkusForDisk(db, *dset.Template)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return err
		}
		if err = diskCostRange(db, *dset, pi); err != nil {
			return err
		}
	}
	for _, img := range p.Images {
		skus, _ := cache.GetSkusForImage(db, *img)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return err
		}
		if err := imageCost(db, *img, pi); err != nil {
			return err
		}
	}
	for _, nw := range p.Networks {
		for _, snw := range nw.Subnetworks {
			ingressSkus, _ := cache.GetSkusForIngress(
				db, snw.Region, "PREMIUM")
			pi, _ := cache.GetPricingInfo(db, ingressSkus)
			// TODO: this needs to be done per load balancer
			if err := subnetworkCostRange(db, *snw, true, true, pi); err != nil {
				return err
			}
			externalEgressSkus, _ := cache.GetSkusForExternalEgress(
				db, snw.Region, "PREMIUM") // todo: tiers
			pi, _ = cache.GetPricingInfo(db, externalEgressSkus)
			if err := subnetworkCostRange(db, *snw, false, true, pi); err != nil {
				return err
			}
			internalEgressSkus, _ := cache.GetSkusForInternalEgress(
				db, snw.Region)
			pi, _ = cache.GetPricingInfo(db, internalEgressSkus)
			if err := subnetworkCostRange(db, *snw, false, false, pi); err != nil {
				return err
			}
		}
	}

	return nil
}

func subnetworkCostRange(db *sql.DB, subnetwork common.Subnetwork,
                         ingress bool, external bool,
			 pricing map[string](cache.PricingInfo)) error {
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
	var highestTotal float64
	costSummary := fmt.Sprintf("%s appears to cost nothing\n", resourceName)
	for skuId, price := range pricing {
		var summary strings.Builder
		if price.CurrencyConversionRate != 1.0 {
			fmt.Fprintf(&summary,
			"SkuId %s: Base price not in USD? Conversion rate is %f\n",
			skuId, price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		maxTotal := 0.0
		maxRemaining := usage
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
			} else {
				if maxRemaining > 0 {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				}
			}
			if maxRemaining == 0 {
				break
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		if usage > 0 {
			fmt.Fprintf(&summary,
			"%s: projected usage of %d %s would cost %f %s per month\n",
			resourceName, usage, pe.UsageUnit, maxTotal, cc)
		}
		if maxTotal > highestTotal {
			highestTotal = maxTotal
                        costSummary = summary.String()
		}
	}
	fmt.Printf(costSummary)
	return nil
}

func imageCost(db *sql.DB, image common.Image, pricing map[string](cache.PricingInfo)) error {
	sizeGb := image.SizeGb
	for skuId, price := range pricing {
		if price.CurrencyConversionRate != 1.0 {
			fmt.Printf("SkuId %s: Base price not in USD? Conversion rate is %f\n",
			skuId, price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		if pe.UsageUnit != "GiBy.mo" {
			return fmt.Errorf("sku %s has unknown usage unit %s",
			skuId, pe.UsageUnit)
		}
		var remaining uint32
		remaining = sizeGb
		total := 0.0
		length := len(pe.TieredRates)
		for i, r := range pe.TieredRates {
			if r.Nanos == 0 {
				continue  // freebies have no nanos (e.g. ubuntu dev license)
			}
			unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
			if i < length - 1 {
				interval := uint32(pe.TieredRates[i+1].StartUsageAmount - r.StartUsageAmount)
				if remaining > 0 {
					if remaining < interval {
						total += float64(remaining) * unitPrice
						remaining = 0
					} else {
						total += float64(interval) * unitPrice
						remaining -= interval
					}
				}
			} else {
				if remaining > 0 {
					total += float64(remaining) * unitPrice
					remaining = 0
				}
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		if total > 0 {
			fmt.Printf("image cost %f %s per month\n", total, cc)
		}
	}
	return nil
}

func diskCostRange(db *sql.DB, disk common.DiskSet, pricing map[string](cache.PricingInfo)) error {
	diskCount := disk.Count
	sizeGb := disk.Template.ActualSizeGb
	var maxUsage uint64
	maxUsage = sizeGb * 30 * 24 * uint64(diskCount)
	var projectedUsage uint64
	projectedUsage = uint64(diskCount) * maxUsage * uint64(disk.PercentUsedAvg) / 100
	for skuId, price := range pricing {
		if price.CurrencyConversionRate != 1.0 {
			fmt.Printf("SkuId %s: Base price not in USD? Conversion rate is %f\n",
			skuId, price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		if pe.UsageUnit != "GiBy.mo" {
			return fmt.Errorf("sku %s has unknown usage unit %s",
			skuId, pe.UsageUnit)
		}
		maxTotal := 0.0
		maxRemaining := maxUsage
		projectedTotal := 0.0
		projectedRemaining := projectedUsage
		length := len(pe.TieredRates)
		for i, r := range pe.TieredRates {
			if r.Nanos == 0 {
				continue  // freebies have no nanos (e.g. ubuntu dev license)
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
				if projectedRemaining > 0 {
					if projectedRemaining < interval {
						projectedTotal += float64(projectedRemaining) * unitPrice
						projectedRemaining = 0
					} else {
						projectedTotal += float64(interval) * unitPrice
						projectedRemaining -= interval
					}
				}
			} else {
				if maxRemaining > 0 {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				}
				if projectedRemaining > 0 {
					projectedTotal += float64(projectedRemaining) * unitPrice
					projectedRemaining = 0
				}
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		if projectedUsage > 0 {
			fmt.Printf("disk: projected usage of %d %s would cost %f %s per month\n", projectedUsage, pe.UsageUnit, projectedTotal, cc)
		}
		if maxUsage > 0 {
			fmt.Printf("disk: max usage of %d %s would cost %f %s per month\n", maxUsage, pe.UsageUnit, maxTotal, cc)
		}
	}
	return nil
}

func vmCostRange(db *sql.DB, vm common.VMSet, pricing map[string](cache.PricingInfo)) error {
	cpuCount := vm.Template.Type.CpuCount
	memoryGb := vm.Template.Type.MemoryGb
	usage := vm.UsageHoursPerMonth
	vmCount := vm.Count
	var maxUsage uint32
	var projectedUsage uint32
	for skuId, price := range pricing {
		if price.CurrencyConversionRate != 1.0 {
			fmt.Printf("SkuId %s: Base price not in USD? Conversion rate is %f\n",
			skuId, price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		resourceName := ""
		if pe.UsageUnit == "h" {  // cpu hours
			maxUsage = 30 * 24 * cpuCount
			projectedUsage = usage * cpuCount
			resourceName = "cpu"
		} else if pe.UsageUnit == "GiBy.h" {
			maxUsage = 30 * 24 * memoryGb
			projectedUsage = usage * memoryGb
			resourceName = "memory"
		} else {
			return fmt.Errorf("sku %s has unknown usage unit %s",
			skuId, pe.UsageUnit)
		}
		maxTotal := 0.0
		maxRemaining := maxUsage * vmCount
		projectedTotal := 0.0
		projectedRemaining := projectedUsage * vmCount
		length := len(pe.TieredRates)
		for i, r := range pe.TieredRates {
			if r.Nanos == 0 {
				continue  // freebies have no nanos (e.g. ubuntu dev license)
			}
			unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
			if i < length - 1 {
				interval := uint32(pe.TieredRates[i+1].StartUsageAmount - r.StartUsageAmount)
				if maxRemaining > 0 {
					if maxRemaining < interval {
						maxTotal += float64(maxRemaining) * unitPrice
						maxRemaining = 0
					} else {
						maxTotal += float64(interval) * unitPrice
						maxRemaining -= interval
					}
				}
				if projectedRemaining > 0 {
					if projectedRemaining < interval {
						projectedTotal += float64(projectedRemaining) * unitPrice
						projectedRemaining = 0
					} else {
						projectedTotal += float64(interval) * unitPrice
						projectedRemaining -= interval
					}
				}
			} else {
				if maxRemaining > 0 {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				}
				if projectedRemaining > 0 {
					projectedTotal += float64(projectedRemaining) * unitPrice
					projectedRemaining = 0
				}
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		if projectedUsage > 0 {
			fmt.Printf("%s: projected usage of %d %s would cost %f %s per month\n", resourceName, projectedUsage, pe.UsageUnit, projectedTotal, cc)
		}
		if maxUsage > 0 {
			fmt.Printf("%s: max usage of %d %s would cost %f %s per month\n", resourceName, maxUsage, pe.UsageUnit, maxTotal, cc)
		}
	}
	return nil
}
