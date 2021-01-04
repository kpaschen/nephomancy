package cache

import(
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func CreateNewDatabase(filename *string) error {
	os.Remove(*filename)
	file, err := os.Create(*filename)
	if err != nil {
		return err
	}
	file.Close()
	db, _ := sql.Open("sqlite3", *filename)
	defer db.Close()
	return createTables(db)
}

func createTables(db *sql.DB) error {
	createBillingServicesTableSQL := `CREATE TABLE BillingServices (
		"ServiceId" TEXT NOT NULL PRIMARY KEY,
		"DisplayName" TEXT NOT NULL,
		"LastUpdatedTS" INTEGER
	);`
	if err := createTable(db, &createBillingServicesTableSQL); err != nil {
		return err
	}

	// For regions, the ServiceRegions information looks to be more complete.
	createSkuTableSQL := `CREATE TABLE Sku (
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
	createServiceRegionsTableSQL := `CREATE TABLE ServiceRegions (
		"Region" TEXT NOT NULL,
		"SkuId" TEXT NOT NULL,
		FOREIGN KEY (SkuId)
		REFERENCES Sku (SkuId)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, &createServiceRegionsTableSQL); err != nil {
		return err
	}
	createPricingInfoTableSQL := `CREATE TABLE PricingInfo (
		"EffectiveFrom" INTEGER,
		"Summary" TEXT,
		"CurrencyConversionRate" REAL NOT NULL,
		"PricingExpression" TEXT NOT NULL,
		"AggregationInfo" TEXT,
		"SkuId" TEXT NOT NULL,
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

func populateTables(db *sql.DB) {
	insertBillingServiceSQL := `INSERT INTO BillingServices(ServiceId,
	DisplayName) VALUES (?, ?);`
	ServiceId := "123"
	DisplayName := "display name 123"
	statement, err := db.Prepare(insertBillingServiceSQL)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(ServiceId, DisplayName)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func showTables(db *sql.DB) {
	rows, err := db.Query("Select * from BillingServices;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var serviceId string
		var displayName string
		rows.Scan(&serviceId, &displayName)
		log.Printf("serviceId: %s displayName: %s\n", serviceId, displayName)
	}
}
