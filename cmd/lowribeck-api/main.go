package main

import (
	"context"
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
	"github.com/utilitywarehouse/energy-pkg/ops"
	grpcHelper "github.com/utilitywarehouse/energy-services/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/lowribeck"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"
)

const (
	appName = "energy-smart-booking-opt-out"
	appDesc = "handles energy smart booking account opt outs"

	baseURL      = "base-url"
	authUser     = "auth-user"
	authPassword = "auth-password"
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

	g, ctx := errgroup.WithContext(ctx)

	httpClient := &http.Client{Timeout: 30 * time.Second}

	g.Go(func() error {
		grpcServer := grpcHelper.CreateServerWithLogLvl(app.GrpcLogLevel)
		reflection.Register(grpcServer)

		listen, err := net.Listen("tcp", app.GrpcPort)
		if err != nil {
			log.WithError(err).Panic("failed to listen on GRPC port")
		}
		defer listen.Close()

		client := lowribeck.New(httpClient, c.String(authUser), c.String(authPassword), c.String(baseURL))
		lowribeckAPI := api.New(client)
		contracts.RegisterLowriBeckAPIServer(grpcServer, lowribeckAPI)

		return grpcServer.Serve(listen)
	})

	g.Go(func() error {
		return opsServer.Start(ctx)
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
