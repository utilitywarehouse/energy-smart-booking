package gateway

import (
	"context"
	"fmt"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountNumberGateway struct {
	mai    MachineAuthInjector
	client AccountNumberClient
}

func NewAccountNumberGateway(mai MachineAuthInjector, client AccountNumberClient) *AccountNumberGateway {
	return &AccountNumberGateway{mai, client}
}

func (gw *AccountNumberGateway) Get(ctx context.Context, accountID string) (string, error) {

	response, err := gw.client.AccountNumber(gw.mai.ToCtx(ctx), &accountService.AccountNumberRequest{
		AccountId: []string{accountID},
	})
	if err != nil {
		code := status.Convert(err).Code()
		accountAPIResponses.WithLabelValues(code.String()).Inc()
		switch code {
		case codes.NotFound:
			return "", fmt.Errorf("%w, %w", ErrAccountNotFound, err)
		default:
			return "", fmt.Errorf("failed to get account number for account ID: %s, %w", accountID, err)
		}
	}

	return response.AccountNumber[0], nil
}
