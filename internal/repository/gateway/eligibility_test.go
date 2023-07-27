//go:generate mockgen -source=gateway.go -destination ./mocks/gateway_mocks.go

package gateway_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	eligibilityv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
)

func Test_GetEligibility(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	eC := mock_gateways.NewMockEligibilityClient(ctrl)
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	mai.EXPECT().ToCtx(ctx).Return(ctx)

	myGw := gateway.NewEligibilityGateway(mai, eC)

	eC.EXPECT().GetAccountOccupancyEligibleForSmartBooking(ctx, &eligibilityv1.GetAccountOccupancyEligibilityForSmartBookingRequest{
		AccountId:   "account-id-1",
		OccupancyId: "occupancy-id-1",
	}).Return(&eligibilityv1.GetAccountOccupancyEligibilityForSmartBookingResponse{
		Eligible:    true,
		AccountId:   "account-id-1",
		OccupancyId: "occupancy-id-1",
	}, nil)

	actual := true

	expected, err := myGw.GetEligibility(ctx, "account-id-1", "occupancy-id-1")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, actual) {
		t.Fatalf("expected: %t, actual: %t", expected, actual)
	}
}
