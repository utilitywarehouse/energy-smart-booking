package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
)

const (
	appName = "energy-smart-booking-opt-out"
	appDesc = "handles energy smart booking account opt outs"

	httpServerPort = "http-server-port"

	// Kafka
	optOutEventsTopic = "opt-out-events-topic"
	batchSize         = "batch-size"

	postgresDSN = "postgres-dsn"

	accountsAPIHost = "accounts-api-host"

	// bigQuery
	bigQueryProjectID          = "big-query-project-id"
	bigQueryDatasetID          = "big-query-dataset-id"
	bigQueryCredentialsFile    = "big-query-credentials-file" //nolint:gosec
	bigQueryOptOutAddedTable   = "big-query-opt-out-added-table"
	bigQueryOptOutRemovedTable = "big-query-opt-out-removed-table"
)

var gitHash string // populated at compile time

func main() {
	app := &cli.App{
		Name:  appName,
		Usage: appDesc,
		Commands: []*cli.Command{
			{
				Name: "api",
				Flags: app.DefaultFlags().WithCustom(
					&cli.StringSliceFlag{
						Name:     app.KafkaBrokers,
						EnvVars:  []string{"KAFKA_BROKERS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaVersion,
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
				Before: app.Before,
				Action: runServer,
			},
			{
				Name: "projector",
				Flags: app.DefaultFlags().WithCustom(
					&cli.StringSliceFlag{
						Name:     app.KafkaBrokers,
						EnvVars:  []string{"KAFKA_BROKERS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaVersion,
						EnvVars:  []string{"KAFKA_VERSION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaConsumerGroup,
						EnvVars:  []string{"KAFKA_CONSUMER_GROUP"},
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
					&cli.StringFlag{
						Name:    accountsAPIHost,
						EnvVars: []string{"ACCOUNTS_API_HOST"},
					},
				),
				Action: runProjector,
			},
			{
				Name: "event-producer",
				Flags: app.DefaultFlags().WithCustom(
					&cli.StringFlag{
						Name:     postgresDSN,
						EnvVars:  []string{"POSTGRES_DSN"},
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     app.KafkaBrokers,
						EnvVars:  []string{"KAFKA_BROKERS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaVersion,
						EnvVars:  []string{"KAFKA_VERSION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaConsumerGroup,
						EnvVars:  []string{"KAFKA_CONSUMER_GROUP"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    optOutEventsTopic,
						EnvVars: []string{"OPT_OUT_EVENTS_TOPIC"},
					},
				),
				Action: runEventProducer,
			},
			{
				Name: "big-query-indexer",
				Flags: app.DefaultFlags().WithCustom(
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
						Name:     bigQueryOptOutAddedTable,
						EnvVars:  []string{"BIG_QUERY_OPT_OUT_ADDED_TABLE"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     bigQueryOptOutRemovedTable,
						EnvVars:  []string{"BIG_QUERY_OPT_OUT_REMOVED_TABLE"},
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     app.KafkaBrokers,
						EnvVars:  []string{"KAFKA_BROKERS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaVersion,
						EnvVars:  []string{"KAFKA_VERSION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     app.KafkaConsumerGroup,
						EnvVars:  []string{"KAFKA_CONSUMER_GROUP"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    optOutEventsTopic,
						EnvVars: []string{"OPT_OUT_EVENTS_TOPIC"},
					},
					&cli.IntFlag{
						Name:    batchSize,
						EnvVars: []string{"BATCH_SIZE"},
						Value:   1,
					},
					&cli.StringFlag{
						Name:    accountsAPIHost,
						EnvVars: []string{"ACCOUNTS_API_HOST"},
					},
				),
				Action: runBigQueryIndexer,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("unable to run app", "error", err)
		os.Exit(1)
	}
}
