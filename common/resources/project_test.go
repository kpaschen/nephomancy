package resources

import (
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"testing"
)

func TestMakeSampleProject(t *testing.T) {
	wanted, err := ioutil.ReadFile("testdata/nephomancy-sample-project.json")
	if err != nil {
		t.Errorf("%v", err)
	}
	wantedProject := &Project{}
	if err = protojson.Unmarshal(wanted, wantedProject); err != nil {
		t.Errorf("%v", err)
	}
	actual := MakeSampleProject("")
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent:    " ",
	}
	if options.Format(wantedProject) != options.Format(&actual) {
		t.Errorf("wanted %s but got %s", options.Format(wantedProject),
			options.Format(&actual))
	}
}
