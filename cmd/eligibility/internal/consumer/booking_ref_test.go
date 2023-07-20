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

func TestBookingRefConsumer(t *testing.T) {
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
	s := store.NewBookingRef(pool)

	handler := HandleBookingRef(s)

	ev1, err := test_common.MakeMessage(&smart.BookingMpxnReferenceCreatedEvent{
		Mpxn:      "mpxn1",
		Reference: "ref1",
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{ev1})
	assert.NoError(err, "failed to handle booking ref event")

	ref, err := s.GetReference(ctx, "mpxn1")
	assert.NoError(err, "failed to get booking reference")
	assert.Equal("ref1", ref, "mismatch")
}
