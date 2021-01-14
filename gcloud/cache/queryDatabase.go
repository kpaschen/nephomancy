package cache

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"nephomancy/gcloud/assets"
	_ "github.com/mattn/go-sqlite3"
)

// Returns a map of sku id to pricing information.
// Only the most recent pricing info will be used.
func GetPricingInfo(db *sql.DB, asset *assets.SmallAsset) (map[string](PricingInfo), error) {
	skus, err := getRelevantSkus(db, asset)
	if err != nil {
		return nil, err
	}
	if len(skus) == 0 {
		// log.Printf("No skus found for asset %s\n", asset.Name)
		return nil, nil
	}
	return getPricing(db, skus)
}

func getPricing(db *sql.DB, skus []string) (map[string](PricingInfo), error) {
	var queryPricingInfo string
	ret := make(map[string](PricingInfo))
	for _, skuId := range skus {
		queryPricingInfo = fmt.Sprintf(`SELECT CurrencyConversionRate, PricingExpression,
		AggregationInfo FROM PricingInfo where SkuId='%s';`, skuId)
		res, err := db.Query(queryPricingInfo)
		if err != nil {
			return ret, err
		}
		defer res.Close()
		var currencyConversionRate float32
		var pricingExpression string
		var aggregationInfo string
		for res.Next() {
			err = res.Scan(&currencyConversionRate, &pricingExpression, &aggregationInfo)
			if err != nil {
				log.Printf("error scanning row: %v\n", err)
				continue
			}

			pricing, err := FromJson(&pricingExpression)
			if err != nil {
				return ret, err
			}
			ret[skuId] = PricingInfo{
				CurrencyConversionRate: currencyConversionRate,
				PricingExpression: pricing,
				AggregationInfo: aggregationInfo,
			}
			break  // Expect only one result
		}
	}
	return ret, nil
}

// This gets potentially relevant SKUs for an asset that has a resource.
// Missing: Network resources
// Also missing: looking at status of resources (currently treat all resources as active/ready)
// Make sure this returns only SKUs that are relevant.
func getRelevantSkus(db *sql.DB, asset *assets.SmallAsset) ([]string, error) {
	service, err := asset.BillingService()
	if err != nil {
		return nil, err
	}
	if service == "" {
		return nil, nil
	}
	resource, err := asset.ResourceFamily()
	if err != nil {
		return nil, err
	}
	// TODO: network-related skus
	if resource == "Network" {
		return nil, nil
	}
	regions, err := asset.Regions()
	if err != nil {
		log.Printf("failed to get region for resource\n")
		return nil, err
	}

	var querySku strings.Builder
	fmt.Fprintf(&querySku, `SELECT Sku.SkuId
	FROM Sku JOIN ServiceRegions ON Sku.SkuId = ServiceRegions.SkuId 
	WHERE Sku.ServiceId='%s' AND Sku.ResourceFamily='%s'`, service, resource)
	if regions != nil {
		fmt.Fprintf(&querySku, " AND ServiceRegions.Region IN (")
		rcount := len(regions)
		for i, r := range regions {
			if i == rcount - 1 {
				fmt.Fprintf(&querySku, "'%s'", r)
			} else {
				fmt.Fprintf(&querySku, "'%s',", r)
			}
		}
		fmt.Fprintf(&querySku, ") ")
	}
	if err = completeQuery(&querySku, asset); err != nil {
		return nil, err
	}
	fmt.Fprintf(&querySku, ";")

	if err != nil {
		return nil, err
	}
	res, err := db.Query(querySku.String())
	if err != nil {
		return nil, err
	}
	var skuId string
	defer res.Close()
	results := make(map[string]bool)
	for res.Next() {
		err = res.Scan(&skuId)
		if err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}
		results[skuId] = true
	}
	keys := make([]string, len(results))
	i := 0
	for k := range results {
		keys[i] = k
		i++
	}
	return keys, nil
}

func completeInstanceQuery(queryBuilder *strings.Builder, asset *assets.SmallAsset) error {
	// TODO: look for status. TERMINATED does not get charged while SUSPENDED gets charged
	// under D50B-484D-C2A2 which is regional to emea.
	scheduling, err := asset.Scheduling()
	if err != nil {
		return err
	}
	if scheduling != "" {
		fmt.Fprintf(queryBuilder, " AND Sku.UsageType='%s' ", scheduling)
	}
	machineType, err := asset.MachineType()
	if err != nil {
		return err
	}
	// Now the machine type is something like 'n1-standard-2' but the ResourceGroup
	// only uses part of that, so n1-standard-2 needs ResourceGroup N1Standard.
	// TODO: not sure if this is always correct, maybe need to do a lookup table.
	parts := strings.Split(machineType, "-")
	if len(parts) != 3 {
		log.Fatalf("Unexpected machine type format %s\n", machineType)
	}
	resourceGroup := fmt.Sprintf("%s%s", strings.Title(parts[0]), strings.Title(parts[1]))
	if machineType != "" {
		fmt.Fprintf(queryBuilder, " AND Sku.ResourceGroup='%s' ", resourceGroup)
	}
	return nil
}

func completeImageQuery(queryBuilder *strings.Builder, asset *assets.SmallAsset) error {
	fmt.Fprintf(queryBuilder, " AND Sku.ResourceGroup='%s' ", "StorageImage")
	return nil
}

func completeDiskQuery(queryBuilder *strings.Builder, asset *assets.SmallAsset) error {
	// TODO: look for status "READY"
	diskType, err := asset.DiskType()
	if err != nil {
		return err
	}
	if diskType != "" {

		resourceGroup := ""
		switch diskType {
		case "pd-standard":
			resourceGroup = "PDStandard"
		case "ssd":
			resourceGroup = "SSD"
		default:
			return fmt.Errorf(
				"Unknown disk type %s in completeDiskQuery\n",
				diskType)
		}
		fmt.Fprintf(queryBuilder,
		" AND Sku.ResourceGroup='%s' ", resourceGroup)
		bt, err := asset.BaseType()
		if err != nil {
			return err
		}
		if bt == "RegionalDisk" {
			queryBuilder.WriteString(
				" AND Sku.Description like 'Regional %' ")
		} else {
			queryBuilder.WriteString(
				" AND Sku.Description like 'Storage %' ")
		}
	}
	// TODO: create other disks and see which SKU they end up with.
	return nil
}

func completeQuery(queryBuilder *strings.Builder, asset *assets.SmallAsset) error {
	at, err := asset.BaseType()
	if err != nil {
		return err
	}
	switch at {
	case "Instance":
		return completeInstanceQuery(queryBuilder, asset)
	case "Image":
		return completeImageQuery(queryBuilder, asset)
	case "Disk":
		return completeDiskQuery(queryBuilder, asset)
	case "RegionDisk":
		return completeDiskQuery(queryBuilder, asset)
	default:
		log.Printf("Not supporting asset types with resource family %s yet\n", at)
		return nil
	}
}

