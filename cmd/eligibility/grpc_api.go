package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	smart_booking "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	pkgapp "github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	grpcHelper "github.com/utilitywarehouse/energy-services/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	"golang.org/x/sync/errgroup"
)

func runGRPCApi(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logrus.Info("running api")
	opsServer := ops.Default().
		WithPort(c.Int(pkgapp.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	g, ctx := errgroup.WithContext(ctx)

	pg, err := store.Setup(ctx, c.String(postgresDSN))
	if err != nil {
		return fmt.Errorf("couldn't initialise database: %w", err)
	}
	defer pg.Close()
	opsServer.Add("db", sqlhealth.NewCheck(stdlib.OpenDB(*pg.Config().ConnConfig), "unable to connect to the DB"))

	eligibilityStore := store.NewEligibility(pg)
	suppliabilityStore := store.NewSuppliability(pg)
	occupancyStore := store.NewOccupancy(pg)
	accountStore := store.NewAccount(pg)

	g.Go(func() error {
		logrus.Infof("Starts creating grpc server")
		grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(grpcLogLevel))
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(grpcPort)))
		if err != nil {
			return err
		}
		defer listen.Close()

		eligibilityAPI := api.NewEligibilityGRPCApi(eligibilityStore, suppliabilityStore, occupancyStore, accountStore)
		smart_booking.RegisterEligiblityAPIServer(grpcServer, eligibilityAPI)
		logrus.Infof("successfully registered grpc server")

		err = grpcServer.Serve(listen)
		if err != nil {
			logrus.Infof("error starting grpc server: %s", err.Error())
		}
		return err
	})

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigChan:
			logrus.Info("cancelling context")
			cancel()
		}
		return nil
	})

	return g.Wait()
}
