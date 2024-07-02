package auth

import (
	"context"
	"fmt"

	"github.com/utilitywarehouse/uwos-go/iam/pdp"
)

type pdpClient interface {
	Authorize(ctx context.Context, action string, resource *pdp.Resource) (pdp.MultiAuthorizeResult, error)
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
	AccountResource        = "uw.energy.v1.account"
	AccountBookingResource = "uw.energy.v1.account.smart-meter-booking"
	LowribeckAPIResource   = "uw.energy.v1.lowribeck-wrapper-api"
	EligibilityResource    = "uw.energy.v1.account.smart-meter-booking-eligibility"
	POSResource            = "uw.energy.v1.point-of-sale-smart-meter-booking"

	GetAction    = "get"
	CreateAction = "create"
	UpdateAction = "update"
)

func (a *Authorize) Authorize(ctx context.Context, params *PolicyParams) (bool, error) {
	r := pdp.NewResource(params.Resource, params.ResourceID)
	if len(params.Attributes) > 0 {
		r.WithAttributes(params.Attributes)
	}

	res, err := a.pdpClient.Authorize(ctx, params.Action, r)
	if err != nil {
		return false, fmt.Errorf("failed to authorize, %w", err)
	}
	return res.AllowedAny(), nil
}
