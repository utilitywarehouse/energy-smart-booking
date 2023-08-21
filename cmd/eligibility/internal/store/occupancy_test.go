package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func TestOccupancy(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewOccupancy(connect(ctx))
	defer store.pool.Close()

	// add occupancy
	err := store.Add(ctx, "occupancy1", "site1", "account1", time.Now())
	assert.NoError(err, "failed to add occupancy")

	// get occupancy
	occupancy, err := store.Get(ctx, "occupancy1")
	assert.NoError(err, "failed to retrieve occupancy")
	assert.Equal(Occupancy{ID: "occupancy1", SiteID: "site1", AccountID: "account1"}, occupancy, "mismatch")

	// update occupancy
	err = store.AddSite(ctx, "occupancy1", "site2")
	assert.NoError(err, "failed to update occupancy")

	occupancy, err = store.Get(ctx, "occupancy1")
	assert.NoError(err, "failed to retrieve occupancy")
	assert.Equal(Occupancy{ID: "occupancy1", SiteID: "site2", AccountID: "account1"}, occupancy, "mismatch")

	_, err = store.Get(ctx, "occupancy2")
	assert.ErrorIs(err, ErrOccupancyNotFound)

	// get by accountID
	err = store.Add(ctx, "occupancy2", "site3", "account1", time.Now())
	assert.NoError(err, "failed to add occupancy")

	err = store.Add(ctx, "occupancy3", "site3", "account2", time.Now())
	assert.NoError(err, "failed to add occupancy")

	ids, err := store.GetIDsByAccount(ctx, "account1")
	assert.NoError(err, "failed to get occupancies by account ID")
	sort.Strings(ids)
	assert.Equal([]string{"occupancy1", "occupancy2"}, ids, "mismatch")

	// get by siteID
	ids, err = store.GetIDsBySite(ctx, "site3")
	assert.NoError(err, "failed to get occupancies by site ID")
	sort.Strings(ids)
	assert.Equal([]string{"occupancy2", "occupancy3"}, ids, "mismatch")

	q := `
		INSERT INTO services(id, occupancy_id, mpxn, supply_type, is_live)
		VALUES ('id1', 'occupancy_id1', 'mpxn', 'gas', true), 
		       ('id2', 'occupancy_id1', 'mpxn', 'elec', true),
		       ('id3', 'occupancy_id2', 'mpxn', 'gas', false),
		       ('id4', 'occupancy_id2', 'mpxn', 'elec', true);`

	_, err = store.pool.Exec(ctx, q)
	assert.NoError(err)

	ids, err = store.GetLiveOccupanciesPendingEvaluation(ctx)
	assert.NoError(err)
	sort.Strings(ids)
	assert.Equal([]string{"occupancy_id1", "occupancy_id2"}, ids)

	ids, err = store.GetLiveOccupancies(ctx)
	assert.NoError(err)
	sort.Strings(ids)
	assert.Equal([]string{"occupancy_id1", "occupancy_id2"}, ids)
}

func TestLoadOccupancy(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewOccupancy(connect(ctx))
	defer store.pool.Close()

	_, err := store.pool.Exec(ctx, `
	INSERT INTO occupancies(id, site_id, account_id, created_at) VALUES ('occupancyID1', 'siteID1', 'accountID1', now());
	INSERT INTO sites(id, post_code, created_at) VALUES ('siteID1', 'postcode', now());`)
	assert.NoError(err, "failed to prepare db")

	occupancy, err := store.LoadOccupancy(ctx, "occupancyID1")
	assert.NoError(err)
	expected := domain.Occupancy{
		ID:       "occupancyID1",
		Account:  domain.Account{ID: "accountID1"},
		Site:     &domain.Site{ID: "siteID1"},
		Services: nil,
		EvaluationResult: domain.OccupancyEvaluation{
			OccupancyID: "occupancyID1",
		},
	}
	assert.Equal(expected, occupancy)

	_, err = store.pool.Exec(ctx, `
	INSERT INTO postcodes(post_code, wan_coverage) VALUES ('postcode', true);`)
	assert.NoError(err, "failed to prepare db")
	occupancy, err = store.LoadOccupancy(ctx, "occupancyID1")
	assert.NoError(err)

	expected.Site = &domain.Site{
		ID:          "siteID1",
		Postcode:    "postcode",
		WanCoverage: true,
	}
	assert.Equal(expected, occupancy)

	_, err = store.pool.Exec(ctx, `
	INSERT INTO accounts(id, opt_out) VALUES ('accountID1', true);`)
	assert.NoError(err, "failed to prepare db")
	occupancy, err = store.LoadOccupancy(ctx, "occupancyID1")
	assert.NoError(err)

	expected.Account.OptOut = true
	assert.Equal(expected, occupancy)

	_, err = store.pool.Exec(ctx, `INSERT INTO eligibility (occupancy_id, account_id, reasons) VALUES ('occupancyID1', 'account1', '["ComplexTariff", "AlreadySmart"]');`)
	assert.NoError(err)
	occupancy, err = store.LoadOccupancy(ctx, "occupancyID1")
	assert.NoError(err)
	expected.EvaluationResult.Eligibility = domain.IneligibleReasons{domain.IneligibleReasonComplexTariff, domain.IneligibleReasonAlreadySmart}
	expected.EvaluationResult.EligibilityEvaluated = true
	assert.Equal(expected, occupancy)

	_, err = store.pool.Exec(ctx, `INSERT INTO campaignability (occupancy_id, account_id, reasons) VALUES ('occupancyID1', 'account1', '["OptOut"]');`)
	assert.NoError(err)
	occupancy, err = store.LoadOccupancy(ctx, "occupancyID1")
	assert.NoError(err)
	expected.EvaluationResult.Campaignability = domain.IneligibleReasons{domain.IneligibleReasonBookingOptOut}
	expected.EvaluationResult.CampaignabilityEvaluated = true
	assert.Equal(expected, occupancy)
}

func TestGetLiveOccupanciesIDsByAccountID(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	store := NewOccupancy(connect(ctx))
	defer store.pool.Close()

	_, err := store.pool.Exec(ctx, `
	INSERT INTO occupancies(id, site_id, account_id, created_at) 
	VALUES 
	    ('occupancy-id-A', 'site-id-A', 'account-ID-A', now()),
	    ('occupancy-id-B', 'site-id-A', 'account-ID-A', now());
	
	INSERT INTO services(id, mpxn, supply_type, is_live, occupancy_id) 
	VALUES 
	    ('service-id-A', 'service-mpxn-A', 'gas', true, 'occupancy-id-A'),
	    ('service-id-B', 'service-mpxn-B', 'elec', true, 'occupancy-id-A'),
	    ('service-id-C', 'service-mpxn-C', 'elec', false, 'occupancy-id-B');`)
	assert.NoError(err, "failed to prepare db")

	ids, err := store.GetLiveOccupanciesIDsByAccountID(ctx, "account-ID-A")
	assert.NoError(err)
	assert.Equal([]string{"occupancy-id-A"}, ids)

}
