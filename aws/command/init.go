package command

import (
	"fmt"
	"log"
	"nephomancy/aws/cache"
	"nephomancy/aws/provider"
	common "nephomancy/common/command"
	"nephomancy/common/registry"
	"strings"
)

type InitCommand struct {
	common.Command
}

func (*InitCommand) Help() string {
	helpText := `
	"Usage nephomancy aws init [options]

	Initialize a new or existing data directory by building or refreshing
	a database of AWS offerings and prices.

	You should run this when you first start working on a project, and
	whenever you think AWS pricing may have changed.

	This command is safe to run multiple times.

	Options:
	  --workingdir=path	Optional: directory under which the data directory should be. Defaults to current working directory.
`
	return strings.TrimSpace(helpText)
}

func (*InitCommand) Synopsis() string {
	return "Initializes database with AWS cost data."
}

func (c *InitCommand) Run(args []string) int {
	fs := c.Command.DefaultFlagSet("awsInit")
	fs.Parse(args)

	p, err := registry.GetProvider("aws")
	if err != nil {
		log.Fatalf("Failed to get AWS provider: %v\n", err)
	}
	dd, _ := c.DataDir()
	prov, _ := p.(*provider.AwsProvider)
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
