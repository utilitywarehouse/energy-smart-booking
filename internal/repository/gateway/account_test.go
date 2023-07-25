//go:generate mockgen -source=gateway.go -destination ./mocks/gateway_mocks.go

package gateway_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	v1 "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/address/v1"
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

func Test_GetAccountAddressByAccountID(t *testing.T) {
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
			Id: "account-id-1",
			SupplyDetails: &accountService.AddressDetails{
				Address: &v1.Address{
					Uprn: "uprn-1",
					Paf: &v1.Address_PAF{
						Organisation:            "org-1",
						Department:              "dep-1",
						SubBuilding:             "sub-building-1",
						BuildingName:            "building-1",
						BuildingNumber:          "building-nr-1",
						DependentThoroughfare:   "dependent-thoroughfare-1",
						Thoroughfare:            "thoroughfare-1",
						DoubleDependentLocality: "ddl-1",
						DependentLocality:       "dl-1",
						PostTown:                "post-town-1",
						Postcode:                "post-code-1",
					},
				},
			},
		},
	}, nil)

	actual := models.AccountAddress{
		UPRN: "uprn-1",
		PAF: models.PAF{
			Organisation:            "org-1",
			Department:              "dep-1",
			SubBuilding:             "sub-building-1",
			BuildingName:            "building-1",
			BuildingNumber:          "building-nr-1",
			DependentThoroughfare:   "dependent-thoroughfare-1",
			Thoroughfare:            "thoroughfare-1",
			DoubleDependentLocality: "ddl-1",
			DependentLocality:       "dl-1",
			PostTown:                "post-town-1",
			Postcode:                "post-code-1",
		},
	}

	expected, err := myGw.GetAccountAddressByAccountID(ctx, "account-id-1")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatal(diff)
	}
}
