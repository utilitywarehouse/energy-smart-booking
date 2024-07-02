package domain

import (
	"context"

	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type AccountGateway interface {
	GetAccountByAccountID(ctx context.Context, accountID string) (models.Account, error)
}

type AccountNumberGateway interface {
	Get(ctx context.Context, accountID string) (string, error)
}
