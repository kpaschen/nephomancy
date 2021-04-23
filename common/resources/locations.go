package resources

import (
	"fmt"
	"nephomancy/common/geo"
)

// Resolves a country code (2-letter ISO) to a Location.
func CountryCodeToLocation(cc string) (Location, error) {
	if cc == "Unknown" {
		return Location{}, fmt.Errorf("unknown country code %s", cc)
	}
	continent, gr := geo.GetContinent(cc)
	if continent == geo.UnknownC {
		return Location{},
		fmt.Errorf("no continent known for country %s", cc)
	}
	return Location{
		GlobalRegion: gr.String(),
		Continent: continent.String(),
		CountryCode: cc,
	}, nil
}

// Returns nil if loc is consistent with spec, an error otherwise.
func CheckLocation(loc Location, spec Location) error {
	if spec.GlobalRegion != "" && spec.GlobalRegion != loc.GlobalRegion {
		return fmt.Errorf("spec global region %s does not match details: %s",
		spec.GlobalRegion, loc.GlobalRegion)
	}
	if spec.Continent != "" && spec.Continent != loc.Continent {
		return fmt.Errorf("spec continent %s does not match details: %s",
		spec.Continent, loc.Continent)
	}
	if spec.CountryCode != "" && spec.CountryCode != loc.CountryCode {
		return fmt.Errorf("spec country %s does not match details: %s",
		spec.CountryCode, loc.CountryCode)
	}
	return nil
}
