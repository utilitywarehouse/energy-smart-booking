package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
)

func TestWanCoverageConsumer(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)
	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)

	postgresURL, err := postgres.GetTestContainerDSN(container)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := postgres.Setup(ctx, postgresURL, migrations.Source)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = postgres.Teardown(pool, migrations.Source); err != nil {
			t.Fatal(err)
		}
	}()
	s := store.NewPostCode(pool)

	handler := HandleWanCoverage(s, nil, nil, true)

	wanCoverageEv1, err := test_common.MakeMessage(&smart.WanCoverageAtPostcodeStartedEvent{
		Postcode: "postCode",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{wanCoverageEv1})
	assert.NoError(err, "failed to handle wan coverage event")

	covered, err := s.GetWanCoverage(ctx, "postCode")
	assert.NoError(err, "failed to get wan coverage for post code")
	assert.True(covered)

	wanCoverageEv2, err := test_common.MakeMessage(&smart.WanCoverageAtPostcodeEndedEvent{
		Postcode: "postCode",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{wanCoverageEv2})
	assert.NoError(err, "failed to handle wan coverage removed event")

	covered, err = s.GetWanCoverage(ctx, "postCode")
	assert.NoError(err, "failed to get wan coverage for post code")
	assert.False(covered)
}
