package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
)

func TestAccountLinkStore(t *testing.T) {
	ctx := context.Background()

	testContainer, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	db := store.NewLink(pool)

	err = db.Add(ctx, "link-account-A", "link1")
	assert.NoError(t, err)
	err = db.Add(ctx, "link-account-A", "link2")
	assert.NoError(t, err)

	link, err := db.Get(ctx, "link-account-A")
	assert.NoError(t, err)
	assert.Equal(t, "link2", link)

	err = db.Remove(ctx, "link-account-A")
	assert.NoError(t, err)

	_, err = db.Get(ctx, "link-account-A")
	assert.ErrorIs(t, err, store.ErrAccountLinkNotFound)
}
