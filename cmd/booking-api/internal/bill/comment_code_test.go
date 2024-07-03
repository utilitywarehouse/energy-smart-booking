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
		reason           *bookingv1.Reason
		expectedInterest string
		expectedReason   string
		expectedError    string
	}{
		{
			name:             "Valid Interest",
			interested:       true,
			reason:           bookingv1.Reason_REASON_CONTROL.Enum(),
			expectedInterest: "Request smart meter",
			expectedReason:   "I want an In-Home Display so I can be in control of my energy usage",
		},
		{
			name:             "Valid Disinterest",
			interested:       false,
			reason:           bookingv1.Reason_REASON_HEALTH.Enum(),
			expectedInterest: "Decline smart meter",
			expectedReason:   "I have health concerns",
		},
		{
			name:             "Valid Interest with no reason",
			interested:       true,
			expectedInterest: "Request smart meter",
			expectedReason:   "Wouldn't or didn't give a reason",
		},
		{
			name:             "Valid Disinterest with no reason",
			interested:       false,
			expectedInterest: "Decline smart meter",
			expectedReason:   "Wouldn't or didn't give a reason",
		},
		{
			name:          "Invalid Interest",
			interested:    true,
			reason:        bookingv1.Reason_REASON_HEALTH.Enum(),
			expectedError: "invalid reason for smart meter interest: REASON_HEALTH",
		},
		{
			name:          "Invalid Disinterest",
			interested:    false,
			reason:        bookingv1.Reason_REASON_CONTROL.Enum(),
			expectedError: "invalid reason for smart meter disinterest: REASON_CONTROL",
		},
		{
			name:          "Invalid Interest with unknown reason",
			interested:    true,
			reason:        bookingv1.Reason_REASON_UNKNOWN.Enum(),
			expectedError: "invalid reason for smart meter interest: REASON_UNKNOWN",
		},
		{
			name:          "Invalid Disinterest with unknown reason",
			interested:    false,
			reason:        bookingv1.Reason_REASON_UNKNOWN.Enum(),
			expectedError: "invalid reason for smart meter disinterest: REASON_UNKNOWN",
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
