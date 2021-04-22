package main

import (
	"github.com/mitchellh/cli"
	"log"
	awscmds "nephomancy/aws/command"
	"nephomancy/common/command"
	dcscmds "nephomancy/dcs/command"
	gcmds "nephomancy/gcloud/command"
	"os"
)

func main() {
	c := cli.NewCLI("nephomancy", "0.0.1") // TODO: get version from a file
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"resources": func() (cli.Command, error) {
			return &command.ResourcesCommand{}, nil
		},
		"cost": func() (cli.Command, error) {
			return &command.CostCommand{}, nil
		},
		"aws init": func() (cli.Command, error) {
			return &awscmds.InitCommand{}, nil
		},
		"aws regions": func() (cli.Command, error) {
			return &awscmds.RegionsCommand{}, nil
		},
		"aws list": func() (cli.Command, error) {
			return &awscmds.ListCommand{}, nil
		},
		"dcs init": func() (cli.Command, error) {
			return &dcscmds.InitCommand{}, nil
		},
		"gcloud init": func() (cli.Command, error) {
			return &gcmds.InitCommand{}, nil
		},
		"gcloud assets": func() (cli.Command, error) {
			return &gcmds.AssetsCommand{}, nil
		},
		"gcloud services": func() (cli.Command, error) {
			return &gcmds.ServicesCommand{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}
