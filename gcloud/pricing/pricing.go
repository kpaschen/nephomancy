package cache

import (
	"database/sql"
	"fmt"
	"strings"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
)

func GetCost(db *sql.DB, project assets.AssetStructure) error {
	// TODO: can do the Skus by instance group
	for _, instList := range project.Instances {
		for _, instance := range instList {
			// get instance price
			skus, _ := cache.GetSkusForInstance(db, *instance)
			pi, err := cache.GetPricingInfo(db, skus)
			if err != nil {
				return err
			}
			if err = costRange(db, *instance, pi); err != nil {
				return err
			}
		}
	}
	nwGroups, err := project.InstanceGroupsForNetworking()
	if err != nil {
		return err
	}
	for key, count := range nwGroups {
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			return fmt.Errorf("Failed to parse key %s\n", key)
		}
		machineTypeName := parts[0]
		region := parts[1]
		networkTier := parts[2]

		fmt.Printf("there are %d vms of type %s in region %s at network tier %s\n",
		count, machineTypeName, region, networkTier)

		ingressSkus, _ := cache.GetSkusForIngress(db, region, networkTier)
		pi, err := cache.GetPricingInfo(db, ingressSkus)
		if err != nil {
			return err
		}
		if err = networkCostRange(db, machineTypeName, true, true, pi); err != nil {
			return err
		}

		internalEgressSkus, _ := cache.GetSkusForInternalEgress(db, region)
		pi, err = cache.GetPricingInfo(db, internalEgressSkus)
		if err != nil {
			return err
		}
		if err = networkCostRange(db, machineTypeName, false, false, pi); err != nil {
			return err
		}

		externalEgressSkus, _ := cache.GetSkusForExternalEgress(db, region, networkTier)
		pi, err = cache.GetPricingInfo(db, externalEgressSkus)
		if err != nil {
			return err
		}
		if err = networkCostRange(db, machineTypeName, false, true, pi); err != nil {
			return err
		}
		// TODO: the cost has to be returned here and it needs to be multiplied
		// by the number of vms in this group
	}
        for _, disk := range project.Disks {
                // get disk prices
                skus, _ := cache.GetSkusForDisk(db, *disk)
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return err
		}
		if err = costRange(db, *disk, pi); err != nil {
			return err
		}
                if disk.SourceImage != nil {
                        imageSkus, _ := cache.GetSkusForImage(db, *disk.SourceImage)
			pi2, err := cache.GetPricingInfo(db, imageSkus)
			if err != nil {
				return err
			}
			if err = costRange(db, *disk.SourceImage, pi2); err != nil {
				return err
			}
                }
        }
        // TODO: services, licenses

	return nil
}

func networkCostRange(db *sql.DB, machineType string, ingress bool, external bool,
pricing map[string](cache.PricingInfo)) error {
	maxBandwidth := assets.MaxBandwidthGbps(machineType, ingress, external)
	maxUsage := int64(maxBandwidth * 86400 * 30)  // per month
	resourceName := ""
	if ingress {
		resourceName = "ingress traffic"
	} else if external {
		resourceName = "external egress traffic"
	} else {
		resourceName = "internal egress traffic"
	}
	// There can be multiple SKUs for different target regions, but I only
	// want to see the most expensive one.
	costString := ""
	maxCost := 0.0
	for skuId, price := range pricing {
		var summary strings.Builder
		if price.CurrencyConversionRate != 1.0 {
			fmt.Fprintf(&summary,
			"SkuId %s: Base price not in USD? Conversion rate is %f\n",
			skuId, price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		maxTotal := 0.0
		maxRemaining := maxUsage
		length := len(pe.TieredRates)
		for i, r := range pe.TieredRates {
			if r.Nanos == 0 {
				continue  // freebies have no nanos
			}
			unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
			if i < length - 1 {
				fmt.Fprintf(&summary,
				"From %d to %d %s charge %f %s per %s for %s\n",
				r.StartUsageAmount, pe.TieredRates[i+1].StartUsageAmount,
				pe.UsageUnit, unitPrice, r.CurrencyCode, pe.UsageUnit, resourceName)
				interval := pe.TieredRates[i+1].StartUsageAmount - r.StartUsageAmount
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
				fmt.Fprintf(&summary,
				"From %d %s on charge %f %s per %s for %s\n",
				r.StartUsageAmount, pe.UsageUnit, unitPrice, r.CurrencyCode, pe.UsageUnit, resourceName)
				if maxRemaining > 0 {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				}
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		if maxUsage > 0 {
			fmt.Fprintf(&summary,
			"%s: max usage of %d %s would create a charge of %f %s per month\n",
			resourceName, maxUsage, pe.UsageUnit, maxTotal, cc)
		}
		if maxTotal > maxCost {
			maxCost = maxTotal
			costString = summary.String()
		}
	}
	fmt.Println(costString)
	return nil
}

// Max costs in USD per month.
func costRange(db *sql.DB, asset assets.BaseAsset, pricing map[string](cache.PricingInfo)) error {
	for skuId, price := range pricing {
		max, err := asset.MaxResourceUsage()
		if err != nil {
			fmt.Printf("unable to get resource usage: %v\n", err)
			continue
		}
		if price.CurrencyConversionRate != 1.0 {
			fmt.Printf("SkuId %s: Base price not in USD? Conversion rate is %f\n",
			skuId, price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		var maxUsage int64
		found := false
		resourceName := "unknown resource"
		for res, u := range max {
			if u.UsageUnit == pe.UsageUnit {
				maxUsage = u.MaxUsage
				found = true
				resourceName = res
				break
			}
		}
		if !found {
			fmt.Printf("sku %s: No resource known for price %v\n", skuId, pe)
			continue
		}
		maxTotal := 0.0
		maxRemaining := maxUsage
		length := len(pe.TieredRates)
		for i, r := range pe.TieredRates {
			if r.Nanos == 0 {
				continue  // freebies have no nanos (e.g. ubuntu dev license)
			}
			unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
			if i < length - 1 {
				fmt.Printf("From %d to %d %s charge %f %s per %s for %s\n",
				r.StartUsageAmount, pe.TieredRates[i+1].StartUsageAmount,
				pe.UsageUnit, unitPrice, r.CurrencyCode, pe.UsageUnit, resourceName)
				interval := pe.TieredRates[i+1].StartUsageAmount - r.StartUsageAmount
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
				fmt.Printf("From %d %s on charge %f %s per %s for %s\n",
				r.StartUsageAmount, pe.UsageUnit, unitPrice, r.CurrencyCode, pe.UsageUnit, resourceName)
				if maxRemaining > 0 {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				}
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		if maxUsage > 0 {
			fmt.Printf("%s: max usage of %d %s would create a charge of %f %s per month\n", resourceName, maxUsage, pe.UsageUnit, maxTotal, cc)
		}
	}
	return nil
}


