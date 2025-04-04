//go:generate mockgen -source=gateway.go -destination ./mocks/gateway_mocks.go

package gateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway"

	mock_gateways "github.com/utilitywarehouse/energy-smart-booking/internal/repository/gateway/mocks"
)

func Test_GetAccountByAccountID(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	m := mock_gateways.NewMockAccountClient(ctrl)
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	mai.EXPECT().ToCtx(ctx).Return(ctx)

	myGw := gateway.NewAccountGateway(mai, m)

	m.EXPECT().GetAccount(ctx, &accountService.GetAccountRequest{
		AccountId: "account-id-1",
	}).Return(&accountService.GetAccountResponse{
		Account: &accountService.Account{
			PrimaryAccountHolder: &accountService.Person{
				FullName:  "John Doe",
				Title:     "Mr.",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "johndoe@example.com",
				Mobile:    "999-0101",
			},
			Id: "account-id-1",
		},
	}, nil)

	actual := models.Account{
		AccountID: "account-id-1",
		Details: models.AccountDetails{
			Title:     "Mr.",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
			Mobile:    "999-0101",
		},
	}

	expected, err := myGw.GetAccountByAccountID(ctx, "account-id-1")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatal(diff)
	}
}

func Test_GetAccountByAccountID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	defer ctrl.Finish()

	m := mock_gateways.NewMockAccountClient(ctrl)
	mai := mock_gateways.NewMockMachineAuthInjector(ctrl)

	mai.EXPECT().ToCtx(ctx).Return(ctx)

	myGw := gateway.NewAccountGateway(mai, m)

	m.EXPECT().GetAccount(ctx, &accountService.GetAccountRequest{
		AccountId: "account-id-1",
	}).Return(&accountService.GetAccountResponse{
		Account: &accountService.Account{},
	}, status.Error(codes.NotFound, "not found"))

	expectedErr := fmt.Errorf("%w, %w", gateway.ErrAccountNotFound, status.Error(codes.NotFound, "not found"))

	_, actualErr := myGw.GetAccountByAccountID(ctx, "account-id-1")

	if diff := cmp.Diff(actualErr.Error(), expectedErr.Error()); diff != "" {
		t.Fatal(diff)
	}
}
