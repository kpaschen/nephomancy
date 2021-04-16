// +build integration

package provider

import (
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
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
