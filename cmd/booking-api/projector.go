package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage/v2"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/uw-labs/substrate"
	"golang.org/x/sync/errgroup"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
)

var (
	commandNameProjector  = "projector"
	commandUsageProjector = "a projector service that consumes events to internally project state"

	flagBatchSize = "batch-size"

	flagPlatformKafkaBrokers       = "platform-kafka-brokers"
	flagPlatformKafkaVersion       = "platform-kafka-version"
	flagPlatformKafkaConsumerGroup = "platform-kafka-consumer-group"

	flagBookingRefTopic = "booking-reference-topic"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   commandNameProjector,
		Usage:  commandUsageProjector,
		Action: projectorAction,
		Flags: app.DefaultFlags().WithCustom(
			&cli.StringSliceFlag{
				Name:    flagPlatformKafkaBrokers,
				EnvVars: []string{"PLATFORM_KAFKA_BROKERS"},
			},
			&cli.StringFlag{
				Name:    flagPlatformKafkaVersion,
				EnvVars: []string{"PLATFORM_KAFKA_VERSION"},
			},
			&cli.StringFlag{
				Name:    flagPlatformKafkaConsumerGroup,
				EnvVars: []string{"PLATFORM_KAFKA_CONSUMER_GROUP"},
			},
			&cli.StringFlag{
				Name:    flagBookingRefTopic,
				EnvVars: []string{"BOOKING_REFERENCE_TOPIC"},
			},
			&cli.IntFlag{
				Name:    flagBatchSize,
				EnvVars: []string{"BATCH_SIZE"},
			},
		),
	})
}

type SourceMap map[string]substrate.AsyncMessageSource

type kafkaConfig struct {
	FlagBrokers       string
	FlagVersion       string
	FlagConsumerGroup string
	FlagsTopic        []string
}

func makeSources(c *cli.Context, opsrv *ops.Server, configs []*kafkaConfig) (SourceMap, error) {
	sources := make(map[string]substrate.AsyncMessageSource)
	for _, config := range configs {
		for _, flagTopic := range config.FlagsTopic {
			source, err := app.GetKafkaSourceWithBroker(
				c.String(config.FlagConsumerGroup),
				c.String(flagTopic),
				c.String(config.FlagVersion),
				c.StringSlice(config.FlagBrokers))
			if err != nil {
				return nil, fmt.Errorf("could not initialise Kafka source for config [%s]: %w", flagTopic, err)
			}
			opsrv.Add(
				strings.Replace(flagTopic, "topic", "source", 1),
				substratehealth.NewCheck(
					source,
					fmt.Sprintf("unable to consume events from %s", flagTopic)))
			sources[flagTopic] = source
		}
	}
	return sources, nil
}

type consumerConfig struct {
	FlagTopic string
	BatchSize int
	Handler   substratemessage.Handler
}

func startConsumers(ctx context.Context, g *errgroup.Group, sources SourceMap, configs []*consumerConfig) error {
	for _, config := range configs {
		batchSize := config.BatchSize
		topic := config.FlagTopic
		source, ok := sources[topic]
		if !ok {
			return fmt.Errorf("source %s not found", topic)
		}
		handler := config.Handler

		g.Go(func() error {
			defer log.Infof("%s consumer finished", topic)
			return substratemessage.BatchConsumer(ctx, batchSize, time.Second, source, handler)
		})
	}
	return nil
}

func projectorAction(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := makeOps(c)
	sources, err := makeSources(
		c,
		opsServer,
		[]*kafkaConfig{
			{
				FlagBrokers:       app.KafkaBrokers,
				FlagVersion:       app.KafkaVersion,
				FlagConsumerGroup: app.KafkaConsumerGroup,
				FlagsTopic: []string{
					flagBookingRefTopic,
					flagBookingTopic,
				},
			},
			{
				FlagBrokers:       flagPlatformKafkaBrokers,
				FlagVersion:       flagPlatformKafkaVersion,
				FlagConsumerGroup: flagPlatformKafkaConsumerGroup,
				FlagsTopic: []string{
					app.ServiceStateTopic,
					app.OccupancyTopic,
					app.SiteTopic,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to initialise kafka sources: %w", err)
	}

	pool, err := store.Setup(ctx, c.String(flagPostgresDSN))
	if err != nil {
		return fmt.Errorf("failed to initialise database: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	batchSize := c.Int(flagBatchSize)
	err = startConsumers(ctx, g, sources, []*consumerConfig{
		{
			FlagTopic: app.SiteTopic,
			BatchSize: batchSize,
			Handler:   consumer.HandleSite(store.NewSite(pool)),
		},
		{
			FlagTopic: app.OccupancyTopic,
			BatchSize: batchSize,
			Handler:   consumer.HandleOccupancy(store.NewOccupancy(pool)),
		},
		{
			FlagTopic: app.ServiceStateTopic,
			BatchSize: batchSize,
			Handler:   consumer.HandleServiceState(store.NewService(pool)),
		},
		{
			FlagTopic: flagBookingRefTopic,
			BatchSize: batchSize,
			Handler:   consumer.HandleBookingReference(store.NewBookingReference(pool)),
		},
		{
			FlagTopic: flagBookingTopic,
			BatchSize: batchSize,
			Handler:   consumer.HandleBooking(store.NewBooking(pool), store.NewOccupancy(pool)),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialise consumers: %w", err)
	}

	g.Go(func() error {
		defer log.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer log.Info("signal handler finished")
		select {
		case <-ctx.Done():
			log.Debug("received on done channel")
			return ctx.Err()
		case <-sigChan:
			log.Debug("received sigterm")
			cancel()
		}
		return nil
	})

	return g.Wait()
}
