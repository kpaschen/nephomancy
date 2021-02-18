// This is currently just a stub.

package provider

import (
	"nephomancy/common/registry"
	"nephomancy/common/resources"
	"os"
	"path/filepath"
)

// AwsProvider implements registry.provider
type AwsProvider struct {
}

var instance registry.Provider = &AwsProvider{}

const name = "aws"

func (g *AwsProvider) FillInProviderDetails(p *resources.Project) error {
	return nil
}

func (*AwsProvider) GetCost(p *resources.Project) ([][]string, error) {
	return nil, nil
}

func init() {
	registry.Register(name, instance)
}

func (p *AwsProvider) Initialize(datadir string) error {
	mydir := filepath.Join(datadir, name)
	err := os.MkdirAll(mydir, 0777)
	if err != nil {
		return err
	}
	return nil
}
