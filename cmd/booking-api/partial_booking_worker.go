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
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/workers"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/go-ops-health-checks/pkg/sqlhealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/uw-labs/substrate"
	"golang.org/x/sync/errgroup"
)

var (
	commandNamePartialBookingWorker  = "partial-booking-worker"
	commandUsagePartialBookingWorker = "the worker for pending partial bookings"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   commandNamePartialBookingWorker,
		Usage:  commandUsagePartialBookingWorker,
		Action: partialBookingWorkerAction,
		Flags: app.DefaultFlags().WithGrpc().WithCustom(
			&cli.StringFlag{
				Name:     flagPostgresDSN,
				EnvVars:  []string{"POSTGRES_DSN"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    flagPartialBookingCron,
				EnvVars: []string{"PARTIAL_BOOKING_CRON"},
				Value:   "* * * * *",
			},
			&cli.DurationFlag{
				Name:    flagRetainedBookingPeriodAlertThreshold,
				EnvVars: []string{"RETAINED_BOOKING_ALERT_THRESHOLD"},
				Value:   time.Hour * 2,
			},
		),
	})
}

func partialBookingWorkerAction(c *cli.Context) error {
	slog.Info("starting app", "git_hash", gitHash, "command", commandNameServer)

	opsServer := makeOps(c)

	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	pool, err := store.Setup(ctx, c.String(flagPostgresDSN))
	if err != nil {
		return err
	}
	opsServer.Add("pool", sqlhealth.NewCheck(stdlib.OpenDB(*pool.Config().ConnConfig), "unable to connect to the DB"))

	bookingSink, err := app.GetKafkaSinkWithBroker(c.String(flagBookingTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to booking [%s] kafka sink: %w", c.String(flagBookingTopic), err)
	}
	defer bookingSink.Close()
	opsServer.Add("booking-sink", substratehealth.NewCheck(bookingSink, "unable to sink booking events"))

	g, ctx := errgroup.WithContext(ctx)

	syncBookingPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(bookingSink), c.App.Name)

	// STORE //
	occupancyStore := store.NewOccupancy(pool)
	partialBookingStore := store.NewPartialBooking(pool)

	//WORKERS
	partialBookingWorker := workers.NewPartialBookingWorker(partialBookingStore, occupancyStore, syncBookingPublisher, c.Duration(flagRetainedBookingPeriodAlertThreshold))

	g.Go(func() error {
		defer slog.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer slog.Info("partial booking cron job finished")
		cron := cron.New()

		cron.Start()
		defer cron.Stop()

		if _, err := cron.AddFunc(c.String(flagPartialBookingCron), func() {
			if err := partialBookingWorker.Run(ctx); err != nil {
				slog.Error("failed to run partial booking cron", "error", err)
			}
		}); err != nil {
			return fmt.Errorf("cron job failed for partial booking cron, %w", err)
		}

		<-ctx.Done()
		return ctx.Err()
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
