package command

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"nephomancy/gcloud/assets"
	pricing "nephomancy/gcloud/pricing"
	"nephomancy/gcloud/cache"
)

type CostCommand struct {
	// Probably want some credentials and config here.
}

func (*CostCommand) Help() string {
	return "Prints cost estimates for your gcloud setup."
}

func (*CostCommand) Synopsis() string {
	return "Prints cost estimates for your gcloud setup."
}

// Run this with
// nephomancy gcloud cost --project=binderhub-test-275512 --dbfile=gcloud/cmds/sku-cache.db
func (*CostCommand) Run(args []string) int {
	fs := flag.NewFlagSet("gcloudCost", flag.ExitOnError)
	var project, dbFile string
	fs.StringVar(&project, "project", "", "project name")
	fs.StringVar(&dbFile, "dbfile", "sku-cache.db", "sku cache db file")
	fs.Parse(args)

	db, _ := sql.Open("sqlite3", dbFile)
	defer db.Close()

	projectPath := fmt.Sprintf("projects/%s", project)

	assets, err := assets.ListAssetsForProject(projectPath)
	if err != nil {
		log.Fatalf("listing assets: %v", err)
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
			log.Printf("zero prices for asset %s\n", a.Name)
			continue
		}
		err = pricing.CostRange(&a, prices)
		if err != nil {
			log.Fatalf("Failed to get price range: %v\n", err)
		}
		fmt.Printf("prices for asset %+v: %+v\n", a, prices)
	}
	return 0
}
