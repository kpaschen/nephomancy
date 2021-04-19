// + build integration

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
        provider, err := registry.GetProvider("gcloud")
        if err != nil {
                t.Errorf("%v", err)
        }
        err = provider.Initialize("../../.nephomancy/data/")
        if err != nil {
                t.Errorf("%v", err)
        }
        startingPoint, err := ioutil.ReadFile("testdata/nephomancy-sample-project.json")
        if err != nil {
                t.Errorf("%v", err)
        }
        wanted, err := ioutil.ReadFile("testdata/gcloud-sample-project.json")
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
        err = provider.FillInProviderDetails(project) // This modifies project.
	if err != nil {
		t.Errorf("%v", err)
	}
	diff := deep.Equal(wantedProject, project)
	if diff != nil {
		options := protojson.MarshalOptions{
			Multiline: true,
			Indent:    " ",
		}
		serializedWantedProject := options.Format(wantedProject)
		t.Errorf("wanted %s but got %s\ndiff: %+v", serializedWantedProject,
                        options.Format(project), diff)
        }

}

func TestGetCost(t *testing.T) {
        provider, err := registry.GetProvider("gcloud")
        if err != nil {
                t.Errorf("%v", err)
        }
        err = provider.Initialize("../../.nephomancy/data/")
        if err != nil {
                t.Errorf("%v", err)
        }
        startingPoint, err := ioutil.ReadFile("testdata/gcloud-sample-project.json")
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
        wantedCosts := make([][]string, 9)

        wantedCosts[0] = []string{
                "Nephomancy sample project", "gcloud", "Sample InstanceSet", "VM memory",
		"1", "2 cpus, 16 gb memory in CH, EMEA",
                "11680 GiBy.h per month", "47.77 USD", "11680 GiBy.h per month",
		"47.77 USD",
        }
        wantedCosts[1] = []string{
                "Nephomancy sample project", "gcloud", "Sample InstanceSet", "VM cpu",
		"1", "2 cpus, 16 gb memory in CH, EMEA",
                "1460 h per month", "44.56 USD", "1460 h per month",
		"44.56 USD",
        }
        wantedCosts[2] = []string{
                "Nephomancy sample project", "gcloud", "Sample InstanceSet",
		"OS license (cpu)", "1", "License for OS Ubuntu",
                "730 h per month", "0.0 USD", "730 h per month", "0.0 USD",
        }
        wantedCosts[3] = []string{
                "Nephomancy sample project", "gcloud", "Sample InstanceSet",
		"OS license (cpu)", "1", "License for OS Ubuntu",
                "730 h per month", "0.0 USD", "730 h per month", "0.0 USD",
        }
        wantedCosts[4] = []string{
                "Nephomancy sample project", "gcloud", "Sample InstanceSet",
		"OS license (memory)", "1", "License for OS Ubuntu",
                "11680 GiBy.h per month", "0.0 USD", "11680 GiBy.h per month", "0.0 USD",
        }
        wantedCosts[5] = []string{
                "Nephomancy sample project", "gcloud", "Sample Disk Set",
		"Disk", "1", "100 GB of SSD in CH, EMEA",
                "100 GiBy/mo", "13.0 USD", "100 GiBy/mo", "13.0 USD",
        }
        wantedCosts[6] = []string{
                "Nephomancy sample project", "gcloud", "",
		"IP Address", "1", "attached to a STANDARD VM in CH, EMEA",
		"730 h", "7.30 USD", "730 h", "7.30 USD",
        }
        wantedCosts[7] = []string{
                "Nephomancy sample project", "gcloud", "default subnetwork",
		"Network", "1", "external egress traffic from europe-west6",
		"unknown", "unknown", "1 Gb", "0.12 USD",
        }
        wantedCosts[8] = []string{
                "Nephomancy sample project", "gcloud", "default subnetwork",
		"Network", "1", "internal egress traffic from europe-west6",
		"unknown", "unknown", "3 Gb", "0.24 USD",
        }
        diff := deep.Equal(wantedCosts, costs)
        if diff != nil {
                t.Errorf("expected costs %+v but got %+v\ndiff: %+v", wantedCosts, costs, diff)
        }
}

