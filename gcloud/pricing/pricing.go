package cache

import (
	"database/sql"
	"fmt"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
)

func GetCost(db *sql.DB, project assets.AssetStructure) error {
	for _, instance := range project.Instances {
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
	for _, nw := range project.Networks {
		skus, err := cache.GetSkusForNetwork(db, *nw)
		if err != nil {
			return err
		}
		pi, err := cache.GetPricingInfo(db, skus)
		if err != nil {
			return err
		}
		if err = costRange(db, *nw, pi); err != nil {
			return err
		}
	}
        // TODO: services, licenses

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
		al, _ := price.AggregationLevel()
		if al != "PROJECT" {
			fmt.Printf("%s: Aggregation level is not PROJECT; this means some discounts cannot be displayed correctly.\n", skuId)
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


