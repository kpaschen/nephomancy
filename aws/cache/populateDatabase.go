package cache

import (
	"database/sql"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	_ "github.com/mattn/go-sqlite3"
	"nephomancy/aws/resources"
	"regexp"
)

func PopulateDatabase(db *sql.DB) error {
	if err := populateRegions(db); err != nil {
		return err
	}
	//if err := getPricesInBulk(db, "", "index.json"); err != nil {
	//	return err
	//}
	return nil
}

func populateRegions(db *sql.DB) error {
	insert := `REPLACE INTO Regions(ID, DisplayName, Country, Continent, Special)
	VALUES(?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`([a-zA-Z ]+) \(([^\)]+)\)`)

	for _, partition := range endpoints.DefaultPartitions() {
		for _, rg := range partition.Regions() {
			regionId := rg.ID()
			desc := rg.Description()
			places := re.FindStringSubmatch(desc)
			continent := ""
			country := ""
			specialRegion := 0
			if len(places) == 0 {
				continent = ContinentFromDisplayName(desc, "").String()
				specialRegion = 1
			} else {
				continent = ContinentFromDisplayName(
					places[1], places[2]).String()
				country = CountryFromDisplayName(
					places[1], places[2])
				if IsSpecial(places[1]) {
					specialRegion = 1
				}
			}
			_, err = stmt.Exec(regionId, desc, country,
				continent, specialRegion)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func InsertInstanceType(db *sql.DB, itype resources.InstanceType) error {
	insert := `REPLACE INTO InstanceTypes(InstanceType, CPU, Memory,
	GPU, NetworkPerformance, StorageType, StorageAmount, SupportsSpot,
	SupportsOndemand) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	var instanceStorage string
	if itype.InstanceStorageSupported {
		instanceStorage = itype.InstanceStorageType
	} else {
		instanceStorage = "ebs"
	}
	var supportsSpot int
	var supportsOnDemand int
	for _, usageClass := range itype.SupportedUsageClasses {
		if usageClass == "spot" {
			supportsSpot = 1
		} else if usageClass == "on-demand" {
			supportsOnDemand = 1
		}
	}
	_, err = stmt.Exec(itype.Name, itype.DefaultCpuCount,
	itype.MemoryMiB, itype.GpuCount, itype.NetworkPerformanceGbit,
	instanceStorage, itype.InstanceStorageMaxSizeGb, supportsSpot, supportsOnDemand)
	if err != nil {
		return err
	}
	insert = `REPLACE INTO CoreCount(InstanceType, CoreCount) VALUES(?, ?);`
	stmt, err = db.Prepare(insert)
	if err != nil {
		return err
	}
	for _, cc := range itype.ValidCores {
		_, err = stmt.Exec(itype.Name, cc)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertInstanceTypesForRegion(db *sql.DB, itypes []string, region string) error {
	insert := `REPLACE INTO InstanceTypeByRegion (InstanceType, Region)
	VALUES(?, ?);`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return err
	}
	for _, it := range itypes {
		_, err = stmt.Exec(it, region)
		if err != nil {
			return err
		}
	}
	return nil
}
