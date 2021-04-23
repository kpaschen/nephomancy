package geo

import (
	"testing"
)

func TestContinentFromString(t *testing.T) {
	names := []string{
		"Africa", "Asia", "Australia", "Europe", "LatinAmerica", "NorthAmerica", "Antarctica",
	}
	continents := []Continent{
		Africa, Asia, Australia, Europe, LatinAmerica, NorthAmerica, UnknownC,
	}
	for idx, name := range names {
		c := ContinentFromString(name)
		if c != continents[idx] {
			t.Errorf("wrong continent %s for input %s\n",
				c.String(), name)
		}
	}
}

func TestRegionFromString(t *testing.T) {
	names := []string{
		"APAC", "EMEA", "LATAM", "NAM", "NOWHERE",
	}
	regions := []GlobalRegion{
		APAC, EMEA, LATAM, NAM, UnknownG,
	}
	for idx, name := range names {
		r := RegionFromString(name)
		if r != regions[idx] {
			t.Errorf("wrong region %s for input %s\n", r.String(), name)
		}
	}
}

func TestGetContinents(t *testing.T) {
	regions := []GlobalRegion{
		APAC, EMEA, LATAM, NAM, UnknownG,
	}
	continents := [][]Continent{
		[]Continent{Asia, Australia},
		[]Continent{Europe, Africa},
		[]Continent{LatinAmerica},
		[]Continent{NorthAmerica},
		[]Continent{UnknownC},
	}
	for idx, r := range regions {
		cs := GetContinents(r)
		if len(cs) != len(continents[idx]) {
			t.Errorf("wrong continents %v for region %s\n", cs, r.String())
		}
		for i, x := range cs {
			if x != continents[idx][i] {
				t.Errorf("wrong continents %v for region %s\n", cs, r.String())
			}
		}
	}
}
