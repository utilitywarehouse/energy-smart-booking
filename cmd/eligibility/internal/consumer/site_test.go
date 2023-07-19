package consumer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
	"github.com/uw-labs/substrate"
)

func TestSiteConsumer(t *testing.T) {
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
	s := store.NewSite(pool)

	handler := HandleSite(s)

	siteEv1, err := test_common.MakeMessage(&platform.SiteDiscoveredEvent{
		SiteId: "siteID",
		Address: &platform.SiteAddress{
			Postcode: "postCode",
		},
	})
	assert.NoError(err)

	err = handler(ctx, []substrate.Message{siteEv1})
	assert.NoError(err, "failed to handle site discovered event")

	site, err := s.Get(ctx, "siteID")
	assert.NoError(err, "failed to get site")
	expected := store.Site{
		ID:       "siteID",
		PostCode: "postCode",
	}
	assert.Equal(expected, site, "site mismatch")
}
