package command

import (
	"fmt"
	"log"
	"nephomancy/aws/cache"
	"nephomancy/aws/ec2"
	"nephomancy/aws/provider"
	"nephomancy/aws/resources"
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
	// This populates the region table based on the default partitions.
	if err := cache.PopulateDatabase(prov.DbHandle); err != nil {
		log.Fatalf("Failed to populate database: %v\n", err)
	}

	sendToDb := make(chan *resources.InstanceType, 1)
	okFromDb := make(chan error, 1)
	retval := make(chan error, 1)
	defer close(sendToDb)
	defer close(okFromDb)

	// Assume us-east-1 has all instance types that exist.
	go ec2.DescribeInstanceTypes(nil, "us-east-1", sendToDb, okFromDb, retval)
	go cache.InsertInstanceTypes(prov.DbHandle, sendToDb, okFromDb)
	select {
	case failure := <-retval:
		if failure != nil {
			log.Fatalf("Could not get instance type descriptions: %v\n", failure)
		}
	}
	close(retval)

	regions, err := cache.AllRegions(prov.DbHandle, true)
	if err != nil {
		log.Fatalf("failed to get regions: %+v\n", err)
	}
	for _, r := range regions {
		itypes, err := ec2.ListInstanceTypesByLocation(r)
		if err != nil {
			log.Fatalf("Failed to list instance types for %s: %+v\n", r, err)
		}
		if err = cache.InsertInstanceTypesForRegion(prov.DbHandle, itypes, r); err != nil {
			log.Fatalf("Failed to insert instance types for %s: %+v\n", r, err)
		}
	}

	fmt.Println("Populated database.")
	return 0
}
