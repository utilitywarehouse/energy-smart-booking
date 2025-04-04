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
	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-pkg/substratemessage/v2"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/bq"
	"github.com/utilitywarehouse/energy-smart-booking/internal/indexer"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/accounts"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/utilitywarehouse/uwos-go/iam/machine"
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

	optOutEventsSource, err := app.GetKafkaSource(c, c.String(app.KafkaConsumerGroup), c.String(optOutEventsTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to opt out events kafka source: %w", err)
	}
	defer optOutEventsSource.Close()
	opsServer.Add("opt-out-events-source", substratehealth.NewCheck(optOutEventsSource, "unable to consume opt out events"))

	grpcConn, err := grpc.CreateConnection(c.String(accountsAPIHost))
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	mn, err := machine.New()
	if err != nil {
		return err
	}
	defer mn.Close()

	accountsClient := accountService.NewNumberLookupServiceClient(grpcConn)
	accountsRepo := accounts.NewAccountLookup(mn, accountsClient)

	bqClient, err := bigquery.NewClient(ctx, c.String(bigQueryProjectID), option.WithCredentialsFile(c.String(bigQueryCredentialsFile)))
	if err != nil {
		log.WithError(err).Panic("unable to create bigquery client")
	}

	dataset := bqClient.Dataset(c.String(bigQueryDatasetID))
	optOutAdded := indexer.New(ctx, dataset.Table(c.String(bigQueryOptOutAddedTable)), &bq.OptOutAdded{}, c.Int(batchSize))
	optOutRemove := indexer.New(ctx, dataset.Table(c.String(bigQueryOptOutRemovedTable)), bq.OptOutRemoved{}, c.Int(batchSize))

	indexer := bq.BigQueryIndexer{
		OptOutAdded:   optOutAdded,
		OptOutRemoved: optOutRemove,
		AccountsRepo:  accountsRepo,
	}

	g.Go(func() error {
		defer log.Info("opt out big query indexer finished")
		//nolint
		return substratemessage.BatchConsumer(ctx, c.Int(batchSize), time.Second, optOutEventsSource, &indexer)
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
