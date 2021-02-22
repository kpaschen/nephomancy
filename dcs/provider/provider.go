// Cost modeling for Swisscom DCS.

package provider

import (
	"database/sql"
	"fmt"
	"log"
	"nephomancy/common/registry"
	"nephomancy/common/resources"
	"nephomancy/dcs/cache"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

// DcsProvider implements registry.provider
type DcsProvider struct {
	DbHandle *sql.DB
}

var instance registry.Provider = &DcsProvider{}

const name = "dcs"

func (d *DcsProvider) FillInProviderDetails(p *resources.Project) error {
	if d.DbHandle == nil {
		return fmt.Errorf("Provider has not been initialized\n")
	}
	return cache.FillInProviderDetails(d.DbHandle, p)
}

func (d *DcsProvider) GetCost(p *resources.Project) ([][]string, error) {
	if d.DbHandle == nil {
		return nil, fmt.Errorf("Provider has not been initialized\n")
	}
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
	dbfile := filepath.Join(mydir, "price-cache.db")
	_, err = os.OpenFile(dbfile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	if db == nil {
		return fmt.Errorf("Failed to open a database file at %s\n", dbfile)
	}
	p.DbHandle = db
	runtime.SetFinalizer(p, finalizer)
	return nil
}

func finalizer(p * DcsProvider) {
	if p.DbHandle != nil {
		if err := p.DbHandle.Close(); err != nil {
			log.Printf("Failure in finalizer: %v\n", err)
		}
	}
}
