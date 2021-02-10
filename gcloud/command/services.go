package command

import (
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	"strings"
)

type ServicesCommand struct {
	Command
}

func (*ServicesCommand) Help() string {
	helpText := `
	Usage: nephomancy gcloud services [options]

	Print a summary of services in use by a project.

	Options:
	  --project=PROJECT  ID of a gcloud project. The user you are authenticating as must have
	                     access to this project. The billing, compute, asset, and monitoring APIs
			     must be enabled for this project.`
	return strings.TrimSpace(helpText)
}

func (*ServicesCommand) Synopsis() string {
	return "Prints services list for a gcloud project."
}

// Run this with
// nephomancy gcloud services --project=binderhub-test-275512
func (c *ServicesCommand) Run(args []string) int {
	fs := c.Command.defaultFlagSet("gcloudServices")
	fs.Parse(args)

	project := c.Command.Project
	if project == "" {
		log.Fatalf("Need a project ID. You can see the IDs on the GCloud console.\n")
	}
	projectPath := fmt.Sprintf("projects/%s", project)

	regions, err := assets.ListRegions(project)
	if err != nil {
		log.Fatalf("Failed to get regions: %v", err)
	}
	fmt.Printf("regions: %v", regions)

	/*
	err = assets.ListServices(projectPath)
	if err != nil {
		log.Fatalf("Failed to get services: %v", err)
	}
	*/

	// err = assets.ListMetrics(projectPath, `metric.type=starts_with("compute.googleapis.com")`)
	err = assets.ListMetrics(projectPath)
	if err != nil {
		log.Fatalf("Failed to get metrics: %v", err)
	}

	return 0
}
