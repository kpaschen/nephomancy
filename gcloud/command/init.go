package command

import (
	"fmt"
	"log"
	"strings"
	"nephomancy/gcloud/cache"
)

type InitCommand struct {
	Command
}

func (*InitCommand) Help() string {
	helpText := `
	"Usage: nephomany gcloud init [options]

	Initialize a new or existing data directory by building or refreshing
	a database of SKUs.

	You should run this when you first start working on a project, and periodically
	afterwards in order to keep Nephomancy's snapshot of billing information up
	to date.

	This command is safe to run multiple times. Subsequent runs will usually be
	faster than the first because they only need to do incremental updates.

	Options:

	  --workingdir=path  Optional: directory under which the data directory should be created.
	                     If not set, defaults to current working directory.

	  --project=PROJECT  Id of a gcloud project. Most initialization will work the same so long
	                     as you use a project under the billing account you are interested in,
			     and for which the relevant APIs are enabled. The user you are authenticated
			     as must have access to the project.
`
        return strings.TrimSpace(helpText)
}

func (*InitCommand) Synopsis() string {
	return "Initializes (builds or refreshes) a database of SKUs."
}

// Run this with
// nephomancy gcloud init --project=binderhub-test-275512
func (c *InitCommand) Run(args []string) int {
	fs := c.Command.defaultFlagSet("gcloudInit")
	fs.Parse(args)

	dbFile, err := c.Command.DbFile()
	if err != nil {
		log.Fatalf("Failed to obtain database filename: %v\n", err)
	}
	err = cache.CreateOrUpdateDatabase(&dbFile)
	if err != nil {
		log.Fatalf("Failed to create database: %v\n", err)
	}
	fmt.Printf("Opened database using file %s\n", dbFile)

	db, err := c.DbHandle()
	if err != nil {
		log.Fatalf("Failed to get database handle: %v\n", err)
	}
	defer db.Close()

	err = cache.PopulateDatabase(db)
	if err != nil {
		log.Fatalf("Failed to populate database: %v\n", err)
	}
	fmt.Println("Populated database.")
	return 0
}
