package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/testcommon"
	"github.com/uw-labs/substrate"
)

func TestAccountOptOutConsumer(t *testing.T) {
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
	s := store.NewAccount(pool)

	handler := HandleAccountOptOut(s, nil, nil, true)

	ev1, err := testcommon.MakeMessage(&smart.AccountBookingOptOutAddedEvent{
		AccountId: "accountID",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{ev1})
	assert.NoError(err, "failed to handle account opt out event")

	account, err := s.GetAccount(ctx, "accountID")
	assert.NoError(err, "failed to get account")
	expected := store.Account{
		ID:       "accountID",
		PSRCodes: nil,
		OptOut:   true,
	}
	assert.Equal(expected, account, "mismatch")

	ev2, err := testcommon.MakeMessage(&smart.AccountBookingOptOutRemovedEvent{
		AccountId: "accountID",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{ev2})
	assert.NoError(err, "failed to handle account opt out event")

	account, err = s.GetAccount(ctx, "accountID")
	assert.NoError(err, "failed to get account")
	expected.OptOut = false
	assert.Equal(expected, account, "mismatch")
}
