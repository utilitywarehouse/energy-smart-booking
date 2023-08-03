package generator

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
)

type LinkProviderConfig struct {
	ExpirationTimeSeconds int
	ClickKeyID            string

	AuthScope      string
	Location       string
	MobileLocation string

	// Tracking
	Subject string
	Intent  string
	Channel string
}

type LinkProvider struct {
	clickGRPC click.IssuerServiceClient
	config    *LinkProviderConfig
}

func NewLinkProvider(client click.IssuerServiceClient, config *LinkProviderConfig) (*LinkProvider, error) {
	err := config.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &LinkProvider{
		clickGRPC: client,
		config:    config,
	}, nil
}

func (p *LinkProvider) GenerateAuthenticated(ctx context.Context, accountNo string) (string, error) {
	clickLink, err := p.clickGRPC.IssueURL(ctx, &click.IssueURLRequest{
		KeyId: p.config.ClickKeyID,
		ValidFor: &types.Duration{
			Seconds: int64(p.config.ExpirationTimeSeconds),
		},
		Target: &click.TargetSpec{
			Web:    p.config.Location,
			Mobile: p.config.MobileLocation,
		},
		Auth: &click.AuthSpec{
			Scope:         p.config.AuthScope,
			AccountNumber: accountNo,
			Ttl: &types.Duration{
				Seconds: int64(p.config.ExpirationTimeSeconds),
			},
		},
		Tracking: &click.TrackingSpec{
			Identity: accountNo,
			Subject:  p.config.Subject,
			Intent:   p.config.Intent,
			Channel:  p.config.Channel,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate click url: %w", err)
	}

	return clickLink.GetUrl(), nil
}

func (p *LinkProvider) GenerateGenericLink(ctx context.Context, accountNo string) (string, error) {
	clickLink, err := p.clickGRPC.IssueURL(ctx, &click.IssueURLRequest{
		Target: &click.TargetSpec{
			Web:    p.config.Location,
			Mobile: p.config.MobileLocation,
		},
		Tracking: &click.TrackingSpec{
			Identity: accountNo,
			Subject:  p.config.Subject,
			Intent:   p.config.Intent,
			Channel:  p.config.Channel,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate click url: %w", err)
	}

	return clickLink.GetUrl(), nil
}

func (c *LinkProviderConfig) validate() error {
	if c.Location == "" && c.MobileLocation == "" {
		return fmt.Errorf("invalid empty location provided for both web and mobile")
	}

	if c.ExpirationTimeSeconds == 0 {
		return fmt.Errorf("invalid expiration time provided")
	}

	if c.ClickKeyID == "" {
		return fmt.Errorf("invalid empty key provided")
	}

	return nil
}
