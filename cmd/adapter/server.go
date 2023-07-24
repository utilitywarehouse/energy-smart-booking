package main

import (
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   "server",
		Usage:  "a listen server handling booking requests",
		Action: serverAction,
		Flags:  app.DefaultFlags().WithGrpc(),
	})
}

func serverAction(c *cli.Context) error {
	opsServer := makeOps(c)
	return opsServer.Start(c.Context)
}
