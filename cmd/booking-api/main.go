package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
)

const (
	appName = "energy-smart-booking-adapter"
	appDesc = "translates generic booking-related requests to specific booking providers"
)

var (
	gitHash string

	application = &cli.App{
		Name:   appName,
		Usage:  appDesc,
		Before: app.Before,
		Flags:  app.DefaultFlags(),
	}
)

func makeOps(c *cli.Context) *ops.Server {
	return ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)
}

func main() {
	if err := application.Run(os.Args); err != nil {
		log.WithError(err).Fatalln("service terminated unexpectedly")
		os.Exit(1)
	}
}
