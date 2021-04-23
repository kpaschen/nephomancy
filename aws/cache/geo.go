package cache

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"log"
	cgeo "nephomancy/common/geo"
	common "nephomancy/common/resources"
	"regexp"
)

// Some regions show up in the list but aren't actually supported yet.
var UnsupportedRegions = map[string]bool{
	"eu-south-1":     true,
	"af-south-1":     true,
	"ap-east-1":      true,
	"me-south-1":     true,
	"cn-north-1":     true,
	"cn-northwest-1": true,
}

type Region struct {
	Description     string
	Country         string
	Continent       string
	IsSpecialRegion bool
}

var Regions map[string]Region

func init() {
	initializeRegions()
}

// Internal method for initializing the Regions map.
// Not threadsafe.
func initializeRegions() {
	Regions = make(map[string]Region)
	re := regexp.MustCompile(`([a-zA-Z ]+) \(([^\)]+)\)`)
	for _, partition := range endpoints.DefaultPartitions() {
		for _, rg := range partition.Regions() {
			regionId := rg.ID()
			desc := rg.Description()
			places := re.FindStringSubmatch(desc)
			continent := ""
			country := ""
			specialRegion := false
			if len(places) == 0 {
				continent = ContinentFromDisplayName(desc, "").String()
				specialRegion = true
			} else {
				continent = ContinentFromDisplayName(
					places[1], places[2]).String()
				country = CountryFromDisplayName(
					places[1], places[2])
				if IsSpecial(places[1]) {
					specialRegion = true
				}
			}
			Regions[regionId] = Region{
				Description:     desc,
				Country:         country,
				Continent:       continent,
				IsSpecialRegion: specialRegion,
			}
		}
	}
}

// Returns the common.geo continent for an AWS region.
// This function gets called when verifying whether a given
// AWS region satisfies a location constraint. The converse
// method (RegionsByContinent) should not return GovCloud or
// ISO clusters for NorthAmerica, nor should it return China
// clusters for Asia unless China is explicitly selected as the
// country.
func ContinentFromDisplayName(p1 string, p2 string) cgeo.Continent {
	switch p1 {
	case "Africa":
		return cgeo.Africa
	case "Europe":
		return cgeo.Europe
	case "EU":
		return cgeo.Europe // Aws is inconsistent about display names
	case "Asia Pacific":
		if p2 == "Sydney" {
			return cgeo.Australia
		}
		return cgeo.Asia
	case "China":
		return cgeo.Asia
	case "Canada":
		return cgeo.NorthAmerica
	case "Middle East":
		return cgeo.Asia
	case "US East":
		return cgeo.NorthAmerica
	case "US West":
		return cgeo.NorthAmerica
	case "South America":
		return cgeo.LatinAmerica
	case "AWS GovCloud":
		return cgeo.NorthAmerica
	case "US ISO East":
		return cgeo.NorthAmerica
	case "US ISOB East":
		return cgeo.NorthAmerica
	default:
		return cgeo.UnknownC
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

func AllRegions(onlySupported bool) []string {
	ret := make([]string, 0, len(Regions))
	for rid, _ := range Regions {
		if onlySupported && UnsupportedRegions[rid] {
			continue
		}
		ret = append(ret, rid)
	}
	return ret
}

func RegionsByContinent(continent cgeo.Continent) []string {
	ret := make([]string, 0, len(Regions))
	for rid, region := range Regions {
		if region.Continent == continent.String() {
			ret = append(ret, rid)
		}
	}
	return ret
}

func CountryByRegion(region string) string {
	r, ok := Regions[region]
	if ok {
		return r.Country
	}
	return "Unknown"
}

func RegionsByCountry(cc string) []string {
	ret := make([]string, 0, len(Regions))
	for rid, region := range Regions {
		if region.Country == cc {
			ret = append(ret, rid)
		}
	}
	return ret
}

func RegionByDisplayName(name string) string {
	for rid, region := range Regions {
		if region.Description == name {
			return rid
		}
	}
	return ""
}

// Returns all regions consistent with loc.
// If preferred is not empty, and is contained in the possible regions,
// return only preferred region.
func RegionsForLocation(loc common.Location, preferred string) []string {
	var regions []string
	if loc.CountryCode != "" {
		regions = RegionsByCountry(loc.CountryCode)
	} else if loc.Continent != "" {
		regions = RegionsByContinent(cgeo.ContinentFromString(loc.Continent))
	} else if loc.GlobalRegion != "" {
		regions = make([]string, 0)
		continents := cgeo.GetContinents(cgeo.RegionFromString(loc.GlobalRegion))
		for _, c := range continents {
			regions = append(regions, RegionsByContinent(c)...)
		}
	}
	if len(regions) == 0 {
		if preferred == "" {
			return RegionsByCountry("US") // default to US
		} else {
			regions = RegionsByCountry("US")
		}
	}
	if preferred != "" {
		for _, r := range regions {
			if r == preferred {
				return []string{r}
			}
		}
	}
	return regions
}
