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

	// Kafka
	energyPlatformKafkaBrokers = "energy-platform-kafka-brokers"
	energyPlatformKafkaVersion = "energy-platform-kafka-version"

	// Kafka topics
	eligibilityTopic     = "eligibility-events-topic"
	campaignabilityTopic = "campaignability-events-topic"
	suppliabilityTopic   = "suppliability-events-topic"

	altHanTopic       = "alt-han-events-topic"
	optOutTopic       = "opt-out-events-topic"
	accountPsrTopic   = "account-psr-events-topic"
	bookingRefTopic   = "booking-reference-events-topic"
	meterTopic        = "meter-events-topic"
	meterpointTopic   = "meterpoint-events-topic"
	occupancyTopic    = "occupancy-events-topic"
	serviceStateTopic = "service-state-events-topic"
	siteTopic         = "site-events-topic"
	wanCoverageTopic  = "wan-coverage-events-topic"

	batchSize    = "batch-size"
	postgresDSN  = "postgres-dsn"
	stateRebuild = "state-rebuild"

	// gRPC
	grpcPort     = "grpc-port"
	grpcLogLevel = "grpc-log-level"
)

var gitHash string // populated at compile time

func main() {
	app := &cli.App{
		Name:  appName,
		Usage: appDesc,
		Commands: []*cli.Command{
			{
				Name:   "grpc-api",
				Usage:  "run a gRPC API for clients to query eligibility for smart booking",
				Before: app.Before,
				Action: runGRPCApi,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    grpcPort,
						Usage:   "The port to listen on for API GRPC connections",
						EnvVars: []string{"GRPC_PORT"},
						Value:   8090,
					},
					&cli.StringFlag{
						Name:    grpcLogLevel,
						Usage:   "gRPC log level [debug|info|warn|error]",
						EnvVars: []string{"GRPC_LOG_LEVEL"},
						Value:   "error",
					},
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
				},
			},
			{
				Name: "evaluator",
				Flags: app.DefaultFlags().WithKafkaRequired().WithCustom(
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     energyPlatformKafkaVersion,
						EnvVars:  []string{"ENERGY_PLATFORM_KAFKA_VERSION"},
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     energyPlatformKafkaBrokers,
						EnvVars:  []string{"ENERGY_PLATFORM_KAFKA_BROKERS"},
						Required: true,
					},
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
					&cli.StringFlag{
						Name:     altHanTopic,
						EnvVars:  []string{"ALT_HAN_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     optOutTopic,
						EnvVars:  []string{"OPT_OUT_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     accountPsrTopic,
						EnvVars:  []string{"ACCOUNT_PSR_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bookingRefTopic,
						EnvVars:  []string{"BOOKING_REF_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     meterTopic,
						EnvVars:  []string{"METER_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     meterpointTopic,
						EnvVars:  []string{"METERPOINT_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     occupancyTopic,
						EnvVars:  []string{"OCCUPANCY_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     serviceStateTopic,
						EnvVars:  []string{"SERVICE_STATE_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     siteTopic,
						EnvVars:  []string{"SITE_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     wanCoverageTopic,
						EnvVars:  []string{"WAN_COVERAGE_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    batchSize,
						EnvVars: []string{"BATCH_SIZE"},
						Value:   1,
					},
					&cli.BoolFlag{
						Name:     stateRebuild,
						EnvVars:  []string{"STATE_REBUILD"},
						Required: true,
					},
				),
				Before: app.Before,
				Action: runEvaluator,
			},
			{
				Name: "projector",
				Flags: app.DefaultFlags().WithKafkaRequired().WithCustom(
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
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
