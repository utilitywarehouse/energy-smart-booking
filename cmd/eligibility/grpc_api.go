package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	smart_booking "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	pkgapp "github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	grpcHelper "github.com/utilitywarehouse/energy-services/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	"github.com/utilitywarehouse/uwos-go/v1/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func runGRPCApi(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(pkgapp.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	g, ctx := errgroup.WithContext(ctx)

	pdp, err := pdp.NewClient()
	if err != nil {
		return err
	}

	auth := auth.New(pdp)

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
	serviceStore := store.NewService(pg)

	closer, err := telemetry.Register(ctx,
		telemetry.WithServiceName(appName),
		telemetry.WithTeam("energy-smart"),
		telemetry.WithServiceVersion(gitHash),
	)
	if err != nil {
		return fmt.Errorf("telemetry cannot be registered: %v", err)
	}
	defer closer.Close()

	g.Go(func() error {
		grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(grpcLogLevel))
		reflection.Register(grpcServer)

		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(grpcPort)))
		if err != nil {
			return err
		}
		defer listen.Close()

		eligibilityAPI := api.NewEligibilityGRPCApi(eligibilityStore, suppliabilityStore, occupancyStore, accountStore, serviceStore, auth)
		smart_booking.RegisterEligiblityAPIServer(grpcServer, eligibilityAPI)

		return grpcServer.Serve(listen)
	})

	// register http gw
	gwMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			UseEnumNumbers:  false,
			EmitUnpopulated: true,
			Indent:          " ",
		},
	}))

	httpServer := &http.Server{
		Addr:              net.JoinHostPort("", c.String(httpPort)),
		Handler:           gwMux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	grpcAddr := net.JoinHostPort("", c.String(grpcPort))

	err = smart_booking.RegisterEligiblityAPIHandlerFromEndpoint(ctx, gwMux, grpcAddr, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		return err
	}

	g.Go(func() error {
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		select {
		case <-ctx.Done():
			httpServer.Close()
			return ctx.Err()
		case <-sigChan:
			logrus.Info("cancelling context")
			cancel()
		}
		return nil
	})

	return g.Wait()
}
