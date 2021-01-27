package cache

import (
	"database/sql"
	"log"
	"strings"
	"time"
	"nephomancy/gcloud/assets"
	_ "github.com/mattn/go-sqlite3"
)

func populateBillingServices(db *sql.DB) error {
	insertBS := `INSERT INTO BillingServices(ServiceId,
	DisplayName, LastUpdatedTS) VALUES (?,?,?)
	ON CONFLICT(ServiceId) DO UPDATE SET
	DisplayName=excluded.DisplayName,
	LastUpdatedTS=excluded.LastUpdatedTS;
	`
	statement, err := db.Prepare(insertBS)
	if err != nil {
		return err
	}
	bServices, bErr := assets.ListBillingServices()
	if bErr != nil {
		return bErr
	}
	for _, b := range bServices {
		now := time.Now().Unix()
		_, err = statement.Exec(b.ServiceId, b.DisplayName, now)
		if err != nil {
			return err
		}
	}
	return nil
}

func populateSkuTable(db *sql.DB, billingServiceName *string) error {
	insertSku := `REPLACE INTO Sku(SkuId, Name, Description,
	ResourceFamily, ResourceGroup, UsageType,
	ServiceId, GeoTaxonomyType, Regions)
	VALUES (?,?,?,?,?,?,?,?,?);`
	statement, err := db.Prepare(insertSku)
	if err != nil {
		return err
	}
	insertServiceRegions := `REPLACE INTO ServiceRegions (Region, SkuId)
	VALUES (?,?);`
	srStatement, err := db.Prepare(insertServiceRegions)
	if err != nil {
		return err
	}
	insertPricingInfo := `REPLACE INTO PricingInfo (
	Summary, CurrencyConversionRate, BaseUnit, BaseUnitConversionFactor, UsageUnit,
	AggregationInfo, SkuId) VALUES (?,?,?,?,?,?,?);`
	pStatement, err := db.Prepare(insertPricingInfo)
	if err != nil {
		return err
	}

	insertTieredRate := `REPLACE INTO TieredRates (CurrencyCode, Nanos, Units,
	StartUsageAmount, SkuId, TierNumber) VALUES (?,?,?,?,?,?);`
	tStatement, err := db.Prepare(insertTieredRate)
	if err != nil {
		return err
	}
	skus, serr := assets.ListSkus(billingServiceName)
	if serr != nil {
		return serr
	}
	for _, s := range skus {
		if s.Category == nil {
			log.Printf("skipping sku %s because its category is nil\n", s.Name)
			continue
		}
		// s.Name is expected to be of the form services/<sid>/skus/<skuid>
                parts := strings.Split(s.Name, "/")
		if len(parts) != 4 {
			log.Printf("skipping sku %s because its name does not have four parts\n", s.Name)
			// could also verify that parts[3] == s.SkuId
			continue
		}
		_, err := statement.Exec(
			s.SkuId, s.Name, s.Description,
			s.Category.ResourceFamily, s.Category.ResourceGroup, s.Category.UsageType,
			parts[1], s.GeoTaxonomyType, s.GeoTaxonomyRegions)
		if err != nil {
			return err
		}
		for _, sr := range s.ServiceRegions {
			_, err := srStatement.Exec(sr, s.SkuId)
			if err != nil {
				log.Printf("not adding region %s to sku %s because %v\n",
				sr, s.SkuId, err)
				continue
			}
		}
		for _, p := range s.PricingInfo {
			pr, err := FromJson(&p.PricingExpression)
			if err != nil {
				log.Printf("Failed to parse pricing expression %s: %v\n",
				p.PricingExpression, err)
				continue
			}
			_, err = pStatement.Exec(
			p.Summary, p.CurrencyConversionRate,
			pr.BaseUnit, pr.BaseUnitConversionFactor, pr.UsageUnit,
			p.AggregationInfo, s.SkuId)
			if err != nil {
				return err
			}
			for tierNo, tier := range pr.TieredRates {
				_, err = tStatement.Exec(
					tier.CurrencyCode, tier.Nanos, tier.Units,
					tier.StartUsageAmount,
					s.SkuId, tierNo)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func populateComputeMetadata(db *sql.DB, project string) error {
	zones, err := assets.ListZones(project)
	if err != nil {
		return err
	}
	insertRegionZones := `INSERT INTO RegionZone(Region, Zone)
	VALUES (?,?)
	ON CONFLICT(Zone) DO UPDATE SET
	Region=excluded.Region;
	`
	statement, err := db.Prepare(insertRegionZones)
	if err != nil {
		return err
	}

	for _, rz := range zones {
		_, err = statement.Exec(rz.Region, rz.Zone)
		if err != nil {
			return err
		}
	}

	insertMachineType := `REPLACE INTO MachineTypes(MachineType,
	CpuCount, MemoryMb, IsSharedCpu) VALUES (?,?,?,?)`;
	insertMachineTypeByZone := `
	REPLACE INTO MachineTypesByZone(Zone, MachineType)
	VALUES (?,?);`
	insertAcceleratorTypes := `REPLACE INTO AcceleratorTypes(
		AcceleratorType, MachineType, AcceleratorCount)
		VALUES (?,?,?)`;

	statement, err = db.Prepare(insertMachineType)
	if err != nil {
		return err
	}
	statement2, err := db.Prepare(insertMachineTypeByZone)
	if err != nil {
		return err
	}
	statement3, err := db.Prepare(insertAcceleratorTypes)
	if err != nil {
		return err
	}

	for _, rz := range zones {
		zone := rz.Zone
		mtypes, err := assets.ListMachineTypes(project, zone)
		if err != nil {
			return err
		}
		for _, mt := range mtypes {
			_, err = statement.Exec(mt.Name, mt.CpuCount,
			mt.MemoryMb, mt.IsSharedCpu)
			if err != nil {
				return err
			}
			_, err = statement2.Exec(zone, mt.Name)
			if err != nil {
				return err
			}
			for _, at := range mt.Accelerators {
				_, err = statement3.Exec(at.Name, mt.Name, at.Count)
				if err != nil {
					return err
				}
			}
		}
	}

	insertDiskType := `REPLACE INTO DiskTypes(DiskType,
	DefaultSizeGb, Region) VALUES(?,?,"None")`;
	insertDiskTypeByZone := `
	REPLACE INTO DiskTypesByZone(Zone, DiskType)
	VALUES (?,?);`

	statement, err = db.Prepare(insertDiskType)
	if err != nil {
		return err
	}
	statement2, err = db.Prepare(insertDiskTypeByZone)
	if err != nil {
		return err
	}

	for _, rz := range zones {
		zone := rz.Zone
		dtypes, err := assets.ListDiskTypes(project, zone)
		if err != nil {
			return err
		}
		for _, dt := range dtypes {
			_, err = statement.Exec(dt.Name, dt.DefaultSizeGb)
			if err != nil {
				return err
			}
			_, err = statement2.Exec(zone, dt.Name)
			if err != nil {
				return err
			}
		}
	}

	insertDiskType = `REPLACE INTO DiskTypes(DiskType,
	DefaultSizeGb, Region) VALUES(?,?,?)`;

	statement, err = db.Prepare(insertDiskType)
	if err != nil {
		return err
	}

	for _, rz := range zones {
		region := rz.Region
		dtypes, err := assets.ListRegionDiskTypes(project, region)
		if err != nil {
			return err
		}
		for _, dt := range dtypes {
			_, err = statement.Exec(dt.Name, dt.DefaultSizeGb, region)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func PopulateDatabase(db *sql.DB, project string) error {
	log.Printf("Adding billing services to db\n")
	err := populateBillingServices(db)
	if err != nil {
		return err
	}
	// This is just a subset of the services that exist.
	const CE = "services/6F81-5844-456A"  // ComputeEngine
	const KE = "services/CCD8-9BF1-090E"  // KubernetesEngine
	const SM = "services/58CD-E7C3-72CA"  // Stackdriver monitoring
	const Stackdriver = "services/879F-1832-8749"
	const StackdriverLogging = "services/5490-F7B7-8DF6"
	const Spanner = "services/CC63-0873-48FD"
	const BigQuery = "services/24E6-581D-38E5"
	const Bigtable = "services/C3BE-24A5-0975"
	const CloudDNS = "services/FA26-5236-A128"
	const CloudSQL = "services/9662-B51E-5089"
	const Functions = "services/29E7-DA93-CA13"
	const AppEngine = "services/F17B-412E-CB64"
	const PubSub = "services/A1E8-BE35-7EBC"
	const SourceRepo = "services/CAE2-A537-4A95"
	const Support = "services/2062-016F-44A2"
	baseServices := [7]string{CE, KE, SM, Stackdriver, StackdriverLogging,
	Functions, AppEngine}
	for _, s := range baseServices {
		log.Printf("Adding skus for base service %s to db\n", s)
		err = populateSkuTable(db, &s)
		if err != nil {
			return err
		}
	}
	err = populateComputeMetadata(db, project)
	return err
}
