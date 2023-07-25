package gateway

import (
	"context"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"google.golang.org/grpc"
)

type MachineAuthInjector interface {
	ToCtx(context.Context) context.Context
}

type AccountClient interface {
	GetAccount(ctx context.Context, in *accountService.GetAccountRequest, opts ...grpc.CallOption) (*accountService.GetAccountResponse, error)
}
