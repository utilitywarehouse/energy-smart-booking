package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	ids, err = store.GetLiveOccupancies(ctx)
	assert.NoError(err)

	sort.Strings(ids)
	assert.Equal([]string{"occupancy_id1", "occupancy_id2"}, ids)

}
