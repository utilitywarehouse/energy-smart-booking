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
	envelope "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/evaluation"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/uw-labs/substrate"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func runEvaluator(c *cli.Context) error {
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

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	accountStore := store.NewAccount(pool)
	bookingRefStore := store.NewBookingRef(pool)
	meterStore := store.NewMeter(pool)
	meterpointStore := store.NewMeterpoint(pool)
	occupancyStore := store.NewOccupancy(pool)
	serviceStore := store.NewService(pool)
	siteStore := store.NewSite(pool)
	postCodeStore := store.NewPostCode(pool)

	altHanSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(altHanTopic))
	if err != nil {
		return fmt.Errorf("unable to create alt han events source [%s]: %w", c.String(altHanTopic), err)
	}
	defer altHanSource.Close()
	opsServer.Add("alt-han-source", substratehealth.NewCheck(altHanSource, "unable to consume alt han events"))

	optOutSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(optOutTopic))
	if err != nil {
		return fmt.Errorf("unable to create opt out events source [%s]: %w", c.String(optOutTopic), err)
	}
	defer optOutSource.Close()
	opsServer.Add("opt-out-source", substratehealth.NewCheck(optOutSource, "unable to consume opt out events"))

	accountPSRSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(accountPsrTopic))
	if err != nil {
		return fmt.Errorf("unable to create account PSR events source [%s]: %w", c.String(accountPsrTopic), err)
	}
	defer accountPSRSource.Close()
	opsServer.Add("account-psr-source", substratehealth.NewCheck(accountPSRSource, "unable to consume account PSR events"))

	bookingRefSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(bookingRefTopic))
	if err != nil {
		return fmt.Errorf("unable to create booking ref events source [%s]: %w", c.String(bookingRefTopic), err)
	}
	defer bookingRefSource.Close()
	opsServer.Add("booking-reference-source", substratehealth.NewCheck(bookingRefSource, "unable to consume account booking reference events"))

	meterSource, err := app.GetKafkaSourceWithBroker(c.String(app.KafkaConsumerGroup), c.String(meterTopic), c.String(energyPlatformKafkaVersion), c.StringSlice(energyPlatformKafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to create meter events source [%s]: %w", c.String(meterTopic), err)
	}
	defer meterSource.Close()
	opsServer.Add("meter-source", substratehealth.NewCheck(meterSource, "unable to consume meter events"))

	meterpointSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(meterpointTopic))
	if err != nil {
		return fmt.Errorf("unable to create meterpoint events source [%s]: %w", c.String(meterpointTopic), err)
	}
	defer meterpointSource.Close()
	opsServer.Add("meterpoint-source", substratehealth.NewCheck(meterpointSource, "unable to consume meterpoint events"))

	occupancySource, err := app.GetKafkaSourceWithBroker(c.String(app.KafkaConsumerGroup), c.String(occupancyTopic), c.String(energyPlatformKafkaVersion), c.StringSlice(energyPlatformKafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to create occupancy events source [%s]: %w", c.String(occupancyTopic), err)
	}
	defer occupancySource.Close()
	opsServer.Add("occupancy-source", substratehealth.NewCheck(occupancySource, "unable to consume occupancy events"))

	serviceSource, err := app.GetKafkaSourceWithBroker(c.String(app.KafkaConsumerGroup), c.String(serviceStateTopic), c.String(energyPlatformKafkaVersion), c.StringSlice(energyPlatformKafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to create service events source [%s]: %w", c.String(serviceStateTopic), err)
	}
	defer serviceSource.Close()
	opsServer.Add("service-source", substratehealth.NewCheck(serviceSource, "unable to consume service events"))

	siteSource, err := app.GetKafkaSourceWithBroker(c.String(app.KafkaConsumerGroup), c.String(siteTopic), c.String(energyPlatformKafkaVersion), c.StringSlice(energyPlatformKafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to create site events source [%s]: %w", c.String(siteTopic), err)
	}
	defer siteSource.Close()
	opsServer.Add("site-source", substratehealth.NewCheck(siteSource, "unable to consume site events"))

	wanCoverageSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(wanCoverageTopic))
	if err != nil {
		return fmt.Errorf("unable to create wan coverage events source [%s]: %w", c.String(wanCoverageTopic), err)
	}
	defer wanCoverageSource.Close()
	opsServer.Add("wan-coverage-source", substratehealth.NewCheck(wanCoverageSource, "unable to consume wan coverage events"))

	eligibilitySink, err := app.GetKafkaSink(c, c.String(eligibilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to eligibility sink: %w", err)
	}
	eligibilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(eligibilitySink), appName)

	suppliabilitySink, err := app.GetKafkaSink(c, c.String(suppliabilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to suppliability sink: %w", err)
	}
	suppliabilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(suppliabilitySink), appName)

	campaignabilitySink, err := app.GetKafkaSink(c, c.String(campaignabilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to campaignability sink: %w", err)
	}
	campaignabilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(campaignabilitySink), appName)

	bookingEligibilitySink, err := app.GetKafkaSinkWithKeyFunc(c, c.String(bookingJourneyEligibilityTopic), keyFunc)
	if err != nil {
		return fmt.Errorf("unable to create booking journey eligibility sink: %w", err)
	}
	bookingEligibilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(bookingEligibilitySink), appName)

	evaluator := evaluation.NewEvaluator(
		occupancyStore,
		serviceStore,
		meterStore,
		eligibilitySyncPublisher,
		suppliabilitySyncPublisher,
		campaignabilitySyncPublisher,
		bookingEligibilitySyncPublisher,
	)

	g.Go(func() error {
		defer slog.Info("alt han events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, altHanSource, consumer.HandleAltHan(meterpointStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("opt out events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, optOutSource, consumer.HandleAccountOptOut(accountStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("account psr events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, accountPSRSource, consumer.HandleAccountPSR(accountStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("booking ref events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, bookingRefSource, consumer.HandleBookingRef(bookingRefStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("meter events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, meterSource, consumer.HandleMeter(meterStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("meterpoint events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, meterpointSource, consumer.HandleMeterpoint(meterpointStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("occupancy events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, occupancySource, consumer.HandleOccupancy(occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("service state events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, serviceSource, consumer.HandleService(serviceStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("site events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, siteSource, consumer.HandleSite(siteStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
	})
	g.Go(func() error {
		defer slog.Info("wan coverage events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, wanCoverageSource, consumer.HandleWanCoverage(postCodeStore, occupancyStore, evaluator, c.Bool(stateRebuild)))
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

func keyFunc(m substrate.Message) []byte {
	var env envelope.Envelope
	if err := proto.Unmarshal(m.Data(), &env); err != nil {
		return nil
	}

	inner, err := env.Message.UnmarshalNew()
	if err != nil {
		slog.Warn("error unmarshalling inner for key function", "error", err)
		return nil
	}
	e, ok := inner.(occupancyEligibility)
	if !ok {
		slog.Warn("message in event does not have an occupancy ID", "event_uuid", env.Uuid)
		return nil
	}

	return []byte(e.GetOccupancyId())
}

type occupancyEligibility interface {
	GetOccupancyId() string
}
