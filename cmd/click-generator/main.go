package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/utilitywarehouse/energy-pkg/app"
)

const (
	appName = "energy-smart-booking-click-generator"
	appDesc = "Generats CTA links for smart booking eligible customer accounts"

	// Click
	clickAPIHost           = "click-api-host"
	clickSigningKeyID      = "click-signing-key-id"
	clickScope             = "click-scope"
	clickWebLocation       = "click-web-location"
	clickMobileLocation    = "click-mobile-location"
	clickLinkExpirySeconds = "click-link-expiry-seconds"

	// Tracking
	subject = "tracking-subject"
	intent  = "tracking-intent"
	channel = "tracking-channel"

	httpPort = "http-port"
)

var gitHash string

func main() {
	app := &cli.App{
		Name:  appName,
		Usage: appDesc,
		Commands: []*cli.Command{
			{
				Name: "api",
				Flags: app.DefaultFlags().WithCustom(
					&cli.StringFlag{
						Name:     clickAPIHost,
						EnvVars:  []string{"CLICK_API_HOST"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     clickSigningKeyID,
						EnvVars:  []string{"CLICK_SIGNING_KEY_ID"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     clickScope,
						EnvVars:  []string{"CLICK_SCOPE"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     clickWebLocation,
						EnvVars:  []string{"CLICK_WEB_LOCATION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     clickMobileLocation,
						EnvVars:  []string{"CLICK_MOBILE_LOCATION"},
						Required: true,
					},
					&cli.IntFlag{
						Name:     clickLinkExpirySeconds,
						EnvVars:  []string{"CLICK_LINK_EXPIRY_SECONDS"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     subject,
						EnvVars:  []string{"TRACKING_SUBJECT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     intent,
						EnvVars:  []string{"TRACKING_INTENT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     channel,
						EnvVars:  []string{"TRACKING_CHANNEL"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    httpPort,
						EnvVars: []string{"HTTP_SERVER_PORT"},
						Value:   8090,
					},
				),
				Before: app.Before,
				Action: runHTTPApi,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Panic("unable to run app")
	}
}
