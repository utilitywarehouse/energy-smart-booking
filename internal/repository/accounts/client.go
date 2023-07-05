package accounts

import (
	"context"

	accountService "github.com/utilitywarehouse/account-platform-protobuf-model/gen/go/account/api/v1"
)

type MachineAuthInjector interface {
	ToCtx(context.Context) context.Context
}

type Client struct {
	mai                MachineAuthInjector
	numberLookupClient accountService.NumberLookupServiceClient
}

func NewAccountLookup(mai MachineAuthInjector, client accountService.NumberLookupServiceClient) *Client {
	return &Client{
		mai:                mai,
		numberLookupClient: client,
	}
}

func (c *Client) AccountID(ctx context.Context, accountNumber string) (string, error) {
	reqCtx := c.mai.ToCtx(ctx)
	resp, err := c.numberLookupClient.AccountID(reqCtx, &accountService.AccountIDRequest{AccountNumber: []string{accountNumber}})
	if err != nil {
		return "", err
	}

	return resp.AccountId[0], nil
}
