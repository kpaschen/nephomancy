package command

import (
	"fmt"
	"log"
	"strings"
	"nephomancy/gcloud/assets"
	pricing "nephomancy/gcloud/pricing"
	"nephomancy/gcloud/cache"
)

type CostCommand struct {
	Command
}

func (*CostCommand) Help() string {
	helpText := `
	Usage: nephomancy gcloud cost [options]

	Print a summary of resources and their estimated costs per month.
	Resources include compute (VMs, licenses); network (IP addresses,
	load balancers, egress); and storage (disks, images, cloud storage).
	So far, resources such as BigTable, Spanner etc are not included.

	Costs are estimated based on SKU pricing info published by Google.

	I make no guarantee that the estimate will be correct; you can cross-check with
	your project's billing console as well as with the Google Pricing Calculator.

	Options:
	  --workingdir=path  Optional: directory under which the cached data can be found.
	                     The cost command needs this in order to locate the sku cache 
			     database. If not set, defaults to current directory. The sku database
			     will be looked for in $workingdir/.nephomancy/gcloud/data/sku-dache.db.
			     If you do not have a cache db yet, create one using nephomancy gcloud init.

	  --project=PROJECT  ID of a gcloud project. The user you are authenticating as must have
	                     access to this project. The billing, compute, asset, and monitoring APIs
			     must be enabled for this project. Costs reported will be for assets associated
			     with this project as well as for assets whose costs are only available at
			     the organization or billing account level.
`
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

	project := c.Command.Project
	if project == "" {
		log.Fatalf("Need a project ID. You can see the IDs on the Gcloud console.\n")
	}

	projectPath := fmt.Sprintf("projects/%s", project)

	assets, err := assets.ListAssetsForProject(projectPath)
	if err != nil {
		log.Fatalf("Listing assets failed: %v", err)
	}
	for _, a := range assets {
		if a.AssetType == "" {
			continue
		}
		prices, err := cache.GetPricingInfo(db, &a)
		if err != nil {
			log.Printf("Could not get prices for asset %+v: %v\n", a, err)
			continue
		}
		if len(prices) == 0 {
			log.Printf("Zero prices for asset %s\n", a.Name)
			continue
		}
		err = pricing.CostRange(&a, prices)
		if err != nil {
			log.Fatalf("Failed to get price range: %v\n", err)
		}
		// fmt.Printf("Prices for asset %+v: %+v\n", a, prices)
	}
	return 0
}
