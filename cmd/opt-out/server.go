package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/accounts"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/substratehealth"
	"github.com/utilitywarehouse/uwos-go/iam"
	"github.com/utilitywarehouse/uwos-go/iam/identity"
	"github.com/utilitywarehouse/uwos-go/iam/machine"
	"github.com/uw-labs/substrate"
	"golang.org/x/sync/errgroup"
)

func runServer(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	pool, err := store.Setup(ctx, c.String(postgresDSN))
	if err != nil {
		return err
	}
	defer pool.Close()

	db := store.NewAccountOptOut(pool)

	optOutSink, err := app.GetKafkaSink(c, c.String(optOutEventsTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to opt-out sink: %w", err)
	}
	opsServer.Add("opt-out-sink", substratehealth.NewCheck(optOutSink, "unable to publish opt-out events"))
	defer optOutSink.Close()

	syncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(optOutSink), appName)

	grpcConn, err := grpc.CreateConnection(ctx, c.String(accountsAPIHost))
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	accountsClient := accountService.NewNumberLookupServiceClient(grpcConn)

	mn, err := machine.New()
	if err != nil {
		return err
	}
	defer mn.Close()

	identityClient, err := identity.NewClient()
	if err != nil {
		return err
	}

	accountsRepo := accounts.NewAccountLookup(mn, accountsClient)

	router := mux.NewRouter()
	apiHandler := api.NewHandler(db, syncPublisher, accountsRepo, identityClient)
	apiHandler.Register(ctx, router)

	chain := alice.New()
	chain = chain.Append(api.EnableCORS, iam.HTTPHandler(true))
	httpHandler := chain.Then(router)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Int(httpServerPort)),
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer log.Info("server exited")
		return httpServer.ListenAndServe()
	})

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer log.Debug("signal handler finished")
		select {
		case <-ctx.Done():
			httpServer.Close()
			return ctx.Err()
		case <-sigChan:
			cancel()
			if err = httpServer.Shutdown(shutdownCtx); err != nil {
				return err
			}
		}
		return nil
	})

	return g.Wait()
}
