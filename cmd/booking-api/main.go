package main

import (
	"os"

	log "github.com/sirupsen/logrus"
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
	flagCommsTopic                          = "comms-topic"
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
				Name:    flagCommsTopic,
				EnvVars: []string{"COMMS_TOPIC"},
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
		log.WithError(err).Fatalln("service terminated unexpectedly")
		os.Exit(1)
	}
}
