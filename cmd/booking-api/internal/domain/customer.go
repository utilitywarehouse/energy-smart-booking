package domain

import (
	"context"

	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type AccountGateway interface {
	GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error)
	GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error)
}

type CustomerDomain struct {
	accounts AccountGateway
}

func NewCustomerDomain(accounts AccountGateway) CustomerDomain {
	return CustomerDomain{
		accounts,
	}
}

func (d CustomerDomain) GetCustomerContactDetails(ctx context.Context, accountID string) (models.Account, error) {
	return d.accounts.GetAccountByAccountID(ctx, accountID)
}

func (d CustomerDomain) GetAccountAddressByAccountID(ctx context.Context, accountID string) (models.AccountAddress, error) {
	return d.accounts.GetAccountAddressByAccountID(ctx, accountID)
}
