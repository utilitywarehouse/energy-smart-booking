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
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/bq"
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

	eligibilitySource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(eligibilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to eligibility events kafka source: %w", err)
	}
	defer eligibilitySource.Close()
	opsServer.Add("eligibility-events-source", substratehealth.NewCheck(eligibilitySource, "unable to consume eligibility events"))

	suppliabilitySource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(suppliabilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to suppliability events kafka source: %w", err)
	}
	defer suppliabilitySource.Close()
	opsServer.Add("suppliability-events-source", substratehealth.NewCheck(suppliabilitySource, "unable to consume suppliability events"))

	campaignabilitySource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(campaignabilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to campaignability events kafka source: %w", err)
	}
	defer campaignabilitySource.Close()
	opsServer.Add("campaignability-events-source", substratehealth.NewCheck(campaignabilitySource, "unable to consume campaignability events"))

	bqClient, err := bigquery.NewClient(ctx, c.String(bigQueryProjectID), option.WithCredentialsFile(c.String(bigQueryCredentialsFile)))
	if err != nil {
		return fmt.Errorf("unable to create bigquery client: %w", err)
	}

	eligibilityIndexer := bq.NewEligibilityIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQueryEligibilityTable))
	suppliabilityIndexer := bq.NewSuppliabilityIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQuerySuppliabilityTable))
	campaignabilityIndexer := bq.NewCampaignabilityIndexer(bqClient, c.String(bigQueryDatasetID), c.String(bigQueryCampaignabilityTable))

	g.Go(func() error {
		defer log.Info("eligibility consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, eligibilitySource, eligibilityIndexer)
	})
	g.Go(func() error {
		defer log.Info("suppliability consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, suppliabilitySource, suppliabilityIndexer)
	})
	g.Go(func() error {
		defer log.Info("campaignability consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, campaignabilitySource, campaignabilityIndexer)
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
