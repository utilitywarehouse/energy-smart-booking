package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-services/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/generator"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/accounts"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/grpchealth"
	"github.com/utilitywarehouse/uwos-go/v1/iam/machine"
	"golang.org/x/sync/errgroup"
)

func runCron(c *cli.Context) error {
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

	evaluationStore := store.NewSmartBookingEvaluation(pool)
	accountLinkStore := store.NewLink(pool)

	accountsGRPCConn, err := grpc.CreateConnection(ctx, c.String(accountsAPIHost))
	if err != nil {
		return err
	}
	opsServer.Add("accounts-api", grpchealth.NewCheckWithConnection(ctx, accountsGRPCConn, "", "", "unable to query accounts lookup api"))
	defer accountsGRPCConn.Close()

	accountsClient := accountService.NewNumberLookupServiceClient(accountsGRPCConn)

	mn, err := machine.New()
	if err != nil {
		return err
	}
	defer mn.Close()

	accountsRepo := accounts.NewAccountLookup(mn, accountsClient)

	clickGRPCConn, err := grpc.CreateConnection(ctx, c.String(clickAPIHost))
	if err != nil {
		return err
	}
	opsServer.Add("click-api", grpchealth.NewCheckWithConnection(ctx, clickGRPCConn, "", "", "unable to query click api"))
	defer clickGRPCConn.Close()

	clickClient := click.NewIssuerServiceClient(clickGRPCConn)

	clickConfig := generator.LinkProviderConfig{
		ExpirationTimeSeconds: c.Int(clickLinkExpirySeconds),
		ClickKeyID:            c.String(clickSigningKeyID),
		AuthScope:             c.String(clickScope),
		Location:              c.String(clickWebLocation),
		MobileLocation:        c.String(clickMobileLocation),
		Subject:               c.String(subject),
		Intent:                c.String(intent),
		Channel:               c.String(channel),
	}
	linkProvider, err := generator.NewLinkProvider(clickClient, &clickConfig)
	if err != nil {
		return fmt.Errorf("failed to create link provider: %w", err)
	}

	cron := generator.NewLink(mn, linkProvider, accountsRepo, evaluationStore, accountLinkStore)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer cancel()
		return cron.Run(ctx)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer log.Debug("signal handler finished")
		select {
		case <-ctx.Done():
			return nil
		case <-sigChan:
			cancel()
		}
		return nil
	})

	return g.Wait()
}
