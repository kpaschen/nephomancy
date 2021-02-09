package assets

import (
	"encoding/json"
	"testing"
)

func TestRegions(t *testing.T) {
	dMap := map[string]interface{}{
		"data": map[string]interface{}{
			"zone": "europe-west1-b",
		},
	}
	dBytes, _ := json.Marshal(dMap)
	a := SmallAsset{
		ResourceAsJson: string(dBytes),
	}
	r, err := a.regions()
	if err != nil || len(r) != 1 || r[0] != "europe-west1" {
		t.Errorf("Expected region %s but got %s, %v\n", "europe-west1", r, err)
	}

	dMap = map[string]interface{}{
		"data": map[string]interface{}{
			"storageLocations": []string{
				"europe-west2",
				"northamerica-east1",
			},
		},
	}
	dBytes, _ = json.Marshal(dMap)
	a = SmallAsset{
		ResourceAsJson: string(dBytes),
	}
	r, err = a.regions()
	if err != nil || len(r) != 2 {
		t.Errorf("Expected to get two regions but got %v with len %d, %v\n",
			r, len(r), err)
	}
	if r[0] != "europe-west2" || r[1] != "northamerica-east1" {
		t.Errorf("Expected to get europe-west2 and northamerica-east1 but got %v\n", r)
	}
}

func TestScheduling(t *testing.T) {
	dMap := map[string]interface{}{
		"data": map[string]interface{}{
			"scheduling": map[string]interface{}{
				"preemptible": true,
			},
		},
	}
	dBytes, _ := json.Marshal(dMap)
	a := SmallAsset{
		ResourceAsJson: string(dBytes),
	}
	scheduling, err := a.scheduling()
	if err != nil || scheduling != "Preemptible" {
		t.Errorf("Expected scheduling to be preemptible but got %s, %v\n",
			scheduling, err)
	}
	dMap = map[string]interface{}{
		"data": map[string]interface{}{
			"scheduling": map[string]interface{}{
				"preemptible": false,
			},
		},
	}
	dBytes, _ = json.Marshal(dMap)
	a = SmallAsset{
		ResourceAsJson: string(dBytes),
	}
	scheduling, err = a.scheduling()
	if err != nil || scheduling != "OnDemand" {
		t.Errorf("Expected scheduling to be OnDemand but got %s, %v\n",
			scheduling, err)
	}
	dMap = map[string]interface{}{
		"data": map[string]interface{}{
			"scheduling": map[string]interface{}{
				"otherEntry": false,
			},
		},
	}
	dBytes, _ = json.Marshal(dMap)
	a = SmallAsset{
		ResourceAsJson: string(dBytes),
	}
	scheduling, err = a.scheduling()
	if err != nil || scheduling != "" {
		t.Errorf("Expected scheduling to be empty but got %s, %v\n",
			scheduling, err)
	}
}
