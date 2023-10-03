package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrEmptyAddress    = errors.New("account address is empty")
	ErrAccountNotFound = errors.New("account was not found")
)

var accountAPIResponses = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "booking_api_account_response_total",
	Help: "The number of account api error responses made by status code",
}, []string{"status"})

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
		code := status.Convert(err).Code()
		accountAPIResponses.WithLabelValues(code.String()).Inc()
		switch code {
		case codes.NotFound:
			return models.Account{}, fmt.Errorf("%w, %w", ErrAccountNotFound, err)
		default:
			return models.Account{}, fmt.Errorf("failed to get account ID: %s, %w", accountID, err)
		}
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
