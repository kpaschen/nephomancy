package cache

import (
	"fmt"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
	"time"
)

// Min to Max costs in USD per month.
func CostRange(asset *assets.SmallAsset, pricing map[string]([]cache.PricingInfo)) error {
	fmt.Printf("asset: %s\n", asset.Name)
	for skuId, prices := range pricing {
		fmt.Printf("sku %s\n", skuId)
		if len(prices) == 0 {
			fmt.Printf("No charge")
			continue
		}
		minMax, err := asset.MinMaxResourceUsage()
		if err != nil {
			fmt.Print("unable to get resource usage: %v\n", err)
			continue
		}
		for res, usage := range minMax {
			fmt.Printf("%s: between %d and %d %s\n", res, usage.MinUsage, usage.MaxUsage, usage.UsageUnit)
		}
		for _, p := range prices {
			fmt.Printf("From: %s\n", time.Unix(p.EffectiveTime, 0).Format("01-02-2006:15:04:05"))
			al, _ := p.AggregationLevel()
			if al != "PROJECT" {
				fmt.Printf("Aggregation level is not PROJECT; this means some discounts cannot be displayed correctly.\n")
			}
			if p.CurrencyConversionRate != 1.0 {
				fmt.Printf("Base price not in USD? Conversion rate is %f\n",
				p.CurrencyConversionRate)
			}
			pe := p.PricingExpression
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
				fmt.Printf("No usage for price %v\n", p)
				continue
			}

			minTotal := 0.0
			maxTotal := 0.0
			minRemaining := min
			maxRemaining := max
			length := len(pe.TieredRates)
			for i, r := range pe.TieredRates {
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
	}
	return nil
}


