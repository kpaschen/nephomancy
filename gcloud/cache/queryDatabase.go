package cache

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"nephomancy/gcloud/assets"
	_ "github.com/mattn/go-sqlite3"
)

// returns a map of skuid to pricing info
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
					AggregationInfo: aggregationInfo,
					PricingExpression: &Pricing{
						UsageUnit: usageUnit,
					},
				}

			}
			rate := Rate{
				CurrencyCode: currencyCode,
				Nanos: nanos,
				StartUsageAmount: startUsageAmount,
				Units: units,
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

func getBeginningOfSkuQuery(querySku *strings.Builder, asset assets.BaseAsset) {
	service := asset.BillingService()
	if service == assets.BS_TODO {
		return
	}
	resource := asset.ResourceFamily()
	regions := asset.Regions()
	fmt.Fprintf(querySku, `SELECT Sku.SkuId
	FROM Sku JOIN ServiceRegions ON Sku.SkuId = ServiceRegions.SkuId 
	WHERE Sku.ServiceId='%s' AND Sku.ResourceFamily='%s'`, service, resource)
	if regions != nil {
		fmt.Fprintf(querySku, " AND ServiceRegions.Region IN (")
		rcount := len(regions)
		for i, r := range regions {
			if i == rcount - 1 {
				fmt.Fprintf(querySku, "'%s'", r)
			} else {
				fmt.Fprintf(querySku, "'%s',", r)
			}
		}
		fmt.Fprintf(querySku, ") ")
	}
}

func GetSkusForInstance(db *sql.DB, asset assets.Instance) ([]string, error) {
	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, asset)
	if asset.Scheduling != "" {
		fmt.Fprintf(&querySku, " AND Sku.UsageType='%s' ", asset.Scheduling)
	}
	machineType := asset.MachineTypeName
	// Now the machine type is something like 'n1-standard-2' but the ResourceGroup
	// only uses part of that, so n1-standard-2 needs ResourceGroup N1Standard.
	// TODO: not sure if this is always correct, maybe need to do a lookup table.
	parts := strings.Split(machineType, "-")
	if len(parts) != 3 {
		log.Fatalf("Unexpected machine type format %s\n", machineType)
	}
	resourceGroup := fmt.Sprintf("%s%s", strings.Title(parts[0]), strings.Title(parts[1]))
	if machineType != "" {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='%s' ", resourceGroup)
	}
	fmt.Fprintf(&querySku, ";")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForDisk(db *sql.DB, asset assets.Disk) ([]string, error) {
	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, asset)

	// TODO: look for status "READY"
	diskType := asset.DiskTypeName
	resourceGroup := ""
	switch diskType {
	case "pd-standard":
		resourceGroup = "PDStandard"
	case "ssd":
		resourceGroup = "SSD"
	default:
		log.Fatalf("Unknown disk type %s in completeDiskQuery\n", diskType)
	}
	fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='%s' ", resourceGroup)
	// TODO the region query isn't quite right for all disk types.
	if asset.IsRegional {
		querySku.WriteString(" AND Sku.Description like 'Regional %' ")
	} else {
		querySku.WriteString(" AND Sku.Description like 'Storage %' AND Sku.GeoTaxonomyType='MULTI_REGIONAL'")
	}
	// TODO: create other disks and see which SKU they end up with.

	fmt.Fprintf(&querySku, ";")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForImage(db *sql.DB, asset assets.Image) ([]string, error) {
	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, asset)
	fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='%s'; ", "StorageImage")
	return getSkusForQuery(db, querySku.String())
}

func GetSkusForNetwork(db *sql.DB, asset assets.Network) ([]string, error) {
	// egress region to region service: services/compute. rf: Network. ServiceRegion: the subnetwork's region
	// rg can be VPNInternetEgress, InterregionEgress, VPNInterregionEgress, PremiumInternetEgress, InterzoneEgress
	// LoadBalancing, InterconnectPort, VPNTunnel, NAT

	// get the network tier. default is set on project ("defaultNetworkTier" can be "PREMIUM" or "STANDARD"). can also be set on vm nic.
	// probably only need to look at egress from regions where there are vms or storage.
	// actually only vms for now, since storage and spanner have their own pricing,
	// which apparently includes network.

	// generally no charge for ingress, but there is a charge for load balancers, nat,
	// protocol forwarding (which all process ingress traffic that comes from outside the
	// gcloud network)

	// generally 20c per GB is a good upper bound in premium tier. standard is less than
	// half that.

	var querySku strings.Builder
	getBeginningOfSkuQuery(&querySku, asset)
	maxTier := "STANDARD"
	regions := make([]string, 0)
	for _, sw := range asset.Subnetworks {
		if sw.MaxTier != "" {
			regions = append(regions, sw.Region)
			if sw.MaxTier != "STANDARD" {
				maxTier = sw.MaxTier
			}
		}
	}
	// There are more types of egress, e.g. for VPNs. But this should give an
	// upper bound?
	if maxTier == "PREMIUM" {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='PremiumInternetEgress' ")
	} else {
		fmt.Fprintf(&querySku, " AND Sku.ResourceGroup='InterregionEgress' ")
	}
	if len(regions) > 0 {
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
	fmt.Fprintf(&querySku, ";")

	return getSkusForQuery(db, querySku.String())
}

func getSkusForQuery(db *sql.DB, query string) ([]string, error) {
	fmt.Printf("query: %s\n", query)
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
	return keys, nil
}

