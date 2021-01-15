package cache

import (
	"encoding/json"
	"fmt"
)

type Rate struct {
	// ISO currency code. Default is USD
	CurrencyCode string
	// nanos of /Units/ bits of /CurrencyCode/.
	Nanos int64
	// previous tiered rate until this usage amount is reached.
	StartUsageAmount int64
	// This can be 0 or negative (for discounts?) or positive.
	Units int64
}

type Pricing struct {
	// This is the base unit for the Sku, usually a much smaller unit
	// than UsageUnit. E.g. Bytes per second.
	BaseUnit string
	// Conversion factor from usageUnit to baseUnit price.
	// according to documentation:
	// unitPrice / baseUnitConversionFactor = price per baseUnit.
	BaseUnitConversionFactor int64
	// This is the unit the pricing is specified in. E.g. Gibibytes/month.
	UsageUnit string
	// The tiered rates is how you compute the unit price.
	TieredRates []Rate
}

type PricingInfo struct {
	CurrencyConversionRate float32
	PricingExpression *Pricing
	AggregationInfo string
}

func (pricingInfo *PricingInfo) AggregationLevel() (string, error) {
	if pricingInfo.AggregationInfo == "" {
		return "", nil
	}
	aBytes := []byte(pricingInfo.AggregationInfo)
	var ag map[string]interface{}
	json.Unmarshal(aBytes, &ag)
	lvl, ok := ag["aggregationLevel"].(string)
	if !ok {
		return "", fmt.Errorf("aggregation level has unexpected type %T\n", ag["aggregationLevel"])
	}
	return lvl, nil
}

func FromJson(pricingExpression *string) (*Pricing, error) {
	pBytes := []byte(*pricingExpression)
	var pi map[string]interface{}
	json.Unmarshal(pBytes, &pi)
	baseUnit, ok := pi["baseUnit"].(string)
	if !ok {
		return nil, fmt.Errorf("missing baseunit on pricing expression %s\n", *pricingExpression)
	}
	cf, ok := pi["baseUnitConversionFactor"].(float64)
	if !ok {
		return nil, fmt.Errorf("conv factor is type %T\n", pi["baseUnitConversionFactor"])
	}
	conversionFactor := int64(cf)
	usageUnit, ok := pi["usageUnit"].(string)
	if !ok {
		usageUnit = ""
	}
	rates, ok := pi["tieredRates"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("tiered rates are type %T\n", pi["tieredRates"])
	}
	tieredRates := make([]Rate, 0)
	for _, r := range rates {
		// Each 'rate' is a map
		rate, ok := r.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("tiered rate is type %T\n", r)
		}

		uMap, ok := rate["unitPrice"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unit price type is %T\n", rate["unitPrice"])
		}
		currencyCode, ok := uMap["currencyCode"].(string)
		if !ok {
			currencyCode = "USD"
		}
		var nanos, unit, startUsageAmount int64
		n, ok := uMap["nanos"].(float64)
		if !ok {
			nanos = 0
		} else {
			nanos = int64(n)
		}
		u, ok := uMap["units"].(float64)
		if !ok {
			unit = 1
		} else {
			unit = int64(u)
		}
		freebie, ok := rate["startUsageAmount"].(float64)
		if !ok {
			startUsageAmount = 0
		} else {
			startUsageAmount = int64(freebie)
		}
		tieredRates = append(tieredRates, Rate{
			CurrencyCode: currencyCode,
			Nanos: nanos,
			Units: unit,
			StartUsageAmount: startUsageAmount,
		})
	}
	return &Pricing{
		BaseUnit: baseUnit,
                BaseUnitConversionFactor: conversionFactor,
		UsageUnit: usageUnit,
		TieredRates: tieredRates,
	}, nil
}
