package assets

import (
	"bufio"
	"encoding/json"
	"github.com/go-test/deep"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"io/ioutil"
	common "nephomancy/common/resources"
	"os"
	"sort"
	"testing"
)

func readAssetsFile(filename string) ([]SmallAsset, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var line string
	rt := make([]SmallAsset, 0)
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		var a map[string]interface{}
		json.Unmarshal([]byte(line), &a)
		dm, _ := a["data"].(map[string]interface{})
		name, _ := dm["name"].(string)
		atype, _ := a["discoveryName"].(string)
		sa := SmallAsset{
			Name:           name,
			AssetType:      "service/" + atype,
			ResourceAsJson: string(line),
		}
		rt = append(rt, sa)

	}
	return rt, nil
}

type SortableDiskSet []*common.DiskSet

func(a SortableDiskSet) Len() int { return len(a) }
func(a SortableDiskSet) Less(i, j int) bool { return a[i].Name < a[j].Name }
func(a SortableDiskSet) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func TestBuildProject(t *testing.T) {
	assets, err := readAssetsFile("testdata/assets")
	if err != nil {
		t.Errorf("%v", err)
	}
	p, err := BuildProject(assets)
	if err != nil {
		t.Errorf("%v", err)
	}
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent:    " ",
	}
	wanted, err := ioutil.ReadFile("testdata/project")
	if err != nil {
		t.Errorf("%v", err)
	}
	wantedProject := &common.Project{}
	if err = protojson.Unmarshal(wanted, wantedProject); err != nil {
		t.Errorf("%v", err)
	}
	// protojson unmarshals both empty lists and nil into nil. BuildProject deliberately
	// creates empty lists for some types. So the following lines just avoid a false
	// positive diff on deep.Equal.
	if wantedProject.Networks[0].Subnetworks[0].Gateways == nil {
		wantedProject.Networks[0].Subnetworks[0].Gateways = make([]*common.Gateway, 0)
	}
	// deep.Equal needs the sort order of slices to be the same.
	sort.Sort(SortableDiskSet(wantedProject.DiskSets))
	sort.Sort(SortableDiskSet(p.DiskSets))

	diff := deep.Equal(wantedProject, p)
	if diff != nil {
		t.Errorf("wanted %s \n but got %s\ndiff is %+v", options.Format(wantedProject),
			options.Format(p), diff)
	}

}
