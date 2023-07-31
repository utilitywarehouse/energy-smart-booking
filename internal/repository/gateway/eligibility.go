package gateway

import (
	"context"
	"fmt"

	eligibilityv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
)

type EligibilityGateway struct {
	mai    MachineAuthInjector
	client EligibilityClient
}

func NewEligibilityGateway(mai MachineAuthInjector, client EligibilityClient) EligibilityGateway {
	return EligibilityGateway{mai, client}
}

func (g EligibilityGateway) GetEligibility(ctx context.Context, accountID, occupancyID string) (bool, error) {
	eligibility, err := g.client.GetAccountOccupancyEligibleForSmartBooking(g.mai.ToCtx(ctx), &eligibilityv1.GetAccountOccupancyEligibilityForSmartBookingRequest{
		AccountId:   accountID,
		OccupancyId: occupancyID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to call get account occupancy eligible for smart booking with accountId, occupancyID: [%s|%s], %w", accountID, occupancyID, err)
	}

	return eligibility.GetEligible(), nil
}
