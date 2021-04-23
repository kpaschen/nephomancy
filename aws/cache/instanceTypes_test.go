// +build integration

package cache

import (
	"database/sql"
	"fmt"
	common "nephomancy/common/resources"
	"testing"
)

func getDbHandle() (*sql.DB, error) {
	dbfile := "../../.nephomancy/data/aws/price-cache.db"
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, fmt.Errorf("failed to open database file at %s\n", dbfile)
	}
	return db, nil
}

func TestGetInstanceTypeForSpec(t *testing.T) {
	db, err := getDbHandle()
	if err != nil {
		t.Errorf("could not get db handle: %v", err)
	}
	spec := common.MachineType{
		CpuCount: 1,
		MemoryGb: 1,
	}
	it, regions, err := getInstanceTypeForSpec(db, spec, []string{ "us-east-1" })
	if err != nil {
		t.Errorf("getInstanceTypeForSpec failed: %v\n", err)
	}
	if it != "t3.micro" || len(regions) != 1 || regions[0] != "us-east-1" {
		t.Errorf("unexpected return values from getInstanceTypeForSpec. Wanted t3.micro, [ us-east-1 ] but got %s %v\n",
		it, regions)

	}
}
