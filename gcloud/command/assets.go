package command

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"nephomancy/gcloud/assets"
	common "nephomancy/common/resources"
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

	infile, err := c.ModelInFile()
	if err != nil {
		log.Fatalf("Could not use model in file: %v\n", err)
	}
	var project *common.Project
	var projectName string
	if infile != "" {
		pdata, err := ioutil.ReadFile(infile)
		if err != nil {
			log.Fatalf("Failed to read model from file %s: %v\n",
			infile, err)
		}
		p, err := assets.UnmarshalProject(pdata)
		if err != nil {
			log.Fatalf("Failed to decode project: %v\n", err)
		}
		if p != nil {
			project = p
		}
	}

	projectName = c.Command.Project
	if projectName == "" && project == nil {
		log.Fatalf("Need a project ID. You can see the IDs on the GCloud console.\n")
	}
	if projectName == "" {
		projectName = project.Name
	}

	projectPath := fmt.Sprintf("projects/%s", projectName)

	db, err := c.DbHandle()
	if err != nil {
		log.Fatalf("Could not open sku cache db: %v\n", err)
	}
	defer c.CloseDb()
	_ = db

	outfile, err := c.ModelOutFile()
	if err != nil {
		log.Fatalf("Could not use model out file: %v\n", err)
	}
	var f *os.File
	if outfile != "" {
		x, err := os.OpenFile(outfile, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
		if err != nil {
			log.Fatalf("Failed to create file %s: %v\n", outfile, err)
		}
		f = x
	}

	/*
	err = assets.GetProject(project)
	if err != nil {
		log.Fatalf("Failed to get project: %v", err)
	}
	*/

	if project == nil {
		fmt.Printf("project path: %s\n", projectPath)
		ax, err := assets.ListAssetsForProject(projectPath)
		if err != nil {
			log.Fatalf("Listing assets failed: %v", err)
		}
		p, err := assets.BuildProject(ax)
		if err != nil {
			log.Fatalf("Building project failed: %v", err)
		}
		project = p
	}
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent: "  ",
	}
	str := options.Format(project)
	if f != nil {
		f.WriteString(str)
		fmt.Printf("Asset model written to %s\n", outfile)
	} else {
		fmt.Printf("project: %s\n", str)
	}

	return 0
}
