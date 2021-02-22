package command

import (
	"fmt"
	"log"
	common "nephomancy/common/command"
	"nephomancy/common/registry"
	"nephomancy/dcs/cache"
	"nephomancy/dcs/provider"
	"strings"
)

type InitCommand struct {
	common.Command
}

func (*InitCommand) Help() string {
	helpText := `
	"Usage nephomancy dcs init [options]

	Initialize a new or existing data directory by building or refreshing
	a database of DCS offerings and prices.

	You should run this when you first start working on a project, and after
	upgrading to a version that contains newer DCS cost data.

	This command is safe to run multiple times.

	Options:
	  --workingdir=path	Optional: directory under which the data directory should be. Defaults to current working directory.
`
	return strings.TrimSpace(helpText)
}

func (*InitCommand) Synopsis() string {
	return "Initializes database with DCS cost data."
}

func (c *InitCommand) Run(args []string) int {
	fs := c.Command.DefaultFlagSet("dcsInit")
	fs.Parse(args)

	p, err := registry.GetProvider("dcs")
	if err != nil {
		log.Fatalf("Failed to get DCS provider: %v\n", err)
	}

	dd, _ := c.DataDir()
	prov, _ := p.(*provider.DcsProvider)
	if err := prov.Initialize(dd); err != nil {
		log.Fatalf("Failed to initialize provider: %v\n", err)
	}

	if err := cache.CreateOrUpdateDatabase(prov.DbHandle); err != nil {
		log.Fatalf("Failed to create database: %v\n", err)
	}

	if err := cache.PopulateDatabase(prov.DbHandle); err != nil {
		log.Fatalf("Failed to populate database: %v\n", err)
	}

	fmt.Println("Populated database.")
	return 0
}
