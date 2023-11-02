//go:generate mockgen -source=eligibility.go -destination ./mocks/eligibility_mocks.go

package gateway_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	eligibilityv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
)

func Test_Eligibility_GetMeterpointEligibility(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	mEligibility := mock_gateways.NewMockEligibilityClient(ctrl)
	mai := fakeMachineAuthInjector{}
	mai.ctx = ctx

	myGw := gateway.NewEligibilityGateway(mai, mEligibility)

	mEligibility.EXPECT().GetMeterpointEligibility(ctx, &eligibilityv1.GetMeterpointEligibilityRequest{
		AccountNumber: "14010",
		Mpan:          "10301031",
		Mprn:          toStr("120301230"),
		Postcode:      "E2 1ZZ",
	}).Return(&eligibilityv1.GetMeterpointEligibilityResponse{
		Eligible: true,
	}, nil)

	actual := true

	expected, err := myGw.GetMeterpointEligibility(ctx, "14010", "10301031", "120301230", "E2 1ZZ")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual, cmpopts.IgnoreUnexported()) {
		t.Fatalf("expected: %+v, actual: %+v", expected, actual)
	}
}

func toStr(s string) *string {
	return &s
}
