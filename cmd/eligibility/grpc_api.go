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
	ecoesv2 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/ecoes/v2"
	xoservev1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/xoserve/v1"
	pkgapp "github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/evaluation"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	grpchealth "github.com/utilitywarehouse/go-ops-health-checks/pkg/grpchealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	uwgrpc "github.com/utilitywarehouse/uwos-go/grpc"
	"github.com/utilitywarehouse/uwos-go/iam"
	"github.com/utilitywarehouse/uwos-go/iam/machine"
	"github.com/utilitywarehouse/uwos-go/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/telemetry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	mn, err := machine.New()
	if err != nil {
		return fmt.Errorf("unable to create new IAM machine, %w", err)
	}
	defer mn.Close()

	pdp, err := pdp.NewClient()
	if err != nil {
		return err
	}

	auth := auth.New(pdp.Multi())

	pg, err := store.Setup(ctx, c.String(postgresDSN))
	if err != nil {
		return fmt.Errorf("couldn't initialise database: %w", err)
	}
	defer pg.Close()
	opsServer.Add("db", sqlhealth.NewCheck(stdlib.OpenDB(*pg.Config().ConnConfig), "unable to connect to the DB"))

	ecoesConn, err := uwgrpc.NewClient(c.String(ecoesHost), uwgrpc.WithDialIAM(iam.WithMachine(mn)))
	if err != nil {
		return fmt.Errorf("could not connect to ecoes gRPC integration: %w", err)
	}
	opsServer.Add("ecoes-api", grpchealth.NewCheck(c.String(ecoesHost), "", "cannot find mpans address"))
	defer ecoesConn.Close()

	xoserveConn, err := uwgrpc.NewClient(c.String(xoserveHost), uwgrpc.WithDialIAM(iam.WithMachine(mn)))
	if err != nil {
		return fmt.Errorf("could not connect to xoserve gRPC integration: %w", err)
	}
	opsServer.Add("xoserve-api", grpchealth.NewCheck(c.String(xoserveHost), "", "cannot find mprns address"))
	defer xoserveConn.Close()

	// GATEWAYS //
	ecoesGateway := gateway.NewEcoesGateway(ecoesv2.NewEcoesServiceClient(ecoesConn))
	xoserveGateway := gateway.NewXOServeGateway(xoservev1.NewXoserveAPIClient(xoserveConn))

	eligibilityStore := store.NewEligibility(pg)
	suppliabilityStore := store.NewSuppliability(pg)
	occupancyStore := store.NewOccupancy(pg)
	accountStore := store.NewAccount(pg)
	serviceStore := store.NewService(pg)
	meterpointStore := store.NewMeterpoint(pg)
	postcodeStore := store.NewPostCode(pg)

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
		grpcServer := uwgrpc.NewServer(
			uwgrpc.WithServerNetwork("tcp"),
			uwgrpc.WithServerAddress(fmt.Sprintf(":%d", c.Int(grpcPort))),
		)

		eligibilityAPI := api.NewEligibilityGRPCApi(
			eligibilityStore,
			suppliabilityStore,
			occupancyStore,
			accountStore,
			serviceStore,
			auth,
			evaluation.NewMeterpointEvaluator(
				postcodeStore,
				meterpointStore,
				ecoesGateway,
				xoserveGateway,
			),
		)
		smart_booking.RegisterEligiblityAPIServer(grpcServer, eligibilityAPI)

		return grpcServer.ServeContext(ctx)
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
