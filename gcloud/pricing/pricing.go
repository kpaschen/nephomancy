package cache

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
)

type ResourceUsage struct {
        UsageUnit string  // same kind of string as what is used in PricingExpression, e.g. By.s for bytes-seconds (how many bytes for how many seconds)
        MinUsage int64
        MaxUsage int64
}

func MinMaxResourceUsage(asset *assets.SmallAsset, rmtd *assets.ResourceMetadata) (map[string]ResourceUsage, error) {
	rf, _ := asset.ResourceFamily()
	switch rf {
		case "Network":
			return map[string]ResourceUsage{
				"egress": ResourceUsage{
					UsageUnit: "gibibyte", // TODO: is this right? Not gibibyte/hour?
					MinUsage: 0,
					MaxUsage: -1, // unlimited?
				}, }, nil
		case "Compute":
			if rmtd.Mt.CpuCount == 0 {
				return nil, fmt.Errorf("Missing compute metadata for asset %+v\n", asset)
			}
			cpuCount := rmtd.Mt.CpuCount
			memoryGb := rmtd.Mt.MemoryMb / 1024.0
			// TODO: use information on shared cpu

			return map[string]ResourceUsage{
                        "cpu": ResourceUsage{
                                UsageUnit: "h",
                                MinUsage: 0,
                                MaxUsage: 30 * 24 * cpuCount,
                        },
                        "memory": ResourceUsage{
                                UsageUnit: "GiBy.h",
                                MinUsage: 0,
                                MaxUsage: 30 * 24 * memoryGb,
                        },
                }, nil
	case "Storage":
		diskSize, err := asset.StorageSize()
		if err != nil {
			return nil, err
		}
                return map[string]ResourceUsage{
                        "diskspace": ResourceUsage{
                                UsageUnit: "GiBy.mo",
                                MinUsage: 0,
                                MaxUsage: diskSize,
                        }, }, nil
        default:
                log.Printf("No known unit for resource family %s\n", rf)
                return nil, nil
        }
        return nil, nil
}

// Min to Max costs in USD per month.
func CostRange(db *sql.DB, asset *assets.SmallAsset, pricing map[string](cache.PricingInfo)) error {
	fmt.Printf("asset: %+v\n", asset)
	md, err := cache.GetResourceMetadata(db, asset)
	if err != nil {
		return err
	}
	fmt.Printf("md: %+v\n", md)
	for skuId, price := range pricing {
		fmt.Printf("sku %s\n", skuId)
		minMax, err := MinMaxResourceUsage(asset, md)
		if err != nil {
			fmt.Printf("unable to get resource usage: %v\n", err)
			continue
		}
		for res, usage := range minMax {
			fmt.Printf("%s: between %d and %d %s\n", res, usage.MinUsage, usage.MaxUsage, usage.UsageUnit)
		}
		al, _ := price.AggregationLevel()
		if al != "PROJECT" {
			fmt.Printf("Aggregation level is not PROJECT; this means some discounts cannot be displayed correctly.\n")
		}
		if price.CurrencyConversionRate != 1.0 {
			fmt.Printf("Base price not in USD? Conversion rate is %f\n",
			price.CurrencyConversionRate)
		}
		pe := price.PricingExpression
		var min, max int64
		found := false
		for _, u := range minMax {
			if u.UsageUnit == pe.UsageUnit {
				min = u.MinUsage
				max = u.MaxUsage
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("No usage for price %v\n", price)
		}
		minTotal := 0.0
		maxTotal := 0.0
		minRemaining := min
		maxRemaining := max
		length := len(pe.TieredRates)
		for i, r := range pe.TieredRates {
			if r.Nanos == 0 {
				continue  // freebies have no nanos (e.g. ubuntu dev license)
			}
			unitPrice := (float64(r.Units) * float64(r.Nanos)) / 1000000000.0
			if i < length - 1 {
				fmt.Printf("From %d to %d %s charge %f %s per %s\n",
				r.StartUsageAmount, pe.TieredRates[i+1].StartUsageAmount,
				pe.UsageUnit, unitPrice, r.CurrencyCode, pe.UsageUnit)
				interval := pe.TieredRates[i+1].StartUsageAmount - r.StartUsageAmount
				if minRemaining > 0 {
					if minRemaining < interval {
						minTotal += float64(minRemaining) * unitPrice
						minRemaining = 0
					} else {
						minTotal += float64(interval) * unitPrice
						minRemaining -= interval
					}
				}
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
				fmt.Printf("From %d %s charge %f %s per %s\n",
				r.StartUsageAmount, pe.UsageUnit, unitPrice, r.CurrencyCode, pe.UsageUnit)
				if minRemaining > 0 {
					minTotal += float64(minRemaining) * unitPrice
					minRemaining = 0
				}
				if maxRemaining > 0 {
					maxTotal += float64(maxRemaining) * unitPrice
					maxRemaining = 0
				}
			}
		}
		cc := pe.TieredRates[0].CurrencyCode
		fmt.Printf("Min usage %d: charge %f %s\nMax usage %d charge %f %s\n",
		min, minTotal, cc, max, maxTotal, cc)
	}
	return nil
}


