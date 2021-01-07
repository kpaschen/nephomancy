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
	return err
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
	insertPricingInfo := `REPLACE INTO PricingInfo (EffectiveFrom,
	Summary, CurrencyConversionRate, PricingExpression,
	AggregationInfo, SkuId) VALUES (?,?,?,?,?,?);`
	pStatement, err := db.Prepare(insertPricingInfo)
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
			_, err := pStatement.Exec(p.EffectiveFrom,
			p.Summary, p.CurrencyConversionRate,
			p.PricingExpression,
			p.AggregationInfo, s.SkuId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func PopulateDatabase(db *sql.DB) error {
	log.Printf("Adding billing services to db\n")
	err := populateBillingServices(db)
	if err != nil {
		return err
	}
	const CE = "services/6F81-5844-456A"
	const KE = "services/CCD8-9BF1-090E"
	const SM = "services/58CD-E7C3-72CA"
	baseServices := [3]string{CE, KE, SM}
	for _, s := range baseServices {
		log.Printf("Adding skus for base service %s to db\n", s)
		err = populateSkuTable(db, &s)
	}
	return err
}
