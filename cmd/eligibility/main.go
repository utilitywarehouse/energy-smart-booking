package main

import (
	"log/slog"
	"os"

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
	eligibilityTopic               = "eligibility-events-topic" //nolint:gosec
	campaignabilityTopic           = "campaignability-events-topic"
	suppliabilityTopic             = "suppliability-events-topic"
	bookingJourneyEligibilityTopic = "booking-journey-eligibility-topic"

	altHanTopic       = "alt-han-events-topic"
	optOutTopic       = "opt-out-events-topic"
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

	//eligibility
	msnExceptionFilePath = "msn-exception-file-path"

	// gRPC
	grpcPort    = "grpc-port"
	xoserveHost = "xoserve-host"
	ecoesHost   = "ecoes-host"

	// http
	httpPort = "http-port"

	// BigQuery
	bigQueryProjectID                         = "big-query-project-id"
	bigQueryDatasetID                         = "big-query-dataset-id"
	bigQueryCredentialsFile                   = "big-query-credentials-file" //nolint:gosec
	bigQueryCampaignabilityTable              = "big-query-campaignability-table"
	bigQuerySuppliabilityTable                = "big-query-suppliability-table"
	bigQueryEligibilityTable                  = "big-query-eligibility-table"
	bigQueryBookingJourneyEligibilityRefTable = "big-query-booking-journey-eligibility-ref-table"
)

var gitHash string // populated at compile time

func main() {
	app := &cli.App{
		Name:  appName,
		Usage: appDesc,
		Commands: []*cli.Command{
			{
				Name:  "grpc-api",
				Usage: "run a gRPC API for clients to query eligibility for smart booking",
				Flags: app.DefaultFlags().WithCustom(

					&cli.IntFlag{
						Name:    grpcPort,
						Usage:   "The port to listen on for API GRPC connections",
						EnvVars: []string{"GRPC_PORT"},
						Value:   8090,
					},
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    httpPort,
						Usage:   "The port to listen on for API http connections",
						EnvVars: []string{"HTTP_PORT"},
						Value:   8091,
					},
					&cli.StringFlag{
						Name:     xoserveHost,
						Usage:    "The xoserve host endpoint address",
						EnvVars:  []string{"XOSERVE_HOST"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     ecoesHost,
						Usage:    "The ecoes host endpoint address",
						EnvVars:  []string{"ECOES_HOST"},
						Required: true,
					},
				),
				Before: app.Before,
				Action: runGRPCApi,
			},
			{
				Name:  "http-api",
				Usage: "run a http API to trigger full eligibility evaluation",
				Flags: app.DefaultFlags().WithKafkaRequired().WithCustom(

					&cli.IntFlag{
						Name:    httpPort,
						Usage:   "The port to listen on for API http connections",
						EnvVars: []string{"HTTP_PORT"},
						Value:   8091,
					},
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
					&cli.StringFlag{
						Name:     bookingJourneyEligibilityTopic,
						EnvVars:  []string{"BOOKING_JOURNEY_ELIGIBILITY_EVENTS_TOPIC"},
						Required: true,
					},
				),
				Before: app.Before,
				Action: runHTTPApi,
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
					&cli.StringFlag{
						Name:     bookingJourneyEligibilityTopic,
						EnvVars:  []string{"BOOKING_JOURNEY_ELIGIBILITY_EVENTS_TOPIC"},
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
					&cli.StringFlag{
						Name:     msnExceptionFilePath,
						EnvVars:  []string{"MSN_EXCEPTION_FILE_PATH"},
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
			{
				Name: "bq-indexer",
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
					&cli.StringFlag{
						Name:     bookingJourneyEligibilityTopic,
						EnvVars:  []string{"BOOKING_JOURNEY_ELIGIBILITY_EVENTS_TOPIC"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryProjectID,
						EnvVars:  []string{"BIG_QUERY_PROJECT_ID"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryDatasetID,
						EnvVars:  []string{"BIG_QUERY_DATASET_ID"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryCredentialsFile,
						EnvVars:  []string{"BIG_QUERY_CREDENTIALS_FILE"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryCampaignabilityTable,
						EnvVars:  []string{"BIG_QUERY_CAMPAIGNABILITY_TABLE"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQuerySuppliabilityTable,
						EnvVars:  []string{"BIG_QUERY_SUPPLIABILITY_TABLE"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryEligibilityTable,
						EnvVars:  []string{"BIG_QUERY_ELIGIBILITY_TABLE"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryBookingJourneyEligibilityRefTable,
						EnvVars:  []string{"BIG_QUERY_BOOKING_JOURNEY_REF_TABLE"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    batchSize,
						EnvVars: []string{"BATCH_SIZE"},
						Value:   1,
					},
				),
				Before: app.Before,
				Action: runBigQueryIndexer,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("unable to run app", "error", err)
		os.Exit(1)
	}
}
