package cache

import(
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func CreateOrUpdateDatabase(filename *string) error {
	handle, err := os.OpenFile(*filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	handle.Close()
	db, _ := sql.Open("sqlite3", *filename)
	defer db.Close()
	return createTables(db)
}

func createTables(db *sql.DB) error {
	createBillingServicesTableSQL := `CREATE TABLE IF NOT EXISTS BillingServices (
		"ServiceId" TEXT NOT NULL PRIMARY KEY,
		"DisplayName" TEXT NOT NULL,
		"LastUpdatedTS" INTEGER
	);`
	if err := createTable(db, &createBillingServicesTableSQL); err != nil {
		return err
	}

	// For regions, the ServiceRegions information looks to be more complete.
	createSkuTableSQL := `CREATE TABLE IF NOT EXISTS Sku (
		"SkuId" TEXT NOT NULL PRIMARY KEY,
		"Name" TEXT NOT NULL,
		"Description" TEXT,
		"ResourceFamily" TEXT NOT NULL,
		"ResourceGroup" TEXT NOT NULL,
		"UsageType" TEXT,
		"ServiceId" TEXT NOT NULL,
		"GeoTaxonomyType" TEXT NOT NULL,
		"Regions" TEXT NOT NULL,
		FOREIGN KEY (ServiceId)
		REFERENCES BillingServices (ServiceId)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, &createSkuTableSQL); err != nil {
		return err
	}
	createServiceRegionsTableSQL := `CREATE TABLE IF NOT EXISTS ServiceRegions (
		"Region" TEXT NOT NULL,
		"SkuId" TEXT NOT NULL,
		UNIQUE(Region, SkuId)
		FOREIGN KEY (SkuId)
		REFERENCES Sku (SkuId)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, &createServiceRegionsTableSQL); err != nil {
		return err
	}
	createPricingInfoTableSQL := `CREATE TABLE IF NOT EXISTS PricingInfo (
		"EffectiveFrom" INTEGER,
		"Summary" TEXT,
		"CurrencyConversionRate" REAL NOT NULL,
		"PricingExpression" TEXT NOT NULL,
		"AggregationInfo" TEXT,
		"SkuId" TEXT NOT NULL,
		UNIQUE(SkuId, EffectiveFrom)
		FOREIGN KEY (SkuId)
		REFERENCES Sku (SkuId)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, &createPricingInfoTableSQL); err != nil {
		return err
	}
	return nil
}

func createTable(db *sql.DB, ct *string) error {
	stmt, err := db.Prepare(*ct)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

