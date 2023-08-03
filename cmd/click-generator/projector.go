package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage/v2"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
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

	evaluationStore := store.NewSmartBookingEvaluation(pool)
	linkStore := store.NewLink(pool)

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

	g.Go(func() error {
		defer logrus.Info("eligibility events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, eligibilitySource, consumer.NewEligibility(evaluationStore, linkStore))
	})
	g.Go(func() error {
		defer logrus.Info("suppliability events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, suppliabilitySource, consumer.NewSuppliability(evaluationStore, linkStore))
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer logrus.Info("signal handler finished")
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
