package resources

import (
	"nephomancy/common/geo"
	"testing"
)

func TestCountryCodeToLocaton(t *testing.T) {
	countryCodes := []string{
		"AR", "DE", "CH", "JP", "ZA",
	}
	locs := []Location{
		Location{
			GlobalRegion: geo.LATAM.String(),
			Continent:    geo.LatinAmerica.String(),
			CountryCode:  "AR",
		},
		Location{
			GlobalRegion: geo.EMEA.String(),
			Continent:    geo.Europe.String(),
			CountryCode:  "DE",
		},
		Location{
			GlobalRegion: geo.EMEA.String(),
			Continent:    geo.Europe.String(),
			CountryCode:  "CH",
		},
		Location{
			GlobalRegion: geo.APAC.String(),
			Continent:    geo.Asia.String(),
			CountryCode:  "JP",
		},
		Location{
			GlobalRegion: geo.EMEA.String(),
			Continent:    geo.Africa.String(),
			CountryCode:  "ZA",
		},
	}
	for idx, cc := range countryCodes {
		l, err := CountryCodeToLocation(cc)
		if err != nil {
			t.Errorf("failed to get location for country code %s\n", cc)
		}
		if !LocationEqual(l, locs[idx]) {
			t.Errorf("wrong location %v for country code %s\n", l, cc)
		}
	}
}

func TestCheckLocation(t *testing.T) {
	loc := Location{
		GlobalRegion: geo.EMEA.String(),
		Continent:    geo.Europe.String(),
		CountryCode:  "CH",
	}
	spec := Location{
		GlobalRegion: geo.EMEA.String(),
	}
	if err := CheckLocation(loc, spec); err != nil {
		t.Errorf("%v", err)
	}
	spec.Continent = geo.Europe.String()
	if err := CheckLocation(loc, spec); err != nil {
		t.Errorf("%v", err)
	}
	spec.CountryCode = "CH"
	if err := CheckLocation(loc, spec); err != nil {
		t.Errorf("%v", err)
	}
	spec.CountryCode = "DE"
	if err := CheckLocation(loc, spec); err == nil {
		t.Errorf("%v should not match %v\n", loc, spec)
	}
	spec = Location{
		GlobalRegion: geo.APAC.String(),
	}
	if err := CheckLocation(loc, spec); err == nil {
		t.Errorf("%v should not match %v\n", loc, spec)
	}
	spec = Location{
		GlobalRegion: geo.EMEA.String(),
		Continent:    geo.Africa.String(),
	}
	if err := CheckLocation(loc, spec); err == nil {
		t.Errorf("%v should not match %v\n", loc, spec)
	}
}
