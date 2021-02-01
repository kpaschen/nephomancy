package command

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
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
	  --workingdir=path  Optional: directory with cached sku and machine/disk type data.
	                     The assets command needs this in order to resolve the machine
			     type names contained in the assets and obtain the actual number
			     of cpus and max amount of RAM. If not set, defaults to current directory. If you do not have a cache db yet, create one using nephomancy gcloud init.
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

	db, err := c.DbHandle()
	if err != nil {
		log.Fatalf("Could not open sku cache db: %v\n", err)
	}
	defer c.CloseDb()
	_ = db

	/*
	err = assets.GetProject(project)
	if err != nil {
		log.Fatalf("Failed to get project: %v", err)
	}
	*/

	ax, err := assets.ListAssetsForProject(projectPath)
	if err != nil {
		log.Fatalf("Listing assets failed: %v", err)
	}
	p, err := assets.BuildProject(ax)
	if err != nil {
		log.Fatalf("Building project failed: %v", err)
	}
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent: "  ",
	}
	str := options.Format(p)
	fmt.Printf("project: %s\n", str)

	return 0
}
