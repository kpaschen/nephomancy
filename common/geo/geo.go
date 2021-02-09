// Package geo contains geographical information, like which countries
// are on which continent (for the purposes of determining regions for
// cloud providers).
package geo

type Continent int

const (
	UnknownC Continent = iota
	Africa
	Asia
	Australia
	Europe
	LatinAmerica
	NorthAmerica
	America
)

func (c Continent) String() string {
	return []string{
		"Unknown",
		"Africa",
		"Asia",
		"Australia",
		"Europe",
		"Latin America",
		"North America",
		"America",
	}[c]
}

func ContinentFromString(name string) Continent {
	switch name {
	case "Africa":
		return Africa
	case "Asia":
		return Asia
	case "Australia":
		return Australia
	case "Europe":
		return Europe
	case "Latin America":
		return LatinAmerica
	case "North America":
		return NorthAmerica
	default:
		return UnknownC
	}
}

// Global regions are sometimes used for grouping resources or pricing.
// Providers are not consistent in the region names they use, e.g. some
// will use "Americas" for ("LATAM", "NAM"), and it is not always
// clear which region a given country will be in (e.g. Russia is partly
// in Asia and party in Europe). So any mapping of country to global
// region will be heuristic and occasionally wrong.
type GlobalRegion int

const (
	UnknownG GlobalRegion = iota
	APAC
	EMEA
	LATAM
	NAM
)

func (g GlobalRegion) String() string {
	return []string{
		"Unknown", "APAC", "EMEA", "LATAM", "NAM",
	}[g]
}

func RegionFromString(name string) GlobalRegion {
	switch name {
	case "APAC":
		return APAC
	case "EMEA":
		return EMEA
	case "LATAM":
		return LATAM
	case "NAM":
		return NAM
	default:
		return UnknownG
	}
}

func GetContinents(gr GlobalRegion) []Continent {
	switch gr {
	case UnknownG:
		return []Continent{UnknownC}
	case APAC:
		return []Continent{Asia, Australia}
	case EMEA:
		return []Continent{Europe, Africa} // FIXME
	case LATAM:
		return []Continent{LatinAmerica}
	case NAM:
		return []Continent{NorthAmerica}
	default:
		return []Continent{}
	}
}

// Return continent and global region for iso3 country code.
// Should perhaps use a library for this, or at least get official
// cc-to-region mapping.
func GetContinent(countryCode string) (Continent, GlobalRegion) {
	switch countryCode {
	case "AR":
		return LatinAmerica, LATAM
	case "AU":
		return Australia, APAC
	case "BE":
		return Europe, EMEA
	case "BH":
		return Asia, APAC
	case "BR":
		return LatinAmerica, LATAM
	case "CA":
		return NorthAmerica, NAM
	case "CH":
		return Europe, EMEA
	case "CL":
		return LatinAmerica, LATAM
	case "CN":
		return Asia, APAC
	case "CO":
		return LatinAmerica, LATAM
	case "DE":
		return Europe, EMEA
	case "FI":
		return Europe, EMEA
	case "FR":
		return Europe, EMEA
	case "HK":
		return Asia, APAC
	case "ID":
		return Asia, APAC
	case "IE":
		return Europe, EMEA
	case "IN":
		return Asia, APAC
	case "IT":
		return Europe, EMEA
	case "JP":
		return Asia, APAC
	case "KR":
		return Asia, APAC
	case "NL":
		return Europe, EMEA
	case "PL":
		return Europe, EMEA
	case "QA":
		return Asia, APAC
	case "SE":
		return Europe, EMEA
	case "SG":
		return Asia, APAC
	case "UK":
		return Europe, EMEA
	case "US":
		return NorthAmerica, NAM
	case "TW":
		return Asia, APAC
	case "ZA":
		return Africa, EMEA
	default:
		return UnknownC, UnknownG
	}
}
