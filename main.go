package main

import (
	"github.com/mitchellh/cli"
	"log"
	"nephomancy/common/command"
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
		"gcloud cost": func() (cli.Command, error) {
			return &gcmds.CostCommand{}, nil
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
