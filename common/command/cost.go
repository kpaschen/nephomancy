package command

import (
	"fmt"
	"log"
	"nephomancy/common/registry"
	"nephomancy/common/resources"
	"nephomancy/common/utils"
	"strings"
	// The modules implementing providers have to be loaded
	_ "nephomancy/aws/provider"
	_ "nephomancy/dcs/provider"
	_ "nephomancy/gcloud/provider"
)

type CostCommand struct {
	Command
}

func (r *CostCommand) Help() string {
	helpText := fmt.Sprintf(`
        Usage: nephomancy cost [options]

	Estimate costs for a project.

	This takes a project file with provider details in it and creates a
	CSV file containing estimated monthly costs for each provider and
	resource.

	Costs are estimated based on information published by cloud providers.
	You can view this information by looking in
	$workingdir/.nephomancy/data/<provider name>. That directory should contain
	a sqlite database file which you can inspect or even modify. Some cloud provider
	implementations (e.g. gcloud) let you create this file programmatically via
	the command "nephomancy <provider> init". 

	I provide no guarantee that the estimates, the data they are built on, or the
	way my code derives them, are correct. I recommend cross-checking
	against your cloud provider's cost calculator and/or billing console so you can
	get an idea whether my estimates are approximately correct for you.

        Options:
          --workingdir=path  %s
          --projectin=filename %s
          --costreport=filename %s
`, workingDirDoc, projectInDoc, costReportDoc)
	return strings.TrimSpace(helpText)
}

func (*CostCommand) Synopsis() string {
	return "Estimates costs for your cloud project."
}

func (r *CostCommand) Run(args []string) int {
	fs := r.Command.DefaultFlagSet("cost")
	fs.Parse(args)

	infile, err := r.ProjectInFile()
	if err != nil {
		log.Fatalf("Bad project infile: %v\n", err)
	}
	if infile == "" {
		log.Fatalf("Please specify a project via the --projectin parameter.\n")
	}
	project, err := r.loadProject()
	if err != nil {
		log.Fatalf("Failed to load project from file %s: %v\n",
			infile, err)
	}
	dd, err := r.DataDir()
	if err != nil {
		log.Fatalf("Failed to set up data directory: %v\n", err)
	}
	projectName := project.Name
	f, err := r.Command.getCostFile(projectName)
	if err != nil {
		log.Fatalf("Failed to create cost report file: %v\n", err)
	}
	reporter := utils.CostReporter{}
	reporter.Init(f)

	providers := resources.GetProviderNames(*project)
	if len(providers) == 0 {
		log.Fatalf("Project spec is missing provider details, please run 'nephomancy resources' first.\n")
	}

	for _, provName := range providers {
		prov, err := registry.GetProvider(provName)
		if err != nil {
			log.Fatalf("Failed to get provider %s: %v\n", provName, err)
		}
		err = prov.Initialize(dd)
		if err != nil {
			log.Fatalf("Failed to initialize provider %s: %v\n", provName, err)
		}
		// Maybe call a consistency checker here?
		costs, err := prov.GetCost(project)
		if err != nil {
			log.Fatalf("Failed to get costs for provider %s: %v\n",
				provName, err)
		}

		for _, c := range costs {
			if err = reporter.AddLine(c); err != nil {
				log.Fatalf("Failed to report cost line %+v: %v\n",
					c, err)
			}
		}
	}
	reporter.Flush()
	log.Printf("Wrote costs to %s\n", f.Name())

	return 0
}
