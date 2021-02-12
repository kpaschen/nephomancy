package command

import (
	"fmt"
	"log"
	"nephomancy/common/resources"
	"strings"
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

	Not yet implemented:
	Call it with a template file containing a spec and say which providers
	you want to get provider details filled in.

        Options:
          --workingdir=path  %s
          --projectin=filename %s
          --projectout=filename %s
`, workingDirDoc, projectInDoc, projectOutDoc)
	return strings.TrimSpace(helpText)
}

func (*ResourcesCommand) Synopsis() string {
	return "Creates or completes a project resource file."
}

func (r *ResourcesCommand) Run(args []string) int {
	fs := r.Command.defaultFlagSet("resources")
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
		project = resources.MakeSampleProject()
	}
	if err = r.saveProject(&project); err != nil {
		log.Fatalf("Failed to save project: %v\n", err)
	}
	return 0
}
