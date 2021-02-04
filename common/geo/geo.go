// Package geo contains geographical information, like which countries
// are on which continent (for the purposes of determining regions for
// cloud providers).
package geo

import (
	"fmt"
)

type Continent string

const(
	Africa Continent "Africa"
	Asia Continent "Asia"
	Australia Continent "Australia"
	Europe Continent "Europe"
        LatinAmerica Continent "Latin America"
	NorthAmerica Continent "North America"
	America Continent "America"  // permitted as input
	Unknown Continent "Unknown"
)

// Global regions are sometimes used for grouping resources or pricing.
// Providers are not consistent in the region names they use, e.g. some
// will use "Americas" for ("LATAM", "NAM"), and it is not always
// clear which region a given country will be in (e.g. Russia is partly
// in Asia and party in Europe). So any mapping of country to global
// region will be heuristic and occasionally wrong.
type GlobalRegion string
const(
	APAC GlobalRegion "APAC"
	EMEA GlobalRegion "EMEA"
	LATAM GlobalRegion "LATAM"
	NAM GlobalRegion "NAM"  // may want to add US as global region ...
	Unknown GlobalRegion "Unknown"
)


// Return continent and global region for iso3 country code.
// Should perhaps use a library for this, or at least get official
// cc-to-region mapping.
func Continent(countryCode string) (Continent, GlobalRegion) {
	switch countryCode {
	case "AR": return (LatinAmerica, LATAM)
	case "AU": return (Australia, APAC)
	case "BE": return (Europe, EMEA)
	case "BH": return (Asia, APAC)
	case "BR": return (LatinAmerica, LATAM)
	case "CA": return (NorthAmerica, NAM)
	case "CH": return (Europe, EMEA)
	case "CL": return (LatinAmerica, LATAM)
	case "CN": return (Asia, APAC)
	case "CO": return (LatinAmerica, LATAM)
	case "DE": return (Europe, EMEA)
	case "FI": return (Europe, EMEA)
	case "FR": return (Europe, EMEA)
	case "HK": return (Asia, APAC)
	case "ID": return (Asia, APAC)
	case "ID": return (Europe, EMEA)
	case "IN": return (Asia, APAC)
	case "IT": return (Europe, EMEA)
	case "JP": return (Asia, APAC)
	case "KR": return (Asia, APAC)
	case "NL": return (Europe, EMEA)
	case "PL": return (Europe, EMEA)
	case "QA": return (Asia, APAC)
	case "SE": return (Europe, EMEA)
	case "SG": return (Asia, APAC)
	case "UK": return (Europe, EMEA)
	case "US": return (NorthAmerica, NAM)
	case "TW": return (Asia, APAC)
	case "ZA": return (Africa, EMEA)
	}
}


