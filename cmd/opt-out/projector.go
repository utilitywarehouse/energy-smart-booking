package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage/v2"
	"github.com/utilitywarehouse/energy-services/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/accounts"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/grpchealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/utilitywarehouse/uwos-go/v1/iam/machine"
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
	defer pool.Close()

	db := store.NewAccountOptOut(pool)

	mn, err := machine.New()
	if err != nil {
		return err
	}
	defer mn.Close()

	grpcConn, err := grpc.CreateConnection(ctx, c.String(accountsAPIHost))
	if err != nil {
		return err
	}
	opsServer.Add("accounts-api", grpchealth.NewCheckWithConnection(ctx, grpcConn, "", "", "unable to query accounts lookup api"))
	defer grpcConn.Close()

	accountsClient := accountService.NewNumberLookupServiceClient(grpcConn)
	accountsRepo := accounts.NewAccountLookup(mn, accountsClient)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	optOutEventsSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(optOutEventsTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to opt out events kafka source: %w", err)
	}
	defer optOutEventsSource.Close()
	opsServer.Add("opt-out-events-source", substratehealth.NewCheck(optOutEventsSource, "unable to consume opt out events"))

	g.Go(func() error {
		defer logrus.Info("opt out events consumer finished")
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, optOutEventsSource, consumer.Handle(db, accountsRepo))
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
