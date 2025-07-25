package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/stdlib"
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
	"github.com/utilitywarehouse/uwos-go/iam/machine"
	"github.com/utilitywarehouse/uwos-go/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/telemetry"
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

	flagCommentCodeTopic = "comment-code-topic"
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
			&cli.StringFlag{
				Name:     flagCommentCodeTopic,
				EnvVars:  []string{"COMMENT_CODE_TOPIC"},
				Required: true,
			},
		),
	})
}

func serverAction(c *cli.Context) error {
	slog.Info("starting app", "git_hash", gitHash, "command", commandNameServer)

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

	auth := auth.New(pdp.Multi())

	accountsConn, err := grpc.CreateConnectionWithLogLvl(ctx, c.String(accountsAPIHost), c.String(app.GrpcLogLevel))
	if err != nil {
		return fmt.Errorf("error connecting to accounts-api host [%s]: %w", c.String(accountsAPIHost), err)
	}
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

	commsPoSBookingSink, err := app.GetKafkaSinkWithBroker(c.String(flagBookingCommsTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to comms [%s] kafka sink: %w", c.String(flagBookingCommsTopic), err)
	}
	defer commsPoSBookingSink.Close()
	opsServer.Add("comms-pos-booking-sink", substratehealth.NewCheck(commsPoSBookingSink, "unable to sink point of sale bookings comms events"))

	commsRescheduleSink, err := app.GetKafkaSinkWithBroker(c.String(flagRescheduleCommsTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to comms [%s] kafka sink: %w", c.String(flagRescheduleCommsTopic), err)
	}
	defer commsRescheduleSink.Close()
	opsServer.Add("comms-reschedule-sink", substratehealth.NewCheck(commsRescheduleSink, "unable to sink reschedule comms events"))

	billCommentCodeSink, err := app.GetKafkaSinkWithBroker(c.String(flagCommentCodeTopic), c.String(app.KafkaVersion), c.StringSlice(app.KafkaBrokers))
	if err != nil {
		return fmt.Errorf("unable to connect to bill comment code [%s] kafka sink: %w", c.String(flagCommentCodeTopic), err)
	}
	defer billCommentCodeSink.Close()
	opsServer.Add("bill-comment-code", substratehealth.NewCheck(billCommentCodeSink, "unable to sink bill comment code events"))

	g, ctx := errgroup.WithContext(ctx)

	closer, err := telemetry.Register(ctx,
		telemetry.WithServiceName(appName),
		telemetry.WithTeam("energy-smart"),
		telemetry.WithServiceVersion(gitHash),
	)
	if err != nil {
		slog.Error("Telemetry cannot be registered", "error", err)
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
	accountNumberGw := gateway.NewAccountNumberGateway(mn, accountService.NewNumberLookupServiceClient(accountsConn))
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
	syncCommsPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(commsPoSBookingSink), c.App.Name)
	syncRescheduleCommsPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(commsRescheduleSink), c.App.Name)
	syncBillCommentCodePublisher := publisher.NewBillPublisher(substrate.NewSynchronousMessageSink(billCommentCodeSink))

	// STORE //
	occupancyStore := store.NewOccupancy(pool)
	siteStore := store.NewSite(pool)
	bookingStore := store.NewBooking(pool)
	partialBookingStore := store.NewPartialBooking(pool)
	smartMeterInterestStore := store.NewSmartMeterInterestStore(pool)

	// DOMAIN //
	bookingDomain := domain.NewBookingDomain(
		accountGw,
		accountNumberGw,
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

	interestDomain := domain.NewSmartMeterInterestDomain(
		accountNumberGw,
		smartMeterInterestStore,
	)

	bookingAPI := api.New(
		bookingDomain,
		interestDomain,
		syncBookingPublisher,
		syncCommsPublisher,
		syncRescheduleCommsPublisher,
		syncBillCommentCodePublisher,
		auth,
		true,
	)
	bookingv1.RegisterBookingAPIServer(grpcServer, bookingAPI)

	g.Go(func() error {
		defer slog.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer slog.Info("grpc server finished")
		return grpcServer.Serve(listen)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer slog.Info("signal handler finished")
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
