package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_SmartMeterInterest_Insert(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	smartMeterInterestStore := store.NewSmartMeterInterestStore(db)
	timeNow := time.Now().UTC()

	input := models.SmartMeterInterest{
		RegistrationID: "registration-id",
		AccountID:      "account-id",
		Interested:     true,
		Reason:         "reason",
		CreatedAt:      timeNow,
	}

	assert := assert.New(t)

	err = smartMeterInterestStore.Insert(ctx, input)
	assert.NoError(err)

	result, err := smartMeterInterestStore.Get(ctx, "registration-id")
	assert.NoError(err)

	assert.EqualValues(&input, result)

	_, err = smartMeterInterestStore.Get(ctx, "unknown-registration-id")
	assert.EqualError(err, store.ErrRegistrationNotFound.Error())
}
