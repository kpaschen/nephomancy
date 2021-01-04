package cache

import (
	"testing"
)

func TestFromJson(t *testing.T) {
	base := `{"baseUnit":"By.s","baseUnitConversionFactor":2875910101401600,
	"baseUnitDescription":"byte second", "displayQuantity":1,
	"tieredRates":[{"unitPrice":{"currencyCode":"USD"}},
	{"startUsageAmount":30, "unitPrice":{"currencyCode":"USD","nanos":40000000}}],
	"usageUnit":"GiBy.mo","usageUnitDescription":"gibibyte month"}`

	pi, err := FromJson(&base)
	if err != nil {
		t.Errorf("Expected parsing of %s to work but got %v\n",
		base, err)
	}
	if pi.BaseUnitConversionFactor != 2875910101401600 {
		t.Errorf("Misparsed base unit conversionfactor. Got %d expected %d\n",
		pi.BaseUnitConversionFactor, 2875910101401600)
	}
	if len(pi.TieredRates) != 2 {
		t.Errorf("tiered rates: expected 2 but found %d\n",
		len(pi.TieredRates))
	}
	if pi.TieredRates[1].Nanos != 40000000 {
		t.Errorf("second tier rate should have nanos 40000000 but has %d\n",
		pi.TieredRates[1].Nanos)
	}
}
