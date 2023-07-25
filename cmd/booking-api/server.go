package main

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	grpcHelper "github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/api"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"
)

var (
	commandNameServer  = "server"
	commandUsageServer = "a listen server handling booking requests"

	flagAccountsAPIHost = "accounts-api-host"

	flagRedisAddr = "redis-addr"
)

func init() {
	application.Commands = append(application.Commands, &cli.Command{
		Name:   commandNameServer,
		Usage:  commandUsageServer,
		Action: serverAction,
		Flags: app.DefaultFlags().WithGrpc().WithCustom(
			&cli.StringFlag{
				Name:     flagAccountsAPIHost,
				EnvVars:  []string{"ACCOUNTS_API_HOST"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagRedisAddr,
				EnvVars:  []string{"REDIS_ADDR"},
				Required: true,
			},
		),
	})
}

func serverAction(c *cli.Context) error {
	log.WithField("git_hash", gitHash).WithField("command", commandNameServer).Info("starting app")

	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	opsServer := makeOps(c)

	grpcServer := grpcHelper.CreateServerWithLogLvl(c.String(app.GrpcLogLevel))
	reflection.Register(grpcServer)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int(app.GrpcPort)))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port [%d]: %w", c.Int(app.GrpcPort), err)
	}
	defer listen.Close()

	// TODO: initialise client for LB API gateway here and pass to booking API
	bookingAPI := api.New(struct{}{})
	bookingv1.RegisterBookingAPIServer(grpcServer, bookingAPI)

	g.Go(func() error {
		defer log.Info("ops server finished")
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer log.Info("grpc server finished")
		return grpcServer.Serve(listen)
	})

	return g.Wait()
}
