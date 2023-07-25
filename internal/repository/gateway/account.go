package gateway

import (
	"context"
	"errors"
	"fmt"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrEmptyAddress = errors.New("account address is empty")
)

type AccountGateway struct {
	mai    MachineAuthInjector
	client AccountClient
}

func NewAccountGateway(mai MachineAuthInjector, client AccountClient) AccountGateway {
	return AccountGateway{
		mai, client,
	}
}

func (c AccountGateway) GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error) {
	account, err := c.client.GetAccount(c.mai.ToCtx(ctx), &accountService.GetAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to get account ID: %s, %w", accountID, err)
	}

	return models.Account{
		AccountID: account.GetAccount().Id,
		Details: models.AccountDetails{
			Title:     account.GetAccount().GetPrimaryAccountHolder().GetTitle(),
			FirstName: account.GetAccount().GetPrimaryAccountHolder().GetFirstName(),
			LastName:  account.GetAccount().GetPrimaryAccountHolder().GetLastName(),
			Email:     account.GetAccount().GetPrimaryAccountHolder().GetEmail(),
			Mobile:    account.GetAccount().GetPrimaryAccountHolder().GetMobile(),
		},
	}, nil
}

func (c AccountGateway) GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error) {
	account, err := c.client.GetAccount(c.mai.ToCtx(ctx), &accountService.GetAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		return models.AccountAddress{}, fmt.Errorf("failed to get account with account id: %s, %w", accountID, err)
	}

	address := account.GetAccount().GetSupplyDetails().GetAddress()

	if address == nil {
		return models.AccountAddress{}, ErrEmptyAddress
	}

	return models.AccountAddress{
		UPRN: address.GetUprn(),
		PAF: models.PAF{
			BuildingName:            address.GetPaf().BuildingName,
			BuildingNumber:          address.GetPaf().BuildingNumber,
			Department:              address.GetPaf().Department,
			DependentLocality:       address.GetPaf().DependentLocality,
			DependentThoroughfare:   address.GetPaf().DependentThoroughfare,
			DoubleDependentLocality: address.GetPaf().DoubleDependentLocality,
			Organisation:            address.GetPaf().Organisation,
			PostTown:                address.GetPaf().PostTown,
			Postcode:                address.GetPaf().Postcode,
			SubBuilding:             address.GetPaf().SubBuilding,
			Thoroughfare:            address.GetPaf().Thoroughfare,
		},
	}, nil
}
