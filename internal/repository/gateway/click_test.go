//go:generate mockgen -source=click.go -destination ./mocks/click_mocks.go

package gateway_test

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	click "github.com/utilitywarehouse/click.uw.co.uk/generated/contract"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
)

func Test_GenerateAuthenticated(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mockIssuerServiceClient := mock_gateways.NewMockClickIssuerServiceClient(ctrl)

	config := gateway.ClickLinkProviderConfig{
		ExpirationTimeSeconds: 300,
		ClickKeyID:            "smart_energy_meter_booking_journey",
		AuthScope:             "smart-meter-installation",
		WebLocation:           "https://myaccount.uw.co.uk/energy/smart/upgrade",
		MobileLocation:        "uw://services/energy/smart/upgrade",
		Subject:               "smart_meter_installation",
		Intent:                "appointment_booking",
		Channel:               "email",
	}

	myGw, err := gateway.NewClickLinkProvider(mockIssuerServiceClient, &config)
	if err != nil {
		t.Fatal(err)
	}

	mockIssuerServiceClient.EXPECT().IssueURL(ctx, &click.IssueURLRequest{
		KeyId: "smart_energy_meter_booking_journey",
		ValidFor: &types.Duration{
			Seconds: int64(300),
		},
		Target: &click.TargetSpec{
			Web:    "https://myaccount.uw.co.uk/energy/smart/upgrade",
			Mobile: "uw://services/energy/smart/upgrade",
		},
		Auth: &click.AuthSpec{
			Scope:         "smart-meter-installation",
			AccountNumber: "1001",
			Ttl: &types.Duration{
				Seconds: int64(300),
			},
		},
		Tracking: &click.TrackingSpec{
			Identity: "1001",
			Subject:  "smart_meter_installation",
			Intent:   "appointment_booking",
			Channel:  "email",
			Attributes: map[string]string{
				"journey_type": "point_of_sale",
			},
		},
	}).Return(&click.IssueURLResponse{
		Url: "https://click.uw.co.uk/your-link-is-ready",
	}, nil)

	actual := "https://click.uw.co.uk/your-link-is-ready"

	expected, err := myGw.GenerateAuthenticated(ctx, "1001", map[string]string{
		"journey_type": "point_of_sale",
	})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatal(diff)
	}
}

func Test_GenerateGeneric(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mockIssuerServiceClient := mock_gateways.NewMockClickIssuerServiceClient(ctrl)

	config := gateway.ClickLinkProviderConfig{
		ExpirationTimeSeconds: 300,
		ClickKeyID:            "smart_energy_meter_booking_journey",
		AuthScope:             "smart-meter-installation",
		WebLocation:           "https://myaccount.uw.co.uk/energy/smart/upgrade",
		MobileLocation:        "uw://services/energy/smart/upgrade",
		Subject:               "smart_meter_installation",
		Intent:                "appointment_booking",
		Channel:               "email",
	}

	myGw, err := gateway.NewClickLinkProvider(mockIssuerServiceClient, &config)
	if err != nil {
		t.Fatal(err)
	}

	mockIssuerServiceClient.EXPECT().IssueURL(ctx, &click.IssueURLRequest{
		KeyId: "smart_energy_meter_booking_journey",
		ValidFor: &types.Duration{
			Seconds: int64(300),
		},
		Target: &click.TargetSpec{
			Web:    "https://myaccount.uw.co.uk/energy/smart/upgrade",
			Mobile: "uw://services/energy/smart/upgrade",
		},
		Tracking: &click.TrackingSpec{
			Identity: "1001",
			Subject:  "smart_meter_installation",
			Intent:   "appointment_booking",
			Channel:  "email",
		},
	}).Return(&click.IssueURLResponse{
		Url: "https://click.uw.co.uk/your-link-is-ready",
	}, nil)

	actual := "https://click.uw.co.uk/your-link-is-ready"

	expected, err := myGw.GenerateGenericLink(ctx, "1001")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatal(diff)
	}
}
