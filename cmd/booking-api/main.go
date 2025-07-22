package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
	"github.com/utilitywarehouse/energy-pkg/ops"
)

const (
	appName = "energy-smart-booking-api"
	appDesc = "receives booking requests, checks customer information and forwards to provider-specific API"
)

var (
	gitHash string

	// shared flags
	flagPostgresDSN                         = "postgres-dsn"
	flagBookingTopic                        = "booking-topic"
	flagBookingCommsTopic                   = "booking-comms-topic"
	flagRescheduleCommsTopic                = "reschedule-comms-topic"
	flagPartialBookingCron                  = "partial-booking-cron"
	flagRetainedBookingPeriodAlertThreshold = "retained-booking-period-alert-threshold"

	application = &cli.App{
		Name:   appName,
		Usage:  appDesc,
		Before: app.Before,
		Flags: app.DefaultFlags().WithKafka().WithCustom(
			&cli.StringFlag{
				Name:    flagBookingTopic,
				EnvVars: []string{"BOOKING_TOPIC"},
			},
			&cli.StringFlag{
				Name:    flagBookingCommsTopic,
				EnvVars: []string{"BOOKING_COMMS_TOPIC"},
			},
			&cli.StringFlag{
				Name:    flagRescheduleCommsTopic,
				EnvVars: []string{"RESCHEDULE_BOOKING_COMMS_TOPIC"},
			},
		),
	}
)

func makeOps(c *cli.Context) *ops.Server {
	return ops.Default().
		WithPort(c.Int(app.OpsPort)).
		WithHash(gitHash).
		WithDetails(appName, appDesc)
}

func main() {
	if err := application.Run(os.Args); err != nil {
		slog.Error("service terminated unexpectadly", "error", err)
		os.Exit(1)
	}
}
