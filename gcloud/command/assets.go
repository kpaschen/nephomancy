package command

import (
	"fmt"
	"log"
	"strings"
	"nephomancy/gcloud/assets"
)

type AssetsCommand struct {
	Command
}

func (*AssetsCommand) Help() string {
	helpText := `
	Usage: nephomancy gcloud assets [options]

	Print a summary of assets in use by a project.

	Options:
	  --project=PROJECT  ID of a gcloud project. The user you are authenticating as must have
	                     access to this project. The billing, compute, asset, and monitoring APIs
			     must be enabled for this project.
`
	return strings.TrimSpace(helpText)
}

func (*AssetsCommand) Synopsis() string {
	return "Prints assets list for a gcloud project."
}

// Run this with
// nephomancy gcloud assets --project=binderhub-test-275512
func (c *AssetsCommand) Run(args []string) int {
	fs := c.Command.defaultFlagSet("gcloudAssets")
	fs.Parse(args)

	project := c.Command.Project
	if project == "" {
		log.Fatalf("Need a project ID. You can see the IDs on the Gcloud console.\n")
	}

	projectPath := fmt.Sprintf("projects/%s", project)

	ax, err := assets.ListAssetsForProject(projectPath)
	if err != nil {
		log.Fatalf("Listing assets failed: %v", err)
	}
	for _, a := range ax {
		if a.AssetType == "" {
			continue
		}
		fmt.Printf("Asset %+v\n", a)
	}

	err = assets.ListInstances(project)
	if err != nil {
		log.Fatalf("listing instances via compute api failed: %v\n", err)
	}
	err = assets.ListDisks(project)
	if err != nil {
		log.Fatalf("listing disks via compute api failed: %v\n", err)
	}

	return 0
}
