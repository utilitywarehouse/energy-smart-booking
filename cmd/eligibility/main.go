package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
)

const (
	appName = "energy-smart-booking-eligibility"
	appDesc = "evaluates and produces eligibility, suppliability and campaignability events for smart booking"

	// Kafka topics
	eligibilityTopic     = "eligibility-events-topic"
	campaignabilityTopic = "campaignability-events-topic"
	suppliabilityTopic   = "suppliability-events-topic"

	batchSize   = "batch-size"
	postgresDSN = "postgres-dsn"
)

var gitHash string // populated at compile time

func main() {
	app := &cli.App{
		Name:  appName,
		Usage: appDesc,
		Commands: []*cli.Command{
			{
				Name: "projector",
				Flags: app.DefaultFlags().WithKafkaRequired().WithCustom(
					&cli.StringFlag{
						Name:     eligibilityTopic,
						EnvVars:  []string{"ELIGIBILITY_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     suppliabilityTopic,
						EnvVars:  []string{"SUPPLIABILITY_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     campaignabilityTopic,
						EnvVars:  []string{"CAMPAIGNABILITY_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    batchSize,
						EnvVars: []string{"BATCH_SIZE"},
						Value:   1,
					},
				),
				Before: app.Before,
				Action: runProjector,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Panic("unable to run app")
	}
}
