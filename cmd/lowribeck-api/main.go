package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	grpcHelper "github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/mapper"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/go-operational/op"
	"github.com/utilitywarehouse/uwos-go/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/telemetry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"
)

const (
	appName = "energy-smart-booking-lowri-beck-api"
	appDesc = "communicates with lowri beck third party API and exposes a gRPC server"

	//LowriBeck config
	baseURL         = "base-url"
	sendingSystem   = "sending-system"
	receivingSystem = "receiving-system"
	authUser        = "auth-user"
	authPassword    = "auth-password"
	useHeathcheck   = "use-healthcheck"

	// LowriBeck job type codes
	electricityJobTypeCodeCredit     = "electricity-job-type-code-credit"
	electricityJobTypeCodePrepayment = "electricity-job-type-code-prepayment"
	gasJobTypeCodeCredit             = "gas-job-type-code-credit" //nolint: gosec
	gasJobTypeCodePrepayment         = "gas-job-type-code-prepayment"
)

var gitHash string // populated at compile time

func main() {
	app := &cli.App{
		Name:  appName,
		Usage: appDesc,
		Commands: []*cli.Command{
			{
				Name: "api",
				Flags: app.DefaultFlags().WithGrpc().WithCustom(
					&cli.StringFlag{
						Name:     baseURL,
						EnvVars:  []string{"BASE_URL"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     sendingSystem,
						EnvVars:  []string{"SENDING_SYSTEM"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     receivingSystem,
						EnvVars:  []string{"RECEIVING_SYSTEM"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     authUser,
						EnvVars:  []string{"AUTH_USER"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     authPassword,
						EnvVars:  []string{"AUTH_PASSWORD"},
						Required: true,
					},
					&cli.BoolFlag{
						Name:    useHeathcheck,
						EnvVars: []string{"USE_HEALTHCHECK"},
					},
					&cli.StringFlag{
						Name:     electricityJobTypeCodeCredit,
						EnvVars:  []string{"ELECTRICITY_JOB_TYPE_CODE_CREDIT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     electricityJobTypeCodePrepayment,
						EnvVars:  []string{"ELECTRICITY_JOB_TYPE_CODE_PREPAYMENT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     gasJobTypeCodeCredit,
						EnvVars:  []string{"GAS_JOB_TYPE_CODE_CREDIT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     gasJobTypeCodePrepayment,
						EnvVars:  []string{"GAS_JOB_TYPE_CODE_PREPAYMENT"},
						Required: true,
					},
				),
				Before: app.Before,
				Action: runServer,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("unable to run app", "error", err)
		os.Exit(1)
	}
}

func runServer(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	pdp, err := pdp.NewClient()
	if err != nil {
		return err
	}

	auth := auth.New(pdp.Multi())

	g, ctx := errgroup.WithContext(ctx)

	closer, err := telemetry.Register(ctx,
		telemetry.WithServiceName(appName),
		telemetry.WithTeam("energy-smart"),
		telemetry.WithServiceVersion(gitHash),
	)
	if err != nil {
		slog.Error("telemetry cannot be registered", "error", err)
	}
	defer closer.Close()

	httpClient := &http.Client{Timeout: 30 * time.Second}
	client := lowribeck.New(httpClient, c.String(authUser), c.String(authPassword), c.String(baseURL))

	if c.Bool(useHeathcheck) {
		opsServer.Add("lowribeck-api", lowribeckChecker(ctx, client.HealthCheck))
	}

	grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(app.GrpcLogLevel))
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(app.GrpcPort)))
	if err != nil {
		slog.Error("failed to listen on grpc port", "error", err)
		return err
	}
	defer listen.Close()

	mapper := mapper.NewLowriBeckMapper(c.String(sendingSystem),
		c.String(receivingSystem),
		c.String(electricityJobTypeCodeCredit),
		c.String(electricityJobTypeCodePrepayment),
		c.String(gasJobTypeCodeCredit),
		c.String(gasJobTypeCodePrepayment))

	lowribeckAPI := api.New(client, mapper, auth)
	contracts.RegisterLowriBeckAPIServer(grpcServer, lowribeckAPI)

	g.Go(func() error {
		defer slog.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer slog.Info("grpc server finished")
		return grpcServer.Serve(listen)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	g.Go(func() error {
		defer slog.Info("signal handler finished")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGTERM:
				slog.Info("cancelling context")
				cancel()
			}
		}
		return nil
	})

	return g.Wait()
}

func lowribeckChecker(ctx context.Context, healthCheckFn func(context.Context) error) func(cr *op.CheckResponse) {
	return func(cr *op.CheckResponse) {
		err := healthCheckFn(ctx)
		if err != nil {
			metrics.LBAPIRunning.Set(0.0)
			slog.Debug("health check got error", "error", err)
			cr.Degraded("health check failed "+err.Error(), "Check LowriBeck VPN connection/Third Party service provider")
			return
		}
		metrics.LBAPIRunning.Set(1.0)
		cr.Healthy("LowriBeck connection is healthy")
	}
}
