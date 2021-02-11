package registry

import (
	"fmt"
	"nephomancy/common/resources/"
)

type Provider interface {
	ReconcileSpecAndProviderDetails(*resources.Project) error
	GetCost(*resources.Project) ([][]string, error)
	name() string
}

var Registry map[string]Provider

func init() {
	Registry = make(map[string]Provider)
}

func Register(name string, Provider p) {
	if Registry[name] == nil {
		Registry[name] = p
	} else {
		fmt.Printf("Duplicate provider for name %s\n", name)
	}
}
