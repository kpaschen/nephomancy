// Prices are based on a price list published by Swisscom on Nov 27th 2020.
// Please note that the Swisscom can change prices at any time without
// notice.
// link: https://documents.swisscom.com/product/filestore/lib/a9e6ce54-97c6-4e5c-a670-9d751bc1fe44/dcsplus%20leistungen%20und%20preise-en.pdf
package cache

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"math"
)

var nanoFactor = math.Pow(10, 9)

func populateCPUCosts(db *sql.DB) error {
	insert := `REPLACE INTO CpuCosts(SLA, UsageUnit, Summary, CurrencyCode, Nanos)
	VALUES(?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Basic", "h", "One vCPU for one hour at Basic SLA",
		"CHF", 0.02777*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Standard", "h", "One vCPU for one hour at Standard SLA",
		"CHF", 0.0361*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Advanced", "h", "One vCPU for one hour at Advanced SLA",
		"CHF", 0.06137*nanoFactor)

	return err
}

func populateMemoryCosts(db *sql.DB) error {
	insert := `REPLACE INTO MemoryCosts(SLA, UsageUnit, Summary, CurrencyCode, Nanos)
	VALUES(?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Basic", "GB", "One GB of RAM for one hour at Basic SLA",
		"CHF", 0.01458*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Standard", "GB", "One GB of RAM for one hour at Standard SLA",
		"CHF", 0.01895*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Advanced", "GB", "One GB of RAM for one hour at Advanced SLA",
		"CHF", 0.03222*nanoFactor)

	return err
}

func populateDiskCosts(db *sql.DB) error {
	insert := `REPLACE INTO DiskCosts(SLA, DiskType, Backup, UsageUnit, 
	Summary, CurrencyCode, Nanos)
	VALUES(?, ?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Basic", "Fast", 0, "GB",
		"One GB of Fast Storage for one hour at Basic SLA",
		"CHF", 0.00017*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Basic", "Fast with Backup", 1, "GB",
		"One GB of Fast Storage with backup for one hour at Basic SLA",
		"CHF", 0.00046*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Basic", "UltraFast", 0, "GB",
		"One GB of Ultra Storage for one hour at Basic SLA",
		"CHF", 0.00027*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Basic", "UltraFast with Backup", 1, "GB",
		"One GB of Ultra Storage with backup for one hour at Basic SLA",
		"CHF", 0.00056*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Standard", "Fast", 0, "GB",
		"One GB of Fast Storage for one hour at Standard SLA",
		"CHF", 0.00017*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Standard", "Fast with Backup", 1, "GB",
		"One GB of Fast Storage with backup for one hour at Standard SLA",
		"CHF", 0.00046*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Standard", "UltraFast", 0, "GB",
		"One GB of Ultra Storage for one hour at Standard SLA",
		"CHF", 0.00027*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Standard", "UltraFast with Backup", 1, "GB",
		"One GB of Ultra Storage with backup for one hour at Standard SLA",
		"CHF", 0.00056*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Advanced", "Fast", 0, "GB",
		"One GB of Fast Storage for one hour at Advanced SLA",
		"CHF", 0.00034*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Advanced", "Fast with Backup", 1, "GB",
		"One GB of Fast Storage with backup for one hour at Advanced SLA",
		"CHF", 0.00063*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Advanced", "UltraFast", 0, "GB",
		"One GB of Ultra Storage for one hour at Advanced SLA",
		"CHF", 0.00054*nanoFactor)
	if err != nil {
		return err
	}
	_, err = stmt.Exec("Advanced", "UltraFast with Backup", 1, "GB",
		"One GB of Ultra Storage with backup for one hour at Advanced SLA",
		"CHF", 0.00083*nanoFactor)
	if err != nil {
		return err
	}

	return err
}

func populateIpAddrCosts(db *sql.DB) error {
	insert := `REPLACE INTO IpAddrCosts(SLA, Cidr, UsageUnit, Summary,
	CurrencyCode, Nanos)
	VALUES(?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	// The prices are currently the same at all SLAs
	stmt.Exec("Basic", 29, "Range", "One hour of 3 public IP Addresses",
		"CHF", 0.02499*nanoFactor)
	stmt.Exec("Basic", 28, "Range", "One hour of 11 public IP Addresses",
		"CHF", 0.06250*nanoFactor)
	stmt.Exec("Basic", 27, "Range", "One hour of 27 public IP Addresses",
		"CHF", 0.09028*nanoFactor)
	stmt.Exec("Basic", 26, "Range", "One hour of 59 public IP Addresses",
		"CHF", 0.11806*nanoFactor)
	stmt.Exec("Basic", 25, "Range", "One hour of 123 public IP Addresses",
		"CHF", 0.14583*nanoFactor)
	stmt.Exec("Basic", 24, "Range", "One hour of 251 public IP Addresses",
		"CHF", 0.17361*nanoFactor)
	stmt.Exec("Standard", 29, "Range", "One hour of 3 public IP Addresses",
		"CHF", 0.02499*nanoFactor)
	stmt.Exec("Standard", 28, "Range", "One hour of 11 public IP Addresses",
		"CHF", 0.06250*nanoFactor)
	stmt.Exec("Standard", 27, "Range", "One hour of 27 public IP Addresses",
		"CHF", 0.09028*nanoFactor)
	stmt.Exec("Standard", 26, "Range", "One hour of 59 public IP Addresses",
		"CHF", 0.11806*nanoFactor)
	stmt.Exec("Standard", 25, "Range", "One hour of 123 public IP Addresses",
		"CHF", 0.14583*nanoFactor)
	stmt.Exec("Standard", 24, "Range", "One hour of 251 public IP Addresses",
		"CHF", 0.17361*nanoFactor)
	stmt.Exec("Advanced", 29, "Range", "One hour of 3 public IP Addresses",
		"CHF", 0.02499*nanoFactor)
	stmt.Exec("Advanced", 28, "Range", "One hour of 11 public IP Addresses",
		"CHF", 0.06250*nanoFactor)
	stmt.Exec("Advanced", 27, "Range", "One hour of 27 public IP Addresses",
		"CHF", 0.09028*nanoFactor)
	stmt.Exec("Advanced", 26, "Range", "One hour of 59 public IP Addresses",
		"CHF", 0.11806*nanoFactor)
	stmt.Exec("Advanced", 25, "Range", "One hour of 123 public IP Addresses",
		"CHF", 0.14583*nanoFactor)
	stmt.Exec("Advanced", 24, "Range", "One hour of 251 public IP Addresses",
		"CHF", 0.17361*nanoFactor)
	return nil
}

func populateBandwidthCosts(db *sql.DB) error {
	insert := `REPLACE INTO BandwidthCosts(SLA, UsageUnit, Summary,
	CurrencyCode, Nanos, MaxMbits) VALUES(?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	// The prices are currently the same at all SLAs
	// You can get higher bandwidths but the prices are just multiples of
	// this one, so 100MBit/s is 0.20800 etc.
	stmt.Exec("Basic", "Mbit/s",
		"One hour of symmetric bandwidth 10Mbit/s", "CHF", 0.02080*nanoFactor, 10)
	stmt.Exec("Standard", "Mbit/s",
		"One hour of symmetric bandwidth 10Mbit/s", "CHF", 0.02080*nanoFactor, 10)
	stmt.Exec("Advanced", "Mbit/s",
		"One hour of symmetric bandwidth 10Mbit/s", "CHF", 0.02080*nanoFactor, 10)
	return nil
}

func populateGatewayCosts(db *sql.DB) error {
	insert := `REPLACE INTO GatewayCosts(SLA, UsageUnit, Summary,
	CurrencyCode, Nanos, Type) VALUES(?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	stmt.Exec("Basic", "Piece", "Eco Edge Gateway per hour", "CHF", 0, "Eco")
	stmt.Exec("Basic", "Piece",
		"Fast Edge Gateway per hour", "CHF", 0.08333*nanoFactor, "Fast")
	stmt.Exec("Basic", "Piece",
		"Ultra Fast Edge Gateway per hour", "CHF", 0.12500*nanoFactor, "Ultra Fast")
	stmt.Exec("Basic", "Piece",
		"Backup Edge Gateway per hour", "CHF", 0.25*nanoFactor, "Backup")
	stmt.Exec("Standard", "Piece", "Eco Edge Gateway per hour", "CHF", 0, "Eco")
	stmt.Exec("Standard", "Piece",
		"Fast Edge Gateway per hour", "CHF", 0.08333*nanoFactor, "Fast")
	stmt.Exec("Standard", "Piece",
		"Ultra Fast Edge Gateway per hour", "CHF", 0.12500*nanoFactor, "Ultra Fast")
	stmt.Exec("Standard", "Piece",
		"Backup Edge Gateway per hour", "CHF", 0.25*nanoFactor, "Backup")
	stmt.Exec("Advanced", "Piece", "Eco Edge Gateway per hour", "CHF", 0, "Eco")
	stmt.Exec("Advanced", "Piece",
		"Fast Edge Gateway per hour", "CHF", 0.08333*nanoFactor, "Fast")
	stmt.Exec("Advanced", "Piece",
		"Ultra Fast Edge Gateway per hour", "CHF", 0.12500*nanoFactor, "Ultra Fast")
	stmt.Exec("Advanced", "Piece",
		"Backup Edge Gateway per hour", "CHF", 0.25*nanoFactor, "Backup")
	return nil
}

func populateOSCosts(db *sql.DB) error {
	insert := `REPLACE INTO OSCosts(SLA, UsageUnit, Summary,
	CurrencyCode, Nanos, Vendor) VALUES(?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	stmt.Exec("Basic", "VM", "Microsoft Windows OS license for one VM for one hour",
		"CHF", 0.03894*nanoFactor, "Microsoft Windows")
	stmt.Exec("Basic", "VM", "RedHat OS license for one VM for one hour",
		"CHF", 0.05833*nanoFactor, "Red Hat")
	stmt.Exec("Standard", "VM", "Microsoft Windows OS license for one VM for one hour",
		"CHF", 0.03894*nanoFactor, "Microsoft Windows")
	stmt.Exec("Standard", "VM", "RedHat OS license for one VM for one hour",
		"CHF", 0.05833*nanoFactor, "Red Hat")
	stmt.Exec("Advanced", "VM", "Microsoft Windows OS license for one VM for one hour",
		"CHF", 0.03894*nanoFactor, "Microsoft Windows")
	stmt.Exec("Advanced", "VM", "RedHat OS license for one VM for one hour",
		"CHF", 0.05833*nanoFactor, "Red Hat")
	return nil
}

func populateObjectStorageCosts(db *sql.DB) error {
	insert := `REPLACE INTO ObjectStorageCosts(SLA, UsageUnit, Summary,
	CurrencyCode, Nanos) VALUES(?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	// This only exists at "Advanced" SLA
	stmt.Exec("Advanced", "GB usage",
		"1 GB/hour (hourly peak) of Object Storage usage", "CHF", 0.00006*nanoFactor)

	return nil
}

func PopulateDatabase(db *sql.DB) error {
	if err := populateCPUCosts(db); err != nil {
		return err
	}
	if err := populateMemoryCosts(db); err != nil {
		return err
	}
	if err := populateDiskCosts(db); err != nil {
		return err
	}
	if err := populateIpAddrCosts(db); err != nil {
		return err
	}
	if err := populateBandwidthCosts(db); err != nil {
		return err
	}
	if err := populateGatewayCosts(db); err != nil {
		return err
	}
	if err := populateOSCosts(db); err != nil {
		return err
	}
	if err := populateObjectStorageCosts(db); err != nil {
		return err
	}
	return nil
}
