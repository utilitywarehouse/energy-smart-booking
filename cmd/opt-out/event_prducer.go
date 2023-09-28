package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/uw-labs/substrate"
)

func runEventProducer(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	pool, err := store.Setup(ctx, c.String(postgresDSN))
	if err != nil {
		return err
	}
	defer pool.Close()
	db := store.NewAccountOptOut(pool)

	optOutSink, err := app.GetKafkaSink(c, c.String(optOutEventsTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to opt-out sink: %w", err)
	}

	syncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(optOutSink), appName)

	accounts, err := db.List(ctx)
	if err != nil {
		return err
	}
	for _, a := range accounts {
		err = syncPublisher.Sink(ctx, &smart.AccountBookingOptOutAddedEvent{
			AccountId: a.ID,
			AddedBy:   a.AddedBy,
		}, a.AddedAt)
		if err != nil {
			return err
		}
	}

	return nil
}
