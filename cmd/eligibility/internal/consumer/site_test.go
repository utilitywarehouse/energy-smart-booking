package consumer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/consumer"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/uw-labs/substrate"
)

func TestSiteConsumer(t *testing.T) {
	s := store.NewSite(pool)
	defer truncateDB(t)

	handler := consumer.HandleSite(s, nil, nil, true)

	siteEv1, err := makeMessage(&platform.SiteDiscoveredEvent{
		SiteId: "siteID",
		Address: &platform.SiteAddress{
			Postcode: "postCode",
		},
	})
	assert.NoError(t, err)

	err = handler(t.Context(), []substrate.Message{siteEv1})
	assert.NoError(t, err, "failed to handle site discovered event")

	site, err := s.Get(t.Context(), "siteID")
	assert.NoError(t, err, "failed to get site")
	expected := store.Site{
		ID:       "siteID",
		PostCode: "postCode",
	}
	assert.Equal(t, expected, site, "site mismatch")
}
