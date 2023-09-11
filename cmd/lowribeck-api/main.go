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

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	grpcHelper "github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/mapper"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/go-operational/op"
	"github.com/utilitywarehouse/uwos-go/v1/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry"
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
				),
				Before: app.Before,
				Action: runServer,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Panic("unable to run app")
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

	auth := auth.New(pdp)

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

	httpClient := &http.Client{Timeout: 30 * time.Second}

	client := lowribeck.New(httpClient, c.String(authUser), c.String(authPassword), c.String(baseURL))
	opsServer.Add("lowribeck-api", lowribeckChecker(ctx, client.HealthCheck))

	grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(app.GrpcLogLevel))
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(app.GrpcPort)))
	if err != nil {
		log.WithError(err).Panic("failed to listen on GRPC port")
	}
	defer listen.Close()

	mapper := mapper.NewLowriBeckMapper(c.String(sendingSystem), c.String(receivingSystem))
	lowribeckAPI := api.New(client, mapper, auth)
	contracts.RegisterLowriBeckAPIServer(grpcServer, lowribeckAPI)

	g.Go(func() error {
		defer log.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer log.Info("grpc server finished")
		return grpcServer.Serve(listen)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	g.Go(func() error {
		defer log.Info("signal handler finished")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGTERM:
				log.Info("cancelling context")
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
			log.Debugf("health check got error: %s", err)
			cr.Unhealthy("health check failed "+err.Error(), "Check LowriBeck VPN connection/Third Party service provider", "booking management and booking slots compromised")

			return
		}

		cr.Healthy("LowriBeck connection is healthy")
	}
}
