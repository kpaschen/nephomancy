package command

import (
	"fmt"
	"log"
	"nephomancy/common/utils"
	"nephomancy/gcloud/assets"
	"nephomancy/gcloud/cache"
	pricing "nephomancy/gcloud/pricing"
	"strings"
)

type CostCommand struct {
	Command
}

func (*CostCommand) Help() string {
	helpText := fmt.Sprintf(`
	Usage: nephomancy gcloud cost [options]

	Print a summary of assets and their estimated costs per month.
	Not all assets supported by Google Cloud are included, and thus the
	price estimate will be incomplete. For example, licenses as well as
	most backend service costs are not included in the estimate.

	Costs are estimated based on SKU pricing info published by Google.

	I provide no guarantee that the estimate will be correct; you can cross-check with
	your project's billing console as well as with the Google Pricing Calculator.

	Options:
	  --project=PROJECT  %s
	  --workingdir=path  %s
	  --projectin=filename %s
	  --costreport=filename %s
`, projectDoc, workingDirDoc, projectInDoc, costReportDoc)
	return strings.TrimSpace(helpText)
}

func (*CostCommand) Synopsis() string {
	return "Prints cost estimates for your gcloud setup."
}

// Run this with
// nephomancy gcloud cost --project=binderhub-test-275512
func (c *CostCommand) Run(args []string) int {
	fs := c.Command.defaultFlagSet("gcloudCost")
	fs.Parse(args)

	db, err := c.DbHandle()
	if err != nil {
		log.Fatalf("Could not open sku cache db: %v\nTry creating it with nephomancy gcloud init. If that does not help, try removing the .nephomancy/gcloud/data directory and then run init again.\n", err)
	}
	defer c.CloseDb()

	project, err := c.loadProject()
	if err != nil {
		log.Fatalf("Failed to load project from file: %v\n", err)
	}

	projectName := c.Command.Project

	if project == nil {
		if projectName == "" {
			log.Fatalf("Need a project ID. You can see the IDs on the Gcloud console.\n")
		}
		projectPath := fmt.Sprintf("projects/%s", project)
		ax, err := assets.ListAssetsForProject(projectPath)
		if err != nil {
			log.Fatalf("Listing assets failed: %v", err)
		}
		proj, err := assets.BuildProject(ax)
		if err != nil {
			log.Fatalf("Building project failed: %v", err)
		}
		project = proj
	}
	err = cache.ReconcileSpecAndAssets(db, project)
	if err != nil {
		log.Fatalf("Could not add resource types: %v", err)
	}
	costs, err := pricing.GetCost(db, project)
	if err != nil {
		log.Fatalf("Failed to get pricing information: %v", err)
	}
	if projectName == "" {
		projectName = project.Name
	}
	reporter := utils.CostReporter{}
	f, err := c.Command.getCostFile(projectName)
	if err != nil {
		log.Fatalf("Failed to create report file: %v", err)
	}
	reporter.Init(f)
	for _, c := range costs {
		if err = reporter.AddLine(c); err != nil {
			log.Fatalf("Failed to report cost line: %v", err)
		}
	}
	reporter.Flush()
	log.Printf("Wrote costs to %s\n", f.Name())
	return 0
}
