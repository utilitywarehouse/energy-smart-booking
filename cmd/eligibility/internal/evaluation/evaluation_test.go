package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
)

func TestPublishCampaignabilityIfChanged(t *testing.T) {
	ctx := context.Background()
	mockSync := test_common.MockSink{}
	evaluator := Evaluator{
		campaignabilitySync: &mockSync,
	}
	occupancy := &domain.Occupancy{
		ID: "occupancyID",
		Account: domain.Account{
			ID: "accountID",
		},
		Site:     nil,
		Services: nil,
		EvaluationResult: domain.OccupancyEvaluation{
			OccupancyID: "occupancyID",
		},
	}
	err := evaluator.publishCampaignabilityIfChanged(ctx, occupancy, domain.IneligibleReasons{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mockSync.Msgs))
	assert.Equal(t, &smart.CampaignableOccupancyAddedEvent{OccupancyId: "occupancyID", AccountId: "accountID"}, mockSync.Msgs[0])
}
