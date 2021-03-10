package cache

import (
	"database/sql"
)

func CreateOrUpdateDatabase(db *sql.DB) error {
	return createTables(db)
}

func createTable(db *sql.DB, ct string) error {
	stmt, err := db.Prepare(ct)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func createTables(db *sql.DB) error {
	// DCS have multiple Locations. You can choose a location manually when
	// creating a DDC (dynamic data center), but since all their locations
	// are in Switzerland and all have the same prices, this is not modelled
	// in nephomancy.

	// DCS don't really have machine types. They do have SLA tiers that determine
	// pricing for everything though. VMs can be configured by number of vCPUs
	// (occasionally 'CU', where it seems like 10 CU ~ 1 vCPU) and GB of RAM,
	// and are charged by the hour. I don't think they have an option for buying
	// dedicated machines.

	createCPUCostsTableSQL := `CREATE TABLE IF NOT EXISTS CpuCosts (
		"SLA" TEXT NOT NULL PRIMARY KEY,
		"UsageUnit" TEXT,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER
	);`
	if err := createTable(db, createCPUCostsTableSQL); err != nil {
		return err
	}

	createMemoryCostsTableSQL := `CREATE TABLE IF NOT EXISTS MemoryCosts (
		"SLA" TEXT NOT NULL PRIMARY KEY,
		"UsageUnit" TEXT,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER
	);`
	if err := createTable(db, createMemoryCostsTableSQL); err != nil {
		return err
	}

	// Disks come in 'fast' or 'ultra fast' and with or without backup.
	// With DCS, Disks appear to be tied to VMs.
	createDiskCostsTableSQL := `CREATE TABLE IF NOT EXISTS DiskCosts (
		"SLA" TEXT NOT NULL,
		"DiskType" STRING NOT NULL,
		"Backup" INTEGER NOT NULL,
		"UsageUnit" STRING,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER,
		PRIMARY KEY (SLA, DiskType)
	);`
	if err := createTable(db, createDiskCostsTableSQL); err != nil {
		return err
	}

	createIPAddrCostsTableSQL := `CREATE TABLE IF NOT EXISTS IpAddrCosts (
		"SLA" TEXT NOT NULL,
		"Cidr" INTEGER NOT NULL,
		"UsageUnit" STRING,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER,
		PRIMARY KEY (SLA, Cidr)
	);`
	if err := createTable(db, createIPAddrCostsTableSQL); err != nil {
		return err
	}

	// All DCS bandwidth costs are symmetric. Some traffic to internal
	// targets is free, that is not modelled here.
	// Documentation says you can get higher bandwidths on request.
	createBandwidthCostsTableSQL := `CREATE TABLE IF NOT EXISTS BandwidthCosts (
		"SLA" TEXT NOT NULL,
		"UsageUnit" STRING,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER,
		"MaxMbits" INTEGER,
		PRIMARY KEY (SLA, MaxMbits)
	);`
	if err := createTable(db, createBandwidthCostsTableSQL); err != nil {
		return err
	}

	// This is for a NAT gateway + Firewall with loadbalancing capabilities.
	createGatewayCostsTableSQL := `CREATE TABLE IF NOT EXISTS GatewayCosts (
		"SLA" TEXT NOT NULL,
		"UsageUnit" STRING,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER,
		"Type" TEXT NOT NULL,
                PRIMARY KEY (SLA, Type)
	);`
	if err := createTable(db, createGatewayCostsTableSQL); err != nil {
		return err
	}

	// The only options at the moment are Windows and RedHat, neither
	// is free.
	createOSCostsTableSQL := `CREATE TABLE IF NOT EXISTS OSCosts (
		"SLA" TEXT NOT NULL,
		"UsageUnit" STRING,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER,
		"Vendor" TEXT NOT NULL,
                PRIMARY KEY (SLA, Vendor)
	);`
	if err := createTable(db, createOSCostsTableSQL); err != nil {
		return err
	}

	// This is only available in the 'advanced' SLA.
	// payment by GB, based on hourly peak API usage.
	createObjectStorageCostsTableSQL := `CREATE TABLE IF NOT EXISTS ObjectStorageCosts (
		"SLA" TEXT NOT NULL PRIMARY KEY,
		"UsageUnit" STRING,
		"Summary" TEXT,
		"CurrencyCode" TEXT,
		"Nanos" INTEGER
	);`
	if err := createTable(db, createObjectStorageCostsTableSQL); err != nil {
		return err
	}
	return nil
}
