package gateway

import (
	"context"
	"fmt"

	eligibilityv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	"google.golang.org/grpc"
)

type EligibilityClient interface {
	GetMeterpointEligibility(ctx context.Context, in *eligibilityv1.GetMeterpointEligibilityRequest, opts ...grpc.CallOption) (*eligibilityv1.GetMeterpointEligibilityResponse, error)
}

type EligibilityGateway struct {
	mai    MachineAuthInjector
	client EligibilityClient
}

func NewEligibilityGateway(mai MachineAuthInjector, client EligibilityClient) *EligibilityGateway {
	return &EligibilityGateway{mai, client}
}

func (gw *EligibilityGateway) GetMeterpointEligibility(ctx context.Context, accountNumber, mpan, mprn, postcode string) (bool, error) {

	result, err := gw.client.GetMeterpointEligibility(gw.mai.ToCtx(ctx), &eligibilityv1.GetMeterpointEligibilityRequest{
		AccountNumber: accountNumber,
		Mpan:          mpan,
		Mprn:          toStr(mprn),
		Postcode:      postcode,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get meterpoint eligibilty, %w", err)
	}

	return result.GetEligible(), nil
}

func toStr(s string) *string {
	return &s
}
