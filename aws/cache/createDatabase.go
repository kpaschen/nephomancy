package cache

import (
	"database/sql"
)

func CreateOrUpdateDatabase(db *sql.DB) error {
	if err := createRegionsTable(db); err != nil {
		return err
	}
	if err := createInstanceTypesTable(db); err != nil {
		return err
	}
	if err := createVolumeTypesTable(db); err != nil {
		return err
	}
	if err := createTypeByRegionTables(db); err != nil {
		return err
	}
	if err := createSkuTable(db); err != nil {
		return err
	}
	return nil
}

func createTable(db *sql.DB, ct string) error {
	stmt, err := db.Prepare(ct)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func createRegionsTable(db *sql.DB) error {
	createRegionsTableSQL := `CREATE TABLE IF NOT EXISTS Regions (
		"ID" TEXT NOT NULL PRIMARY KEY,
		"DisplayName" TEXT NOT NULL,
		"Country" TEXT,
		"Continent" TEXT,
		"Special" INTEGER
	);`
	if err := createTable(db, createRegionsTableSQL); err != nil {
		return err
	}
	return nil
}

// Table for instance types. There are more fields available,
// but I haven't yet included things like network performance,
// or the more specialised items like elastic graphcis support.
// The Storage Type can be "EBS", "SSD", "NvMe SSD", or HDD.
// When the storage type is not "EBS", you can have internal storage
// of the given type, and StorageAmount tells you how much.
// Of course you can always use an EBS volume, the storage type "EBS"
// only tells you EBS is the only block storage you can use with
// this instance type.
func createInstanceTypesTable(db *sql.DB) error {
	createInstanceTypesTableSQL := `CREATE TABLE IF NOT EXISTS InstanceTypes (
		"InstanceType" TEXT NOT NULL PRIMARY KEY,
		"InstanceFamily" TEXT,
		"CPU" INTEGER NOT NULL,
		"Memory" INTEGER NOT NULL,
		"GPU" INTEGER,
		"StorageType" TEXT,
		"StorageAmount" INTEGER
	);`
	if err := createTable(db, createInstanceTypesTableSQL); err != nil {
		return err
	}
	return nil
}

// This is for block storage devices on EBS. Internal Storage is part of
// the instance type. Might want to include the EC2-supported parts
// of Amazon S3 storage here as well? But they don't have sizes or
// storage media, so no need really.
func createVolumeTypesTable(db *sql.DB) error {
	createVolumeTypesTableSQL := `CREATE TABLE IF NOT EXISTS VolumeTypes (
		"VolumeType" TEXT NOT NULL PRIMARY KEY,
		"StorageMedia" TEXT,
		"MaxVolumeSize" INTEGER,
		"MaxIOPS" INTEGER,
		"MaxThroughput" INTEGER
	);`
	if err := createTable(db, createVolumeTypesTableSQL); err != nil {
		return err
	}
	return nil
}

func createTypeByRegionTables(db *sql.DB) error {
	createInstanceTypeRegionTableSQL := `CREATE TABLE IF NOT EXISTS InstanceTypeByRegion (
		"InstanceType" NOT NULL,
		"Region" TEXT NOT NULL,
		UNIQUE (InstanceType, Region)
		FOREIGN KEY (InstanceType)
		REFERENCES InstanceTypes (InstanceType)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
		FOREIGN KEY (Region)
		REFERENCES Regions (ID)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, createInstanceTypeRegionTableSQL); err != nil {
		return err
	}
	createVolumeTypeRegionTableSQL := `CREATE TABLE IF NOT EXISTS VolumeTypeByRegion (
		"VolumeType" NOT NULL,
		"Region" TEXT NOT NULL,
		UNIQUE (VolumeType, Region)
		FOREIGN KEY (VolumeType)
		REFERENCES VolumeTypes (VolumeType)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
		FOREIGN KEY (Region)
		REFERENCES Regions (ID)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, createVolumeTypeRegionTableSQL); err != nil {
		return err
	}
	return nil
}

// The SKU depends on the product code, usage type and operation.
// The product code is instance type + location (for instances).
// The usage type is a string of the form "[ShortLocation-]?Usage:<instance type>"
// where ShortLocation is an abbreviation of the region (e.g. EUC1 for eu-central-1),
// and Usage is one of: DedicatedUsage, BoxUsage, UnusedDed, Reservation, DedicatedRes. There is also EBS-Optimized, but it only happens when operation is Hourly,
// so that's for bandwidth.
// Dedicated vs. Box is also in the "Tenancy" field. Shared Tenancy is Box Usage,
// Dedicated Tenancy is Dedicated Usage. With Dedicated machines, you pay for used
// and unused capacity.
// The operation for instances is of the form "RunInstances[:[0-9]4]?" or "Hourly"
// The code after RunInstances means:
// empty: just Linux
// 0002: bare metal
// 0004: Linux with SQL Server
// 0006: Windows with SQL Server
// 0010: Red Hat Enterprise Linux | RHEL
// 0100: Linux with SQL Server Enterprise
// 0102: Windows with SQL Sever Enterprise
// 0200: Linux with SQL Server Web
// 0202: Windows with SQL Server Web
// 0800: Windows BYOL
// 000g: SUSE Linux
//
// There are also codes of the form RunInstances:FFP-<code> but they are all for
// things running in GovCloud.
// Hourly looks to be for something measured in Mbps, probably network bandwidth.
func createSkuTable(db *sql.DB) error {
	createSkuTableSQL := `CREATE TABLE IF NOT EXISTS Sku (
		"Sku" TEXT NOT NULL PRIMARY KEY,
		"ProductType" TEXT NOT NULL,
		"Region" TEXT NOT NULL,
		"Usage" TEXT NOT NULL,
		"Operation" TEXT NOT NULL,
		FOREIGN KEY (ProductType)
		REFERENCES InstanceTypes (InstanceType)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
		FOREIGN KEY (Region)
		REFERENCES Regions (ID)
		ON DELETE CASCADE
		ON UPDATE NO ACTION
	);`
	if err := createTable(db, createSkuTableSQL); err != nil {
		return err
	}
	return nil
}
