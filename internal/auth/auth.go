package auth

import (
	"context"
	"fmt"

	"github.com/utilitywarehouse/uwos-go/v1/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/v1/iam/principal"
)

type pdpClient interface {
	Authorize(ctx context.Context, principal *principal.Model, action string, resource *pdp.Resource) (pdp.AuthorizeResult, error)
}

type Authorize struct {
	pdpClient pdpClient
}

func New(pdpClient pdpClient) *Authorize {
	return &Authorize{
		pdpClient: pdpClient,
	}
}

type PolicyParams struct {
	Action     string
	Resource   string
	ResourceID string
	Attributes map[string]any
}

const (
	AccountResource        = "uw.energy-smart.v1.account"
	AccountBookingResource = "uw.energy-smart.v1.account.booking"

	GetAction    = "get"
	CreateAction = "create"
	UpdateAction = "update"
)

func (a *Authorize) Authorize(ctx context.Context, params *PolicyParams) (bool, error) {
	pr := principal.FromCtx(ctx)
	r := pdp.NewResource(params.Resource, params.ResourceID)
	if len(params.Attributes) > 0 {
		r.WithAttributes(params.Attributes)
	}

	res, err := a.pdpClient.Authorize(ctx, pr, params.Action, r)
	if err != nil {
		return false, fmt.Errorf("failed to authorize, %w", err)
	}

	return res.Allowed(), nil
}
