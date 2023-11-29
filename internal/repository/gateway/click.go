package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gogo/protobuf/types"
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
	"google.golang.org/grpc"
)

var (
	ErrInvalidLocationProvided      = errors.New("invalid empty location provided for both web and mobile")
	ErrInvalidExpirationTimeSeconds = errors.New("invalid expiration time provided")
	ErrInvalidClickKeyID            = errors.New("invalid empty key provided")
)

type ClickLinkProviderConfig struct {
	ExpirationTimeSeconds int64
	ClickKeyID            string

	AuthScope      string
	WebLocation    string
	MobileLocation string

	// Tracking
	Subject string
	Intent  string
	Channel string
}

type ClickIssuerServiceClient interface {
	IssueURL(ctx context.Context, in *click.IssueURLRequest, opts ...grpc.CallOption) (*click.IssueURLResponse, error)
}

type ClickLinkProvider struct {
	issuerServiceClient ClickIssuerServiceClient
	config              *ClickLinkProviderConfig
}

func NewClickLinkProvider(client ClickIssuerServiceClient, config *ClickLinkProviderConfig) (*ClickLinkProvider, error) {
	err := config.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate link provider configs, %w", err)
	}

	return &ClickLinkProvider{
		issuerServiceClient: client,
		config: &ClickLinkProviderConfig{
			ExpirationTimeSeconds: config.ExpirationTimeSeconds,
			ClickKeyID:            config.ClickKeyID,
			AuthScope:             config.AuthScope,
			WebLocation:           config.WebLocation,
			MobileLocation:        config.MobileLocation,
			Subject:               config.Subject,
			Intent:                config.Intent,
			Channel:               config.Channel,
		},
	}, nil
}

func (p *ClickLinkProvider) GenerateAuthenticated(ctx context.Context, accountNo string, attributes map[string]string) (string, error) {

	baseURL, err := url.Parse(p.config.WebLocation)
	if err != nil {
		return "", fmt.Errorf("malformed URL: %v", err)
	}
	params := url.Values{}
	for key, value := range attributes {
		params.Set(key, value)
	}
	baseURL.RawQuery = params.Encode()

	clickLink, err := p.issuerServiceClient.IssueURL(ctx, &click.IssueURLRequest{
		KeyId: p.config.ClickKeyID,
		ValidFor: &types.Duration{
			Seconds: p.config.ExpirationTimeSeconds,
		},
		Target: &click.TargetSpec{
			Web: baseURL.String(),
		},
		Auth: &click.AuthSpec{
			Scope:         p.config.AuthScope,
			AccountNumber: accountNo,
			Ttl: &types.Duration{
				Seconds: p.config.ExpirationTimeSeconds,
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
		return "", fmt.Errorf("failed to generate authenticated click url: %w", err)
	}

	return clickLink.GetUrl(), nil
}

func (p *ClickLinkProvider) GenerateGenericLink(ctx context.Context, accountNo string) (string, error) {
	clickLink, err := p.issuerServiceClient.IssueURL(ctx, &click.IssueURLRequest{
		KeyId: p.config.ClickKeyID,
		ValidFor: &types.Duration{
			Seconds: p.config.ExpirationTimeSeconds,
		},
		Target: &click.TargetSpec{
			Web:    p.config.WebLocation,
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
		return "", fmt.Errorf("failed to generate generic click url: %w", err)
	}

	return clickLink.GetUrl(), nil
}

func (c *ClickLinkProviderConfig) validate() error {
	if c.WebLocation == "" && c.MobileLocation == "" {
		return ErrInvalidLocationProvided
	}

	if c.ExpirationTimeSeconds == 0 {
		return ErrInvalidExpirationTimeSeconds
	}

	if c.ClickKeyID == "" {
		return ErrInvalidClickKeyID
	}

	return nil
}
