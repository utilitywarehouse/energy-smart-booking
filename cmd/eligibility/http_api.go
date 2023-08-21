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
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/api"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/evaluation"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/go-ops-health-checks/v3/pkg/sqlhealth"
	"github.com/uw-labs/substrate"
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

	meterStore := store.NewMeter(pool)
	occupancyStore := store.NewOccupancy(pool)
	serviceStore := store.NewService(pool)

	eligibilitySink, err := app.GetKafkaSink(c, c.String(eligibilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to eligibility sink: %w", err)
	}
	eligibilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(eligibilitySink), appName)

	suppliabilitySink, err := app.GetKafkaSink(c, c.String(suppliabilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to suppliability sink: %w", err)
	}
	suppliabilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(suppliabilitySink), appName)

	campaignabilitySink, err := app.GetKafkaSink(c, c.String(campaignabilityTopic))
	if err != nil {
		return fmt.Errorf("unable to connect to campaignability sink: %w", err)
	}
	campaignabilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(campaignabilitySink), appName)

	bookingEligibilitySink, err := app.GetKafkaSinkWithKeyFunc(c, c.String(bookingJourneyEligibilityTopic), keyFunc)
	if err != nil {
		return fmt.Errorf("unable to create booking journey eligibility sink: %w", err)
	}
	bookingEligibilitySyncPublisher := publisher.NewSyncPublisher(substrate.NewSynchronousMessageSink(bookingEligibilitySink), appName)

	evaluator := evaluation.NewEvaluator(
		occupancyStore,
		serviceStore,
		meterStore,
		eligibilitySyncPublisher,
		suppliabilitySyncPublisher,
		campaignabilitySyncPublisher,
		bookingEligibilitySyncPublisher,
	)

	router := mux.NewRouter()
	apiHandler := api.NewHandler(occupancyStore, evaluator)
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
