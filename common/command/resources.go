package command

import (
	"fmt"
	"log"
	"nephomancy/common/registry"
	"nephomancy/common/resources"
	"strings"
	// The modules implementing providers have to be loaded
	_ "nephomancy/aws/provider"
	_ "nephomancy/gcloud/provider"
)

type ResourcesCommand struct {
	Command
}

func (r *ResourcesCommand) Help() string {
	helpText := fmt.Sprintf(`
        Usage: nephomancy resources [options]

	Create or complete a project resource file.

	Call this with no projectin parameter set, and you will get
	a generic template file.

	Call it with a template file containing a spec and say which provider
	you want to get provider details filled in.

        Options:
          --workingdir=path  %s
          --projectin=filename %s
          --projectout=filename %s
	  --provider=name %s
`, workingDirDoc, projectInDoc, projectOutDoc, providerDoc)
	return strings.TrimSpace(helpText)
}

func (*ResourcesCommand) Synopsis() string {
	return "Creates or completes a project resource file."
}

func (r *ResourcesCommand) Run(args []string) int {
	fs := r.Command.DefaultFlagSet("resources")
	var location string
	fs.StringVar(&location, "location", "", "Location. Can be a global region (EMEA, APAC, NAM, LATAM), a continent (Africa, Asia, Europe, NorthAmerica, SouthAmerica) or a three-letter ISO country code.")
	fs.Parse(args)

	infile, err := r.ProjectInFile()
	if err != nil {
		log.Fatalf("Bad project infile: %v\n", err)
	}
	var project resources.Project
	if infile != "" {
		p, err := r.loadProject()
		if err != nil {
			log.Fatalf("Failed to load project from file %s: %v\n",
				infile, err)
		}
		project = *p
	} else {
		project = resources.MakeSampleProject(location)
	}

	if r.provider != "" {
		prov, err := registry.GetProvider(r.provider)
		if err != nil {
			log.Fatalf("Failed to get provider %s: %v\n", r.provider, err)
		}
		dd, err := r.DataDir()
		if err != nil {
			log.Fatalf("Failed to set up data directory: %v\n", err)
		}
		err = prov.Initialize(dd)
		if err != nil {
			log.Fatalf("Failed to initialize provider %s: %v\n", r.provider, err)
		}

		err = prov.FillInProviderDetails(&project)
		if err != nil {
			log.Fatalf("Failed to fill in details for provider %s: %v\n",
				r.provider, err)
		}
	}

	if err = r.saveProject(&project); err != nil {
		log.Fatalf("Failed to save project: %v\n", err)
	}
	return 0
}
