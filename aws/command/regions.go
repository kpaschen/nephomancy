package command

import (
	"log"
	"nephomancy/aws/resources"
	"nephomancy/common/command"
)

type RegionsCommand struct {
	command.Command
}

func (*RegionsCommand) Help() string {
	return "list regions"
}

func (*RegionsCommand) Synopsis() string {
	return "lists regions"
}

func (r *RegionsCommand) Run(args []string) int {
	fs := r.Command.DefaultFlagSet("awsRegions")
	fs.Parse(args)

	regions, err := resources.ListRegions()
	// regions, err := resources.ListServices()
	if err != nil {
		log.Fatalf("Failed to list aws regions: %v\n", err)
	}
	_ = regions
	return 0
}
