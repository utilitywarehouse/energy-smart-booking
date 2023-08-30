package auth

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/uwos-go/v1/iam/identity"
	"github.com/utilitywarehouse/uwos-go/v1/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/v1/iam/principal"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrNotCustomer     = errors.New("principal is not customer")
)

type pdpClient interface {
	Authorize(ctx context.Context, principal *principal.Model, action string, resource *pdp.Resource) (pdp.AuthorizeResult, error)
}

type idClient interface {
	WhoAmI(ctx context.Context, in *principal.Model) (identity.WhoAmIResult, error)
}

type Authorize struct {
	pdpClient pdpClient
	idClient  idClient
}

func New(pdpClient pdpClient, idClient idClient) *Authorize {
	return &Authorize{
		pdpClient: pdpClient,
		idClient:  idClient,
	}
}

type PolicyParams struct {
	Action     string
	Resource   string
	ResourceID string
	Attributes map[string]any
}

const (
	BookingResource = "uw.energy-smart.v1.booking-api"

	ReadAction   = "read"
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
		logrus.Error("PDP Authorize error: ", err)
		return res.Allowed(), ErrUnauthenticated
	}

	return res.Allowed(), nil
}
