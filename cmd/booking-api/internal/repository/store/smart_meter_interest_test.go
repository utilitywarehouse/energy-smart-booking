package store_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_SmartMeterInterest_Insert(t *testing.T) {
	smartMeterInterestStore := store.NewSmartMeterInterestStore(pool)
	defer truncateDB(t)

	timeNow := time.Now().UTC()

	testCases := []struct {
		description string
		input       models.SmartMeterInterest
		output      models.SmartMeterInterest
	}{
		{
			description: "No Reason given",
			input: models.SmartMeterInterest{
				RegistrationID: "registration-id-1",
				AccountID:      "account-id-1",
				Interested:     true,
				CreatedAt:      timeNow,
			},
			output: models.SmartMeterInterest{
				RegistrationID: "registration-id-1",
				AccountID:      "account-id-1",
				Interested:     true,
				CreatedAt:      timeNow,
			},
		},
		{
			description: "Reason given",
			input: models.SmartMeterInterest{
				RegistrationID: "registration-id-2",
				AccountID:      "account-id-2",
				Interested:     false,
				Reason:         "reason",
				CreatedAt:      timeNow,
			},
			output: models.SmartMeterInterest{
				RegistrationID: "registration-id-2",
				AccountID:      "account-id-2",
				Interested:     false,
				Reason:         "reason",
				CreatedAt:      timeNow,
			},
		},
		{
			description: "Duplicate insert on registration ID",
			input: models.SmartMeterInterest{
				RegistrationID: "registration-id-1",
				AccountID:      "account-id-3",
				Interested:     false,
				CreatedAt:      timeNow,
			},
			output: models.SmartMeterInterest{
				RegistrationID: "registration-id-1",
				AccountID:      "account-id-1",
				Interested:     true,
				CreatedAt:      timeNow,
			},
		},
	}

	assert := assert.New(t)

	for _, testCase := range testCases {
		t.Run(testCase.description, func(_ *testing.T) {
			err := smartMeterInterestStore.Insert(t.Context(), testCase.input)
			assert.NoError(err, testCase.description)

			result, err := smartMeterInterestStore.Get(t.Context(), testCase.input.RegistrationID)
			assert.NoError(err, testCase.description)

			assert.Equal(&testCase.output, result, testCase.description)
		})
	}
}
