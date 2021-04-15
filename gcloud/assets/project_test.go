package assets

import (
	"bufio"
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"io/ioutil"
	common "nephomancy/common/resources"
	"os"
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
			Name: name,
			AssetType: "service/" + atype,
			ResourceAsJson: string(line),
		}
		rt = append(rt, sa)

	}
	return rt, nil
}


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
		Indent: " ",
	}
        wanted, err := ioutil.ReadFile("testdata/project")
	if err != nil {
		t.Errorf("%v", err)
	}
	wantedProject := &common.Project{}
	if err = protojson.Unmarshal(wanted, wantedProject); err != nil {
		t.Errorf("%v", err)
	}
	// Somehow reflect.DeepEqual finds a difference between projects that
	// should be the same, but formatting both yields the same string.
	if options.Format(wantedProject) != options.Format(p) {
		t.Errorf("wanted %s \n but got %s", options.Format(wantedProject),
		options.Format(p))
	}

}
