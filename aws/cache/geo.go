package cache

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/common/geo"
	"strings"
)

// Returns the common.geo continent for an AWS region.
// This function gets called when verifying whether a given
// AWS region satisfies a location constraint. The converse
// method (RegionsByContinent) should not return GovCloud or
// ISO clusters for NorthAmerica, nor should it return China
// clusters for Asia unless China is explicitly selected as the
// country.
func ContinentFromDisplayName(p1 string, p2 string) geo.Continent {
	switch p1 {
	case "Africa":
		return geo.Africa
	case "Europe":
		return geo.Europe
	case "EU":
		return geo.Europe // Aws is inconsistent about display names
	case "Asia Pacific":
		if p2 == "Sydney" {
			return geo.Australia
		}
		return geo.Asia
	case "China":
		return geo.Asia
	case "Canada":
		return geo.NorthAmerica
	case "Middle East":
		return geo.Asia
	case "US East":
		return geo.NorthAmerica
	case "US West":
		return geo.NorthAmerica
	case "South America":
		return geo.LatinAmerica
	case "AWS GovCloud":
		return geo.NorthAmerica
	case "US ISO East":
		return geo.NorthAmerica
	case "US ISOB East":
		return geo.NorthAmerica
	default:
		return geo.UnknownC
	}
}

func CountryFromDisplayName(p1 string, p2 string) string {
	if p1 == "Europe" {
		switch p2 {
		case "Milan":
			return "IT"
		case "Ireland":
			return "EI"
		case "London":
			return "UK"
		case "Paris":
			return "FR"
		case "Stockholm":
			return "SE"
		case "Frankfurt":
			return "DE"
		default:
			return ""
		}
	} else if p1 == "Asia Pacific" {
		switch p2 {
		case "Sydney":
			return "AU"
		case "Hong Kong":
			return "HK"
		case "Tokyo":
			return "JP"
		case "Singapore":
			return "SG"
		case "Mumbai":
			return "IN"
		case "Seoul":
			return "KR"
		case "Osaka-Local":
			return "JP"
		default:
			return ""
		}
	} else if p1 == "US West" || p1 == "US East" {
		return "US"
	} else if p1 == "AWS GovCloud" || p1 == "US ISO EAST" || p1 == "US ISOB East" {
		return "US"
	} else if p1 == "Canada" {
		return "CA"
	} else if p1 == "China" {
		return "CN"
	} else if p1 == "Africa" {
		if p2 == "Cape Town" {
			return "SA"
		}
		return ""
	} else if p1 == "Middle East" {
		if p2 == "Bahrain" {
			return "BH"
		}
		return ""
	} else if p1 == "South America" {
		if p2 == "Sao Paulo" {
			return "BR"
		}
		return ""
	}
	if !IsWavelengthZone(p2) {
		log.Printf("Unknown region: %s %s\n", p1, p2)
	}
	return ""
}

func IsSpecial(p string) bool {
	return p == "AWS GovCloud" || p == "US ISO EAST" || p == "US ISOB East"
}

// These are special, I don't know how to handle them yet.
func IsWavelengthZone(p string) bool {
	return p == "Verizon" || p == "SKT" || p == "KDDI"
}

func RegionsByContinent(db *sql.DB, continent geo.Continent) ([]string, error) {
	// Retrieve regions from db, then filter out
	// govcloud, iso and cn clusters.
	query := `SELECT ID from Regions WHERE Continent=? AND Special=0 AND Country<>"CN";`
	res, err := db.Query(query, continent.String())
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var regionId string
	regions := make([]string, 0)
	for res.Next() {
		err = res.Scan(&regionId)
		if err != nil {
			return nil, fmt.Errorf("Error scanning row: %v", err)
		}
		regions = append(regions, regionId)
	}
	return regions, nil
}

func RegionsByCountry(db *sql.DB, cc string) ([]string, error) {
	// Retrieve from db. Don't return govcloud or iso clusters.
	query := `SELECT ID from Regions WHERE Country=? AND Special=0;`
	res, err := db.Query(query, cc)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var regionId string
	regions := make([]string, 0)
	for res.Next() {
		err = res.Scan(&regionId)
		if err != nil {
			return nil, fmt.Errorf("Error scanning row: %v", err)
		}
		regions = append(regions, regionId)
	}
	return regions, nil
}

func RegionByDisplayName(db *sql.DB, name string) (string, error) {
	normalName := strings.Replace(name, "EU", "Europe", 1)
	query := `SELECT ID from Regions WHERE DisplayName=?;`
	res, err := db.Query(query, normalName)
	if err != nil {
		return "", err
	}
	defer res.Close()
	var regionId string
	if res.Next() {
		err = res.Scan(&regionId)
		if err != nil {
			return "", fmt.Errorf("Error scanning row: %v", err)
		}
		return regionId, nil
	}
	log.Printf("Did not find a region for display name %s normalized to %s",
		name, normalName)
	return "", nil
}
