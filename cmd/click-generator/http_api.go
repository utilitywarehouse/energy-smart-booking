package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/grpc"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	"golang.org/x/sync/errgroup"
)

func runHTTPApi(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	opsServer := ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)

	clickGRPCConn, err := grpc.CreateConnection(ctx, c.String(clickAPIHost))
	if err != nil {
		return err
	}
	defer clickGRPCConn.Close()

	clickClient := click.NewIssuerServiceClient(clickGRPCConn)

	clickConfig := gateway.ClickLinkProviderConfig{
		ExpirationTimeSeconds: c.Int64(clickLinkExpirySeconds),
		ClickKeyID:            c.String(clickSigningKeyID),
		AuthScope:             c.String(clickScope),
		WebLocation:           c.String(clickWebLocation),
		MobileLocation:        c.String(clickMobileLocation),
		Subject:               c.String(subject),
		Intent:                c.String(intent),
		Channel:               c.String(channel),
	}

	clickLinkProvider, err := gateway.NewClickLinkProvider(clickClient, &clickConfig)
	if err != nil {
		return fmt.Errorf("failed to create click link provider: %w", err)
	}

	router := mux.NewRouter()
	apiHandler := api.NewHandler(clickLinkProvider)
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
		defer slog.Info("server exited")
		return httpServer.ListenAndServe()
	})

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	g.Go(func() error {
		defer slog.Debug("signal handler finished")
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
