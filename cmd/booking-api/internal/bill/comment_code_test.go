package bill_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/bill"
)

func Test_MapReason(t *testing.T) {
	testCases := []struct {
		name             string
		interested       bool
		reason           bookingv1.Reason
		expectedInterest string
		expectedReason   string
		expectedError    string
	}{
		{
			name:             "Valid Interest",
			interested:       true,
			reason:           bookingv1.Reason_REASON_CONTROL,
			expectedInterest: "Request smart meter",
			expectedReason:   "I want an In-Home Display so I can be in control of my energy usage",
		},
		{
			name:             "Valid Disinterest",
			interested:       false,
			reason:           bookingv1.Reason_REASON_HEALTH,
			expectedInterest: "Decline smart meter",
			expectedReason:   "I have health concerns",
		},
		{
			name:             "Valid Interest with no reason",
			interested:       true,
			reason:           bookingv1.Reason_REASON_UNKNOWN,
			expectedInterest: "Request smart meter",
			expectedReason:   "Wouldn't or didn't give a reason",
		},
		{
			name:             "Valid Disinterest with no reason",
			interested:       false,
			reason:           bookingv1.Reason_REASON_UNKNOWN,
			expectedInterest: "Decline smart meter",
			expectedReason:   "Wouldn't or didn't give a reason",
		},
		{
			name:          "Invalid Interest",
			interested:    true,
			reason:        bookingv1.Reason_REASON_HEALTH,
			expectedError: "invalid reason for customer interest: REASON_HEALTH",
		},
		{
			name:          "Invalid Disinterest",
			interested:    false,
			reason:        bookingv1.Reason_REASON_CONTROL,
			expectedError: "invalid reason for customer disinterest: REASON_CONTROL",
		},
	}

	assert := assert.New(t)

	for _, testCase := range testCases {
		interest, reason, err := bill.MapReason(testCase.interested, testCase.reason)
		if testCase.expectedError == "" {
			assert.NoError(err, testCase.name)
			assert.Equal(testCase.expectedInterest, interest, testCase.name)
			assert.Equal(testCase.expectedReason, reason, testCase.name)
		} else {
			assert.EqualError(err, testCase.expectedError, testCase.name)
		}
	}
}
