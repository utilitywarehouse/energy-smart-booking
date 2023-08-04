package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func runBigQueryIndexer(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	bookingSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(flagBookingTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to eligibility events kafka source: %w", err)
	}
	defer bookingSource.Close()
	opsServer.Add("booking-events-source", substratehealth.NewCheck(bookingSource, "unable to consume booking events"))

	bqClient, err := bigquery.NewClient(ctx, c.String(bigQueryProjectID), option.WithCredentialsFile(c.String(bigQueryCredentialsFile)))
	if err != nil {
		return fmt.Errorf("unable to create bigquery client: %w", err)
	}

	/*eligibilityIndexer := bq.NewEligibilityIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQueryEligibilityTable))
	suppliabilityIndexer := bq.NewSuppliabilityIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQuerySuppliabilityTable))
	campaignabilityIndexer := bq.NewCampaignabilityIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQueryCampaignabilityTable))*/

	g.Go(func() error {
		defer log.Info("eligibility consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(flagBatchSize), time.Second, bookingSource, eligibilityIndexer)
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
