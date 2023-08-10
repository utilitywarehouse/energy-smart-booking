package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	bq "github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/big_query"

	"cloud.google.com/go/bigquery"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage/v2"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

var (
	commandNameBigQueryIndexer = "bq-indexer"
	commandUsageBigQueryUsage  = "a consumer that reads from bookings topic and indexes events to big query"

	bigQueryProjectID               = "big-query-project-id"
	bigQueryCredentialsFile         = "big-query-credentials-file" //nolint: gosec
	bigQueryDatasetID               = "big-query-dataset-id"
	bigQueryRescheduledBookingTable = "big-query-rescheduled-booking-table"
	bigQueryBookingTable            = "big-query-booking-table"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   commandNameBigQueryIndexer,
		Usage:  commandUsageBigQueryUsage,
		Action: runBigQueryIndexerAction,
		Flags: app.DefaultFlags().WithCustom(
			&cli.StringFlag{
				Name:     bigQueryProjectID,
				EnvVars:  []string{"BIG_QUERY_PROJECT_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     bigQueryCredentialsFile,
				EnvVars:  []string{"BIG_QUERY_CREDENTIALS_FILE"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     bigQueryDatasetID,
				EnvVars:  []string{"BIG_QUERY_DATASET_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     bigQueryRescheduledBookingTable,
				EnvVars:  []string{"BIG_QUERY_RESCHEDULED_BOOKING_TABLE"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     bigQueryBookingTable,
				EnvVars:  []string{"BIG_QUERY_BOOKING_TABLE"},
				Required: true,
			},
		),
	})
}

func runBigQueryIndexerAction(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName+commandNameBigQueryIndexer, appDesc)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	bookingSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(flagBookingTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to booking events kafka source: %w", err)
	}
	defer bookingSource.Close()
	opsServer.Add("booking-events-source", substratehealth.NewCheck(bookingSource, "unable to consume booking events"))

	bqClient, err := bigquery.NewClient(ctx, c.String(bigQueryProjectID), option.WithCredentialsFile(c.String(bigQueryCredentialsFile)))
	if err != nil {
		return fmt.Errorf("unable to create bigquery client: %w", err)
	}

	bookingIndexer := bq.NewRescheduledBookingIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQueryBookingTable), c.String(bigQueryRescheduledBookingTable))

	g.Go(func() error {
		defer log.Info("booking consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(flagBatchSize), time.Second, bookingSource, bookingIndexer)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer log.Info("signal handler finished")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigChan:
			cancel()
		}
		return nil
	})

	return g.Wait()
}
