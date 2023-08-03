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
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-services/grpc"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/generator"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/grpchealth"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	"golang.org/x/sync/errgroup"
)

func runHTTPApi(c *cli.Context) error {
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
	opsServer.Add("db", sqlhealth.NewCheck(stdlib.OpenDB(*pool.Config().ConnConfig), "unable to connect to the DB"))

	clickGRPCConn, err := grpc.CreateConnection(ctx, c.String(clickAPIHost))
	if err != nil {
		return err
	}
	opsServer.Add("click-api", grpchealth.NewCheckWithConnection(ctx, clickGRPCConn, "", "", "unable to query click api"))
	defer clickGRPCConn.Close()

	clickClient := click.NewIssuerServiceClient(clickGRPCConn)

	clickConfig := generator.LinkProviderConfig{
		ExpirationTimeSeconds: c.Int(clickLinkExpirySeconds),
		ClickKeyID:            c.String(clickSigningKeyID),
		AuthScope:             c.String(clickScope),
		Location:              c.String(clickWebLocation),
		MobileLocation:        c.String(clickMobileLocation),
		Subject:               c.String(subject),
		Intent:                c.String(intent),
		Channel:               c.String(channel),
	}
	linkProvider, err := generator.NewLinkProvider(clickClient, &clickConfig)
	if err != nil {
		return fmt.Errorf("failed to create link provider: %w", err)
	}

	router := mux.NewRouter()
	apiHandler := api.NewHandler(linkProvider)
	apiHandler.Register(ctx, router)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Int(httpPort)),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return opsServer.Start(ctx)
	})

	g.Go(func() error {
		defer logrus.Info("server exited")
		return httpServer.ListenAndServe()
	})

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer logrus.Debug("signal handler finished")
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
