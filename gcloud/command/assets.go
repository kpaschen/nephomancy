package command

import (
	"fmt"
	"encoding/json"
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
		log.Fatalf("Need a project ID. You can see the IDs on the GCloud console.\n")
	}

	projectPath := fmt.Sprintf("projects/%s", project)

	ax, err := assets.ListAssetsForProject(projectPath)
	if err != nil {
		log.Fatalf("Listing assets failed: %v", err)
	}
	assetStructure, err := assets.BuildAssetStructure(ax)
	if err != nil {
		log.Fatalf("Structuring assets failed: %v", err)
	}
	structureAsJsonBytes, err := json.MarshalIndent(*assetStructure, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal json: %v", err)
	}
	fmt.Printf("structure: %s\n", string(structureAsJsonBytes))

	return 0
}
