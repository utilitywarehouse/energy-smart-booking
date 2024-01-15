package gateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"
	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_GetAccountNumberByAccountID(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	m := mock_gateways.NewMockAccountNumberClient(ctrl)
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	mai.EXPECT().ToCtx(ctx).Return(ctx)

	myGw := gateway.NewAccountNumberGateway(mai, m)

	m.EXPECT().AccountNumber(ctx, &accountService.AccountNumberRequest{
		AccountId: []string{"account-id-1"},
	}).Return(&accountService.AccountNumberResponse{
		AccountNumber: []string{"80001"},
	}, nil)

	actual := "80001"

	expected, err := myGw.Get(ctx, "account-id-1")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatal(diff)
	}
}

func Test_GetAccountNumberByAccountID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	m := mock_gateways.NewMockAccountNumberClient(ctrl)
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	mai.EXPECT().ToCtx(ctx).Return(ctx)

	myGw := gateway.NewAccountNumberGateway(mai, m)

	m.EXPECT().AccountNumber(ctx, &accountService.AccountNumberRequest{
		AccountId: []string{"account-id-1"},
	}).Return(&accountService.AccountNumberResponse{
		AccountNumber: []string{},
	}, status.Error(codes.NotFound, "not found"))

	expectedErr := fmt.Errorf("%w, %w", gateway.ErrAccountNotFound, status.Error(codes.NotFound, "not found"))

	_, actualErr := myGw.Get(ctx, "account-id-1")

	if diff := cmp.Diff(actualErr.Error(), expectedErr.Error()); diff != "" {
		t.Fatal(diff)
	}
}
