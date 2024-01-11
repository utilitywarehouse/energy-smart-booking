package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	eligibilityv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/cache"
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
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
	lowribeck_api "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
)

var (
	commandNameServer  = "server"
	commandUsageServer = "a listen server handling booking requests"

	accountsAPIHost  = "accounts-api-host"
	lowribeckAPIHost = "lowribeck-api-host"

	eligibilityAPIHost = "eligibility-api-host"
	clickAPIHost       = "click-api-host"

	flagRedisAddr             = "redis-addr"
	flagRedisTTLHours         = "redis-ttl-hours"
	flagExpirationTimeSeconds = "flag-expiration-time-seconds"
	flagClickKeyID            = "flag-click-key-id"
	flagAuthScope             = "flag-auth-scope"
	flagWebLocation           = "flag-web-location"
	flagMobileLocation        = "flag-mobile-location"
	flagSubject               = "flag-subject"
	flagIntent                = "flag-intent"
	flagChannel               = "flag-channel"
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
			&cli.StringFlag{
				Name:     flagRedisAddr,
				EnvVars:  []string{"REDIS_ADDR"},
				Required: true,
			},
			&cli.IntFlag{
				Name:     flagRedisTTLHours,
				EnvVars:  []string{"REDIS_TTL_HOURS"},
				Required: true,
				Value:    6,
			},
			&cli.StringFlag{
				Name:    flagPartialBookingCron,
				EnvVars: []string{"PARTIAL_BOOKING_CRON"},
				Value:   "* * * * *",
			},
			&cli.StringFlag{
				Name:     eligibilityAPIHost,
				EnvVars:  []string{"ELIGIBILITY_API_HOST"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     clickAPIHost,
				EnvVars:  []string{"CLICK_API_HOST"},
				Required: true,
			},
			&cli.Int64Flag{
				Name:     flagExpirationTimeSeconds,
				EnvVars:  []string{"EXPIRATION_TIME_SECONDS"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagClickKeyID,
				EnvVars:  []string{"CLICK_KEY_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagAuthScope,
				EnvVars:  []string{"AUTH_SCOPE"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagWebLocation,
				EnvVars:  []string{"WEB_LOCATION"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagMobileLocation,
				EnvVars:  []string{"MOBILE_LOCATION"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagSubject,
				EnvVars:  []string{"SUBJECT"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagIntent,
				EnvVars:  []string{"INTENT"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagChannel,
				EnvVars:  []string{"CHANNEL"},
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

	pool, err := store.Setup(ctx, c.String(flagPostgresDSN))
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

	eligibilityConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(eligibilityAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to eligibility-grpc-api host [%s]: %w", c.String(eligibilityAPIHost), err)
	}
	opsServer.Add("eligibility-grpc-api", grpchealth.NewCheck(c.String(eligibilityAPIHost), "", "cannot connect to eligibility-grpc-api"))
	defer eligibilityConn.Close()

	clickUwConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(clickAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to click-uw-api host [%s]: %w", c.String(clickAPIHost), err)
	}
	defer clickUwConn.Close()

	clickClient := click.NewIssuerServiceClient(clickUwConn)

	bookingSink, err := app.GetKafkaSinkWithBroker(c.String(flagBookingTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to booking [%s] kafka sink: %w", c.String(flagBookingTopic), err)
	}
	defer bookingSink.Close()
	opsServer.Add("booking-sink", substratehealth.NewCheck(bookingSink, "unable to sink booking events"))

	commsSink, err := app.GetKafkaSinkWithBroker(c.String(flagBookingCommsTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to comms [%s] kafka sink: %w", c.String(flagBookingCommsTopic), err)
	}
	defer commsSink.Close()
	opsServer.Add("comms-sink", substratehealth.NewCheck(commsSink, "unable to sink comms events"))

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

	grpcServer := grpc.CreateServerWithLogLvl(c.String(app.GrpcLogLevel))
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(app.GrpcPort)))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port [%d]: %w", c.Int(app.GrpcPort), err)
	}
	defer listen.Close()

	redis := redis.NewClient(&redis.Options{Addr: c.String(flagRedisAddr)})
	hoursTTL := time.Duration(c.Int(flagRedisTTLHours)) * time.Hour
	eligibilityCache := store.NewMeterpointEligible(redis, hoursTTL)
	opsServer.Add("redis", eligibilityCache.NewHealthCheck())

	customerDetailsStore := store.NewAccountDetailsStore(redis, hoursTTL)
	opsServer.Add("redis-account-details", customerDetailsStore.NewHealthCheck())

	// GATEWAYS //
	accountGw := gateway.NewAccountGateway(mn, accountService.NewAccountServiceClient(accountsConn))
	lowriBeckGateway := gateway.NewLowriBeckGateway(mn, lowribeck_api.NewLowriBeckAPIClient(lowribeckConn))
	eligibilityGateway := gateway.NewEligibilityGateway(mn, eligibilityv1.NewEligiblityAPIClient(eligibilityConn))
	cachedEligibilityGateway := cache.NewMeterpointEligibilityCacheWrapper(eligibilityGateway, eligibilityCache)
	clickGw, err := gateway.NewClickLinkProvider(clickClient, &gateway.ClickLinkProviderConfig{
		ExpirationTimeSeconds: c.Int64(flagExpirationTimeSeconds),
		ClickKeyID:            c.String(flagClickKeyID),
		AuthScope:             c.String(flagAuthScope),
		WebLocation:           c.String(flagWebLocation),
		MobileLocation:        c.String(flagMobileLocation),
		Subject:               c.String(flagSubject),
		Intent:                c.String(flagIntent),
		Channel:               c.String(flagChannel),
	})
	if err != nil {
		return fmt.Errorf("failed to initialise click link provider, %w", err)
	}

	// PUBLISHERS //

	syncBookingPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(bookingSink), c.App.Name)
	syncCommsPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(commsSink), c.App.Name)

	// STORE //
	occupancyStore := store.NewOccupancy(pool)
	siteStore := store.NewSite(pool)
	bookingStore := store.NewBooking(pool)
	partialBookingStore := store.NewPartialBooking(pool)

	// DOMAIN //
	bookingDomain := domain.NewBookingDomain(
		accountGw,
		lowriBeckGateway,
		occupancyStore,
		siteStore,
		bookingStore,
		partialBookingStore,
		customerDetailsStore,
		cachedEligibilityGateway,
		clickGw,
		true,
	)

	bookingAPI := api.New(bookingDomain, syncBookingPublisher, syncCommsPublisher, auth, true)
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
