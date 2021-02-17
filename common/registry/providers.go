package registry

import (
	"fmt"
	"nephomancy/common/resources"
)

type Provider interface {
	ReconcileSpecAndProviderDetails(*resources.Project) error
	GetCost(*resources.Project) ([][]string, error)
	Initialize(datadir string) error
}

var Registry map[string]Provider

func init() {
	Registry = make(map[string]Provider)
}

func Register(name string, p Provider) {
	if Registry[name] == nil {
		Registry[name] = p
	} else {
		fmt.Printf("Duplicate provider for name %s\n", name)
	}
}

func GetProvider(name string) (Provider, error) {
	if Registry[name] == nil {
		return dummyProvider, fmt.Errorf("unknown provider: %s", name)
	}
	return Registry[name], nil
}

type emptyProvider struct {}

func (emptyProvider) ReconcileSpecAndProviderDetails(*resources.Project) error {
	return nil
}

func (emptyProvider) GetCost(*resources.Project) ([][]string, error) {
	return nil, nil
}

func (emptyProvider) Initialize(string) error {
	return nil
}

var dummyProvider emptyProvider
