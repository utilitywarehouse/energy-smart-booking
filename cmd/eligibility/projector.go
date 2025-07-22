package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"golang.org/x/sync/errgroup"
)

func runProjector(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	pool, err := store.Setup(ctx, c.String(postgresDSN))
	if err != nil {
		return err
	}
	opsServer.Add("db", sqlhealth.NewCheck(stdlib.OpenDB(*pool.Config().ConnConfig), "unable to connect to the DB"))
	defer pool.Close()

	eligibilityDB := store.NewEligibility(pool)
	suppliabilityDB := store.NewSuppliability(pool)
	campaignabilityDB := store.NewCampaignability(pool)

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

	g.Go(func() error {
		defer slog.Info("eligibility events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, eligibilitySource, consumer.HandleEligibility(eligibilityDB))
	})
	g.Go(func() error {
		defer slog.Info("suppliability events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, suppliabilitySource, consumer.HandleSuppliability(suppliabilityDB))
	})
	g.Go(func() error {
		defer slog.Info("campaignability events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, campaignabilitySource, consumer.HandleCampaignability(campaignabilityDB))
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer slog.Info("signal handler finished")
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
