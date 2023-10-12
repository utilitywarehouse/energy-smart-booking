package gateway

import (
	"context"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"google.golang.org/grpc"
)

type MachineAuthInjector interface {
	ToCtx(context.Context) context.Context
}

type AccountClient interface {
	GetAccount(ctx context.Context, in *accountService.GetAccountRequest, opts ...grpc.CallOption) (*accountService.GetAccountResponse, error)
}

type LowriBeckClient interface {
	GetAvailableSlots(ctx context.Context, in *lowribeckv1.GetAvailableSlotsRequest, opts ...grpc.CallOption) (*lowribeckv1.GetAvailableSlotsResponse, error)
	CreateBooking(ctx context.Context, in *lowribeckv1.CreateBookingRequest, opts ...grpc.CallOption) (*lowribeckv1.CreateBookingResponse, error)
	GetAvailableSlotsPointOfSale(ctx context.Context, in *lowribeckv1.GetAvailableSlotsPointOfSaleRequest, opts ...grpc.CallOption) (*lowribeckv1.GetAvailableSlotsPointOfSaleResponse, error)
	CreateBookingPointOfSale(ctx context.Context, in *lowribeckv1.CreateBookingPointOfSaleRequest, opts ...grpc.CallOption) (*lowribeckv1.CreateBookingPointOfSaleResponse, error)
}
