// Cost modeling for Swisscom DCS.

package provider

import (
	"nephomancy/common/registry"
	"nephomancy/common/resources"
	"os"
	"path/filepath"
)

// DcsProvider implements registry.provider
type DcsProvider struct {
}

var instance registry.Provider = &DcsProvider{}

const name = "dcs"

func (g *DcsProvider) FillInProviderDetails(p *resources.Project) error {
	return nil
}

func (*DcsProvider) GetCost(p *resources.Project) ([][]string, error) {
	return nil, nil
}

func init() {
	registry.Register(name, instance)
}

func (p *DcsProvider) Initialize(datadir string) error {
	mydir := filepath.Join(datadir, name)
	err := os.MkdirAll(mydir, 0777)
	if err != nil {
		return err
	}
	return nil
}
