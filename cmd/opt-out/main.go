package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
)

const (
	appName = "energy-smart-booking-opt-out"
	appDesc = "handles energy smart booking account opt outs"

	httpServerPort = "http-server-port"

	// Kafka
	KafkaBrokers      = "kafka-brokers"
	KafkaVersion      = "kafka-version"
	optOutEventsTopic = "opt-out-events-topic"
	batchSize         = "batch-size"

	postgresDSN = "postgres-dsn"

	accountsAPIHost = "accounts-api-host"
)

var gitHash string // populated at compile time

func main() {
	app := &cli.App{
		Name:   appName,
		Usage:  appDesc,
		Before: app.Before,
		Commands: []*cli.Command{
			{
				Name: "api",
				Flags: app.FlagBuilder{}.WithCustom(
					&cli.StringSliceFlag{
						Name:     KafkaBrokers,
						EnvVars:  []string{"KAFKA_BROKERS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     KafkaVersion,
						EnvVars:  []string{"KAFKA_VERSION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    optOutEventsTopic,
						EnvVars: []string{"OPT_OUT_EVENTS_TOPIC"},
					},
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    httpServerPort,
						EnvVars: []string{"HTTP_SERVER_PORT"},
						Value:   8090,
					},
					&cli.StringFlag{
						Name:    accountsAPIHost,
						EnvVars: []string{"ACCOUNTS_API_HOST"},
					},
				),
				Action: runServer,
			},
			{
				Name: "projector",
				Flags: app.FlagBuilder{}.WithCustom(
					&cli.StringSliceFlag{
						Name:     KafkaBrokers,
						EnvVars:  []string{"KAFKA_BROKERS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     KafkaVersion,
						EnvVars:  []string{"KAFKA_VERSION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    optOutEventsTopic,
						EnvVars: []string{"OPT_OUT_EVENTS_TOPIC"},
					},
					&cli.IntFlag{
						Name:    batchSize,
						EnvVars: []string{"BATCH_SIZE"},
						Value:   8090,
					},
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
				),
				Action: runProjector,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Panic("unable to run app")
	}
}
