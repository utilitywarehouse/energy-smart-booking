package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/grpc"
	grpcHelper "github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"github.com/utilitywarehouse/go-ops-health-checks/pkg/grpchealth"
	"github.com/utilitywarehouse/go-ops-health-checks/pkg/sqlhealth"
	"github.com/utilitywarehouse/uwos-go/v1/iam/machine"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	eligiblity_api "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
)

var (
	commandNameServer  = "server"
	commandUsageServer = "a listen server handling booking requests"

	accountsAPIHost    = "accounts-api-host"
	eligibilityAPIHost = "eligibility-api-host"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   commandNameServer,
		Usage:  commandUsageServer,
		Action: serverAction,
		Flags: app.DefaultFlags().WithGrpc().WithCustom(
			&cli.StringFlag{
				Name:     accountsAPIHost,
				EnvVars:  []string{"ACCOUNTS_API_HOST"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     eligibilityAPIHost,
				EnvVars:  []string{"ELIGIBILITY_API_HOST"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagPostgresDSN,
				EnvVars:  []string{"POSTGRES_DSN"},
				Required: true,
			},
		),
	})
}

func serverAction(c *cli.Context) error {
	log.WithField("git_hash", gitHash).WithField("command", commandNameServer).Info("starting app")

	opsServer := makeOps(c)

	mn, err := machine.New()
	if err != nil {
		return fmt.Errorf("unable to create new IAM machine, %w", err)
	}
	defer mn.Close()

	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	pool, err := store.GetPool(ctx, c.String(flagPostgresDSN))
	if err != nil {
		return err
	}
	opsServer.Add("pool", sqlhealth.NewCheck(stdlib.OpenDB(*pool.Config().ConnConfig), "unable to connect to the DB"))

	accountsConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(accountsAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to accounts-api host [%s]: %w", c.String(accountsAPIHost), err)
	}
	opsServer.Add("accounts-api", grpchealth.NewCheck(c.String(accountsAPIHost), "", "cannot query accounts"))
	defer accountsConn.Close()

	eligibilityConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(eligibilityAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to eligibility-api host [%s]: %w", c.String(eligibilityAPIHost), err)
	}
	opsServer.Add("eligibility-api", grpchealth.NewCheck(c.String(eligibilityAPIHost), "", "cannot connect eligibility to eligibility-api"))
	defer eligibilityConn.Close()

	g, ctx := errgroup.WithContext(ctx)

	grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(app.GrpcLogLevel))
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(app.GrpcPort)))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port [%d]: %w", c.Int(app.GrpcPort), err)
	}
	defer listen.Close()

	// GATEWAYS //
	accountGw := gateway.NewAccountGateway(mn, accountService.NewAccountServiceClient(accountsConn))
	eligibilityGw := gateway.NewEligibilityGateway(mn, eligiblity_api.NewEligiblityAPIClient(eligibilityConn))

	// STORE //
	occupancyStore := store.NewOccupancy(pool)
	siteStore := store.NewSite(pool)
	bookingStore := store.NewBooking(pool)

	// DOMAIN //
	customerDomain := domain.NewCustomerDomain(accountGw, eligibilityGw, occupancyStore, siteStore, bookingStore)

	bookingAPI := api.New(customerDomain)
	bookingv1.RegisterBookingAPIServer(grpcServer, bookingAPI)

	g.Go(func() error {
		defer log.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer log.Info("grpc server finished")
		return grpcServer.Serve(listen)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer log.Info("signal handler finished")
		select {
		case <-ctx.Done():
			grpcServer.GracefulStop()
			return ctx.Err()
		case <-sigChan:
			cancel()
		}
		return nil
	})

	return g.Wait()
}
