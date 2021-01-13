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
	if err = createBillingTables(db); err != nil {
		return err
	}
	return createResourceMetadataTables(db)
}

func createBillingTables(db *sql.DB) error {
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

func createResourceMetadataTables(db *sql.DB) error {
	createRegionZoneTableSQL := `CREATE TABLE IF NOT EXISTS RegionZone (
		"Region" STRING NOT NULL,
		"Zone" STRING NOT NULL PRIMARY KEY
	);`
	if err := createTable(db, &createRegionZoneTableSQL); err != nil {
		return err
	}

	// Machine types actually have an id (number), I guess that's
	// per-zone. Need to check whether the same machine type name
	// resolves to different cpu numbers etc depending on zone?

	// TODO: add accelerators?
	createMachineTypeTableSQL := `CREATE TABLE IF NOT EXISTS MachineTypes (
		"MachineType" STRING NOT NULL PRIMARY KEY,
		"CpuCount" INTEGER NOT NULL,
		"MemoryMb" INTEGER NOT NULL,
		"IsSharedCpu" INTEGER  
	);`
	if err := createTable(db, &createMachineTypeTableSQL); err != nil {
		return err
	}

	// TODO: also use ScratchDisks from that table?
	createMachineTypesByZoneTableSQL := `CREATE TABLE IF NOT EXISTS MachineTypesByZone (
		"Zone" STRING NOT NULL,
		"MachineType" STRING NOT NULL,
		UNIQUE (Zone, MachineType)
		FOREIGN KEY (Zone)
		REFERENCES RegionZone (Zone)
		ON DELETE CASCADE
		ON UPDATE NO ACTION,
		FOREIGN KEY (MachineType)
		REFERENCES MachineTypes (MachineType)
		ON DELETE NO ACTION
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, &createMachineTypesByZoneTableSQL); err != nil {
		return err
	}

	// Region can be part of disk type for regional disks.
	createDiskTypesTableSQL := `CREATE TABLE IF NOT EXISTS DiskTypes (
		"DiskType" STRING NOT NULL,
		"DefaultSizeGb" INTEGER NOT NULL,
		"Region" STRING,
		UNIQUE (DiskType, DefaultSizeGb, Region)
	);`
	if err := createTable(db, &createDiskTypesTableSQL); err != nil {
		return err
	}

	createDiskTypesByZoneTableSQL := `CREATE TABLE IF NOT EXISTS DiskTypesByZone (
		"Zone" STRING NOT NULL,
		"DiskType" STRING NOT NULL,
		FOREIGN KEY (Zone)
		REFERENCES RegionZone (Zone)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, &createDiskTypesByZoneTableSQL); err != nil {
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

