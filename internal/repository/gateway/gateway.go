package gateway

import (
	"context"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	eligibilityv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"google.golang.org/grpc"
)

type MachineAuthInjector interface {
	ToCtx(context.Context) context.Context
}

type AccountClient interface {
	GetAccount(ctx context.Context, in *accountService.GetAccountRequest, opts ...grpc.CallOption) (*accountService.GetAccountResponse, error)
}

type EligibilityClient interface {
	GetAccountOccupancyEligibleForSmartBooking(ctx context.Context, in *eligibilityv1.GetAccountOccupancyEligibilityForSmartBookingRequest, opts ...grpc.CallOption) (*eligibilityv1.GetAccountOccupancyEligibilityForSmartBookingResponse, error)
}

type LowriBeckClient interface {
	GetAvailableSlots(ctx context.Context, in *lowribeckv1.GetAvailableSlotsRequest, opts ...grpc.CallOption) (*lowribeckv1.GetAvailableSlotsResponse, error)
	CreateBooking(ctx context.Context, in *lowribeckv1.CreateBookingRequest, opts ...grpc.CallOption) (*lowribeckv1.CreateBookingResponse, error)
}
