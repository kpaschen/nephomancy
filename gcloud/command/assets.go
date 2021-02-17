package command

import (
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
	"strings"
)

// TODO: rename to ResourcesCommand (after extracting common parts)
type AssetsCommand struct {
	Command
}

func (a *AssetsCommand) Help() string {
	helpText := fmt.Sprintf(`
	Usage: nephomancy gcloud assets [options]

	Print a summary of assets in use by a project.

	Options:
	  --project=PROJECT  %s
	  --workingdir=path  %s
	  --projectin=filename %s
	  --projectout=filename %s
`, projectDoc, workingDirDoc, projectInDoc, projectOutDoc)
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

	// A project specified via an infile will be used even if
	// a project id is also specified via the --project flag.
	project, err := c.loadProject()
	if err != nil {
		log.Fatalf("Failed to load project from file: %v\n", err)
	}

	projectName := c.Command.Project

	db, err := c.DbHandle()
	if err != nil {
		log.Fatalf("Could not open database: %v\n", err)
	}
	defer c.CloseDb()

	if project == nil {
		if projectName == "" {
			log.Fatalf("Need a project ID. You can see the IDs on the GCloud console.\n")
		}
		projectPath := fmt.Sprintf("projects/%s", projectName)
		ax, err := assets.ListAssetsForProject(projectPath)
		if err != nil {
			log.Fatalf("Listing assets failed: %v", err)
		}
		p, err := assets.BuildProject(ax)
		if err != nil {
			log.Fatalf("Building project failed: %v", err)
		}
		/*
			err = assets.GetProject(projectName)
			if err != nil {
				log.Fatalf("Failed to get project: %v", err)
			}
		*/
		project = p
	}
	if err = cache.FillInSpec(db, project); err != nil {
		log.Fatalf("resolving project failed: %v", err)
	}

	if err = c.saveProject(project); err != nil {
		log.Fatalf("Failed to save project: %v\n", err)
	}
	return 0
}
