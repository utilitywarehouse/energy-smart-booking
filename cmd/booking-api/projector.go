package main

import (
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
)

var (
	commandNameProjector  = "projector"
	commandUsageProjector = "a projector service that consumes booking events to internally project booking state"

	flagBatchSize = "batch-size"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   commandNameProjector,
		Usage:  commandUsageProjector,
		Action: projectorAction,
		Flags: app.DefaultFlags().WithKafkaRequired().WithCustom(
			&cli.IntFlag{
				Name:    flagBatchSize,
				EnvVars: []string{"BATCH_SIZE"},
			},
		),
	})
}

func projectorAction(c *cli.Context) error {
	opsServer := makeOps(c)
	return opsServer.Start(c.Context)
}
