package main

import (
	"log"
	"os"
        gcmds "nephomancy/gcloud/command"
	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("nephomancy", "0.0.1")  // TODO: get version from a file
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"gcloud cost": func() (cli.Command, error) {
			return &gcmds.CostCommand{
			}, nil
		},
		"gcloud init": func() (cli.Command, error) {
			return &gcmds.InitCommand{
			}, nil
		},
		"gcloud assets": func() (cli.Command, error) {
			return &gcmds.AssetsCommand{
			}, nil
		},
		"gcloud services": func() (cli.Command, error) {
			return &gcmds.ServicesCommand{
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitStatus)
}
