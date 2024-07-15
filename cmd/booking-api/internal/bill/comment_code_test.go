package bill_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/bill"
)

func Test_MapReason(t *testing.T) {
	testCases := []struct {
		description      string
		interested       bool
		reason           *bookingv1.Reason
		expectedInterest string
		expectedReason   string
		expectedError    string
	}{
		{
			description:      "Valid Interest",
			interested:       true,
			reason:           bookingv1.Reason_REASON_CONTROL.Enum(),
			expectedInterest: "Request smart meter",
			expectedReason:   "I want an In-Home Display so I can be in control of my energy usage",
		},
		{
			description:      "Valid Disinterest",
			interested:       false,
			reason:           bookingv1.Reason_REASON_HEALTH.Enum(),
			expectedInterest: "Decline smart meter",
			expectedReason:   "I have health concerns",
		},
		{
			description:      "Valid Interest with no reason",
			interested:       true,
			expectedInterest: "Request smart meter",
			expectedReason:   "Wouldn't or didn't give a reason",
		},
		{
			description:      "Valid Disinterest with no reason",
			interested:       false,
			expectedInterest: "Decline smart meter",
			expectedReason:   "Wouldn't or didn't give a reason",
		},
		{
			description:   "Invalid Interest",
			interested:    true,
			reason:        bookingv1.Reason_REASON_HEALTH.Enum(),
			expectedError: "invalid reason for smart meter interest: REASON_HEALTH",
		},
		{
			description:   "Invalid Disinterest",
			interested:    false,
			reason:        bookingv1.Reason_REASON_CONTROL.Enum(),
			expectedError: "invalid reason for smart meter disinterest: REASON_CONTROL",
		},
		{
			description:      "Unknown reason, interested",
			interested:       true,
			reason:           bookingv1.Reason_REASON_UNKNOWN.Enum(),
			expectedReason:   "Wouldn't or didn't give a reason",
			expectedInterest: "Request smart meter",
		},
		{
			description:      "Unknown reason, not interested",
			interested:       false,
			reason:           bookingv1.Reason_REASON_UNKNOWN.Enum(),
			expectedReason:   "Wouldn't or didn't give a reason",
			expectedInterest: "Decline smart meter",
		},
	}

	assert := assert.New(t)

	for _, testCase := range testCases {
		t.Run(testCase.description, func(_ *testing.T) {
			interest := bill.MapInterested(testCase.interested)
			reason, err := bill.MapReason(testCase.interested, testCase.reason)
			if testCase.expectedError == "" {
				assert.NoError(err, testCase.description)
				assert.Equal(testCase.expectedInterest, interest, testCase.description)
				assert.Equal(testCase.expectedReason, reason, testCase.description)
			} else {
				assert.EqualError(err, testCase.expectedError, testCase.description)
			}
		})
	}
}
