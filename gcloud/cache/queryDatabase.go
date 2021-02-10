package cache

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	"strings"
	// concrete db driver even though the code only refers to interface.
	_ "github.com/mattn/go-sqlite3"
)

const ComputeService = "6F81-5844-456A"
const ContainerService = "CCD8-9BF1-090E"
const MonitoringService = "58CD-E7C3-72CA"

// Returns a map of skuid to pricing info.
func GetPricingInfo(db *sql.DB, skus []string) (map[string](PricingInfo), error) {
	var queryPricingInfo string
	ret := make(map[string](PricingInfo))
	for _, skuId := range skus {
		queryPricingInfo = fmt.Sprintf(`SELECT p.CurrencyConversionRate,
		p.AggregationInfo, p.UsageUnit, tr.CurrencyCode,
		tr.Nanos, tr.Units, tr.StartUsageAmount FROM PricingInfo p JOIN 
		TieredRates tr on p.SkuId=tr.SkuId where p.SkuId='%s';`, skuId)
		res, err := db.Query(queryPricingInfo)
		if err != nil {
			return ret, err
		}
		defer res.Close()
		var currencyConversionRate float32
		var aggregationInfo string
		var usageUnit string
		var currencyCode string
		var nanos int64
		var units int64
		var startUsageAmount int64
		var pi *PricingInfo
		pi = nil
		tieredRates := make([]Rate, 0)
		for res.Next() {
			err = res.Scan(&currencyConversionRate, &aggregationInfo,
				&usageUnit, &currencyCode, &nanos, &units, &startUsageAmount)
			if err != nil {
				log.Printf("error scanning row: %v\n", err)
				continue
			}
			if pi == nil {
				pi = &PricingInfo{
					CurrencyConversionRate: currencyConversionRate,
					AggregationInfo:        aggregationInfo,
					PricingExpression: &Pricing{
						UsageUnit: usageUnit,
					},
				}

			}
			rate := Rate{
				CurrencyCode:     currencyCode,
				Nanos:            nanos,
				StartUsageAmount: startUsageAmount,
				Units:            units,
			}
			tieredRates = append(tieredRates, rate)

		}
		if pi != nil {
			pi.PricingExpression.TieredRates = tieredRates
			ret[skuId] = *pi
		}
	}
	return ret, nil
}

func getBeginningOfSkuQuery(querySku *strings.Builder, service string, resource string, regions []string) {
	fmt.Fprintf(querySku, `SELECT Sku.SkuId
	FROM Sku JOIN ServiceRegions ON Sku.SkuId = ServiceRegions.SkuId 
	WHERE Sku.ServiceId='%s' AND Sku.ResourceFamily='%s'`, service, resource)
	if regions != nil {
		fmt.Fprintf(querySku, " AND ServiceRegions.Region IN (")
		rcount := len(regions)
		for i, r := range regions {
			if i == rcount-1 {
				fmt.Fprintf(querySku, "'%s'", r)
			} else {
				fmt.Fprintf(querySku, "'%s',", r)
			}
		}
		fmt.Fprintf(querySku, ") ")
	}
}

func GetSkusForInstance(db *sql.DB, gvm assets.GCloudVM) ([]string, error) {
	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, ComputeService, "Compute", []string{gvm.Region})
	if gvm.Scheduling != "" {
		fmt.Fprintf(&querySku, " AND Sku.UsageType='%s' ", gvm.Scheduling)
	}
	if gvm.Sharing == "SoleTenancy" {
		querySku.WriteString(" AND Sku.Description like '% Sole Tenancy %' ")
	} else {
		querySku.WriteString(" AND Sku.Description not like '% Sole Tenancy %' ")
	}
	querySku.WriteString(" AND Sku.Description not like '% Custom %' ")
	machineType := gvm.MachineType
	resourceGroups, err := assets.ResourceGroupByMachineType(machineType)
	if err != nil {
		return nil, err
	}
	var groupsClause strings.Builder
	fmt.Fprintf(&groupsClause, "(")
	for idx, gr := range resourceGroups {
		fmt.Fprintf(&groupsClause, "'%s'", gr)
		if idx < len(resourceGroups)-1 {
			fmt.Fprintf(&groupsClause, ",")
		}
	}
	fmt.Fprintf(&groupsClause, ") ")
	fmt.Fprintf(&querySku, " AND Sku.ResourceGroup IN %s ", groupsClause.String())

	parts := strings.Split(machineType, "-")
	first := strings.ToUpper(parts[0])
	fmt.Fprintf(&querySku, " AND Sku.Description like '%s %%' ", first)
	fmt.Fprintf(&querySku, ";")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForDisk(db *sql.DB, gd assets.GCloudDisk) ([]string, error) {
	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, ComputeService, "Storage", []string{gd.Region})
	diskType := gd.DiskType
	resourceGroup := ""
	switch diskType {
	case "pd-standard":
		resourceGroup = "PDStandard"
	case "pd-ssd":
		resourceGroup = "SSD"
	case "pd-balanced":
		resourceGroup = "SSD"
		querySku.WriteString(" AND Sku.Description like 'Balanced %' ")
	default:
		log.Fatalf("Unknown disk type %s in completeDiskQuery\n", diskType)
	}
	fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='%s' ", resourceGroup)
	// TODO the region query isn't quite right for all disk types.
	if gd.IsRegional {
		querySku.WriteString(" AND Sku.GeoTaxonomyType='REGIONAL' ")
	} else {
		querySku.WriteString(" AND Sku.Description not like 'Regional %' ")
	}
	// TODO: create other disks and see which SKU they end up with.

	querySku.WriteString(";")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForImage(db *sql.DB, image assets.GCloudImage) ([]string, error) {
	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, ComputeService, "Storage", []string{image.Region})
	fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='%s'; ", "StorageImage")
	return getSkusForQuery(db, querySku.String())
}

// For most network pricing, the region is not relevant, it is enough to
// look at the high-level geographic area, like EMEA.
func getGlobalRegions() []string {
	return []string{"APAC", "EMEA", "Americas"}
}

func GetSkusForIngress(db *sql.DB, region string, networkTier string) ([]string, error) {
	var querySku strings.Builder
	fmt.Fprintf(&querySku, `SELECT Sku.SkuId FROM Sku JOIN ServiceRegions ON Sku.SkuId = ServiceRegions.SkuId 
	WHERE Sku.ResourceFamily='Network'`)
	if region != "" {
		fmt.Fprintf(&querySku, " AND ServiceRegions.Region='%s'", region)
	}
	// There are other types of ingress (e.g. GoogleIngress), but this should give an ok
	// upper bound for costs.
	if networkTier == "PREMIUM" {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='PremiumInternetIngress' ")
	} else {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='StandardInternetIngress' ")
	}
	for idx, area := range getGlobalRegions() {
		if idx == 0 {
			querySku.WriteString(" AND (Sku.Description like '% from ")
			querySku.WriteString(area)
			querySku.WriteString(" to %'")
		} else {
			querySku.WriteString(" OR Sku.Description like '% from ")
			querySku.WriteString(area)
			querySku.WriteString(" to %'")
		}
	}
	fmt.Fprintf(&querySku, ");")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForExternalEgress(db *sql.DB, region string, networkTier string) ([]string, error) {
	var querySku strings.Builder
	fmt.Fprintf(&querySku, `SELECT Sku.SkuId FROM Sku JOIN ServiceRegions ON Sku.SkuId = ServiceRegions.SkuId 
	WHERE Sku.ResourceFamily='Network'`)
	if region != "" {
		fmt.Fprintf(&querySku, " AND ServiceRegions.Region='%s'", region)
	}
	// There are other types of egress ...
	if networkTier == "PREMIUM" {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='PremiumInternetEgress' ")
	} else {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='StandardInternetEgress' ")
	}
	for idx, area := range getGlobalRegions() {
		if idx == 0 {
			querySku.WriteString(" AND (Sku.Description like '% to ")
			querySku.WriteString(area)
			querySku.WriteString("'")
		} else {
			querySku.WriteString(" OR Sku.Description like '% to ")
			querySku.WriteString(area)
			querySku.WriteString("'")
		}
	}
	fmt.Fprintf(&querySku, ");")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForInternalEgress(db *sql.DB, region string) ([]string, error) {
	var querySku strings.Builder
	fmt.Fprintf(&querySku, `SELECT Sku.SkuId FROM Sku JOIN ServiceRegions ON Sku.SkuId = ServiceRegions.SkuId 
	WHERE Sku.ResourceFamily='Network'`)
	if region != "" {
		fmt.Fprintf(&querySku, " AND ServiceRegions.Region='%s'", region)
	}
	// There are other types of internal egress. VPN is usually less expensive
	// than the other types. Not handling Intrazone (it's free atm) or InterzoneEgress
	// because InterregionEgress should be an upper bound for both.
	fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='InterregionEgress' ")
	for idx, area := range getGlobalRegions() {
		if idx == 0 {
			querySku.WriteString(" AND (Sku.Description like '% to ")
			querySku.WriteString(area)
			querySku.WriteString("'")
		} else {
			querySku.WriteString(" OR Sku.Description like '% to ")
			querySku.WriteString(area)
			querySku.WriteString("'")
		}
	}
	fmt.Fprintf(&querySku, ");")
	return getSkusForQuery(db, querySku.String())
}

func getSkusForQuery(db *sql.DB, query string) ([]string, error) {
	fmt.Printf("running sku query: %s\n", query)
	res, err := db.Query(query)
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
	fmt.Printf("returning skus: %v\n", keys)
	return keys, nil
}
