// +build integration

package provider

import (
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"github.com/go-test/deep"
	"nephomancy/common/registry"
	common "nephomancy/common/resources"
	"testing"
)

func TestFillInProviderDetails(t *testing.T) {
	provider, err := registry.GetProvider("dcs")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = provider.Initialize("../.nephomancy/data/")
	if err != nil {
		t.Errorf("%v", err)
	}
	startingPoint, err := ioutil.ReadFile("testdata/nephomancy-sample-project.json")
	if err != nil {
		t.Errorf("%v", err)
	}
	wanted, err := ioutil.ReadFile("testdata/dcs-sample-project.json")
	if err != nil {
		t.Errorf("%v", err)
	}
	wantedProject := &common.Project{}
	if err = protojson.Unmarshal(wanted, wantedProject); err != nil {
		t.Errorf("%v", err)
	}
	project := &common.Project{}
	if err = protojson.Unmarshal(startingPoint, project); err != nil {
		t.Errorf("%v", err)
	}
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent:    " ",
	}
	serializedWantedProject := options.Format(wantedProject)
	err = provider.FillInProviderDetails(project) // This modifies project.
	if serializedWantedProject != options.Format(project) {
		t.Errorf("wanted %s but got %s", serializedWantedProject,
			options.Format(project))
	}
}

func TestGetCost(t *testing.T) {
	provider, err := registry.GetProvider("dcs")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = provider.Initialize("../../.nephomancy/data/")
	if err != nil {
		t.Errorf("%v", err)
	}
	startingPoint, err := ioutil.ReadFile("testdata/dcs-sample-project.json")
	if err != nil {
		t.Errorf("%v", err)
	}
	project := &common.Project{}
	if err = protojson.Unmarshal(startingPoint, project); err != nil {
		t.Errorf("%v", err)
	}
	costs, err := provider.GetCost(project)
	if err != nil {
		t.Errorf("%v", err)
	}
	wantedCosts := make([][]string, 7)
	wantedCosts[0] = []string{
		"VM CPU", "1", "2 cpus, 16 gb memory in CH, EMEA",
		"1440 h per month", "39.99 CHF", "1460 h per month", "40.54 CHF",
	}
        wantedCosts[1] = []string{
		"VM RAM", "16", "2 cpus, 16 gb memory in CH, EMEA",
		"11520 GB-hours per month", "167.96 CHF", "11680 GB-hours per month",
		"170.29 CHF",
	}
	wantedCosts[2] = []string{
		"VM OS", "Red Hat", "2 cpus, 16 gb memory in CH, EMEA",
		"720 h per month", "42.00 CHF", "730 h per month", "42.58 CHF",
	}
	wantedCosts[3] = []string{
		"Disk", "1", "100 GB of SSD in CH, EMEA", "100 GB for 720 h per month",
		"0.19 CHF", "100 GB for 730 h per month", "0.20 CHF",
	}
	wantedCosts[4] = []string{
		"IP Addresses", "1", "ip addresses: /29 cidr",
		"1 addresses for 720 h per month", "0.81 CHF",
		"1 addresses for 720 h per month", "0.81 CHF",
	}
	wantedCosts[5] = []string{
		"Bandwidth", "150", "Bandwidth in MBit/s", "150 MBit/s for 720 h per month",
		"224.64 CHF", "150 MBit/s for 720 h per month", "224.64 CHF",
	}
	wantedCosts[6] = []string{
		"Gateway", "1", "Gateway of type Eco", "for 720 h per month", "0.00 CHF",
		"for 720 h per month", "0.00 CHF",
	}
	diff := deep.Equal(wantedCosts, costs)
	if diff != nil {
		t.Errorf("expected costs %+v but got %+v\ndiff: %+v", wantedCosts, costs, diff)
	}
}
