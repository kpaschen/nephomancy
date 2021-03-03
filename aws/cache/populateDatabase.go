package cache

import (
	"database/sql"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
)

func PopulateDatabase(db *sql.DB) error {
	return populateRegions(db)
}

func populateRegions(db *sql.DB) error {
	insert := `REPLACE INTO Regions(ID, DisplayName, Country, Continent, Special)
	VALUES(?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	_ = stmt
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
