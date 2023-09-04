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
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"github.com/utilitywarehouse/go-ops-health-checks/pkg/grpchealth"
	"github.com/utilitywarehouse/go-ops-health-checks/pkg/sqlhealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/utilitywarehouse/uwos-go/v1/iam/machine"
	"github.com/utilitywarehouse/uwos-go/v1/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry"
	"github.com/uw-labs/substrate"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	lowribeck_api "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
)

var (
	commandNameServer  = "server"
	commandUsageServer = "a listen server handling booking requests"

	accountsAPIHost  = "accounts-api-host"
	lowribeckAPIHost = "lowribeck-api-host"
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
				Name:     lowribeckAPIHost,
				EnvVars:  []string{"LOWRIBECK_API_HOST"},
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

	pdp, err := pdp.NewClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	pool, err := store.GetPool(ctx, c.String(flagPostgresDSN))
	if err != nil {
		return err
	}
	opsServer.Add("pool", sqlhealth.NewCheck(stdlib.OpenDB(*pool.Config().ConnConfig), "unable to connect to the DB"))

	auth := auth.New(pdp)

	accountsConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(accountsAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to accounts-api host [%s]: %w", c.String(accountsAPIHost), err)
	}
	opsServer.Add("accounts-api", grpchealth.NewCheck(c.String(accountsAPIHost), "", "cannot query accounts"))
	defer accountsConn.Close()

	lowribeckConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(lowribeckAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to lowribeck-api host [%s]: %w", c.String(lowribeckAPIHost), err)
	}
	opsServer.Add("lowribeck-api", grpchealth.NewCheck(c.String(lowribeckAPIHost), "", "cannot connect to lowribeck-api"))
	defer lowribeckConn.Close()

	bookingSink, err := app.GetKafkaSinkWithBroker(c.String(flagBookingTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to booking [%s] kafka sink: %w", c.String(flagBookingTopic), err)
	}
	defer bookingSink.Close()
	opsServer.Add("booking-sink", substratehealth.NewCheck(bookingSink, "unable to sink booking events"))

	g, ctx := errgroup.WithContext(ctx)

	closer, err := telemetry.Register(ctx,
		telemetry.WithServiceName(appName),
		telemetry.WithTeam("energy-smart"),
		telemetry.WithServiceVersion(gitHash),
	)
	if err != nil {
		log.Errorf("Telemetry cannot be registered: %v", err)
	}
	defer closer.Close()

	grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(app.GrpcLogLevel))
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(app.GrpcPort)))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port [%d]: %w", c.Int(app.GrpcPort), err)
	}
	defer listen.Close()

	// GATEWAYS //
	accountGw := gateway.NewAccountGateway(mn, accountService.NewAccountServiceClient(accountsConn))
	lowriBeckGateway := gateway.NewLowriBeckGateway(mn, lowribeck_api.NewLowriBeckAPIClient(lowribeckConn))

	// PUBLISHERS //

	syncBookingPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(bookingSink), c.App.Name)

	// STORE //
	occupancyStore := store.NewOccupancy(pool)
	siteStore := store.NewSite(pool)
	bookingStore := store.NewBooking(pool)

	// DOMAIN //
	bookingDomain := domain.NewBookingDomain(accountGw, lowriBeckGateway, occupancyStore, siteStore, bookingStore)

	bookingAPI := api.New(bookingDomain, syncBookingPublisher, auth)
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
