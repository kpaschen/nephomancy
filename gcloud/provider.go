package gcloud

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/common/registry"
	"nephomancy/common/resources"
	"nephomancy/gcloud/cache"
	"os"
	"path/filepath"
	"runtime"
)

// GcloudProvider implements registry.provider
type GcloudProvider struct {
	// db handle for lookups
	dbHandle *sql.DB
}

var instance registry.Provider = &GcloudProvider{}

const name = "gcloud"

func (g *GcloudProvider) ReconcileSpecAndProviderDetails(p *resources.Project) error {
	if g.dbHandle == nil {
		return fmt.Errorf("Provider has not been initialized\n")
	}
	return cache.ReconcileSpecAndAssets(g.dbHandle, p)
}

func (*GcloudProvider) GetCost(p *resources.Project) ([][]string, error) {
	return nil, nil
}

func init() {
	registry.Register(name, instance)
}

func (p *GcloudProvider) Initialize(datadir string) error {
	if p.dbHandle != nil {
		return nil
	}
	mydir := filepath.Join(datadir, name)
	err := os.MkdirAll(mydir, 0777)
	if err != nil {
		return err
	}
	dbfile := filepath.Join(mydir, "sku-cache.db")
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	p.dbHandle = db
	if p.dbHandle == nil {
		return fmt.Errorf("Failed to open a database file at %s\n", dbfile)
	}
	runtime.SetFinalizer(p, finalizer)
	return nil
}

func finalizer(p *GcloudProvider) {
	if p.dbHandle != nil {
		if err := p.dbHandle.Close(); err != nil {
			log.Printf("Failure in finalizer: %v\n", err)
		}
	}
}
