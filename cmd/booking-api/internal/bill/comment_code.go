package bill

import (
	"fmt"
	"time"

	"github.com/utilitywarehouse/bill-contracts/go/pkg/generated/bill_contracts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/uwos-go/x/bill/commentcode"
)

const unknownReason = "Wouldn't or didn't give a reason"

func BuildCommentCode(requestID, accountNumber string, interested bool, reason bookingv1.Reason, createdAt time.Time) (*bill_contracts.InboundEvent, error) {
	billInterested, billReason, err := MapReason(interested, reason)
	if err != nil {
		return nil, err
	}

	r := commentcode.Record{
		AccountNumber:  accountNumber,
		Status:         commentcode.StatusOpenInternal,
		CommentCode:    commentcode.CommentCodeSmartMeterInterest,
		ComAdditional1: billInterested,
		ComAdditional2: billReason,
	}

	return r.Build(commentcode.WithID(requestID), commentcode.WithCreatedAt(createdAt))
}

func MapReason(interested bool, reason bookingv1.Reason) (string, string, error) {
	var billInterested string
	var billReason string
	if interested {
		billInterested = "Request smart meter"
		switch reason {
		case bookingv1.Reason_REASON_CONTROL:
			billReason = "I want an In-Home Display so I can be in control of my energy usage"
		case bookingv1.Reason_REASON_AUTOMATIC_SENDING:
			billReason = "I don't want to worry about sending my meter readings anymore"
		case bookingv1.Reason_REASON_ACCURACY:
			billReason = "I want accurate bills that reflect exactly what I have used"
		case bookingv1.Reason_REASON_UNKNOWN:
			billReason = unknownReason
		default:
			return "", "", fmt.Errorf("invalid reason for customer interest: %s", reason)
		}
	} else {
		billInterested = "Decline smart meter"
		switch reason {
		case bookingv1.Reason_REASON_HEALTH:
			billReason = "I have health concerns"
		case bookingv1.Reason_REASON_COST:
			billReason = "I have cost concerns"
		case bookingv1.Reason_REASON_TECHNOLOGY:
			billReason = "I have technology concerns"
		case bookingv1.Reason_REASON_SECURITY_AND_PRIVACY:
			billReason = "I have security and privacy concerns"
		case bookingv1.Reason_REASON_INCONVENIENCE:
			billReason = "Having my meter changed is inconvenient"
		case bookingv1.Reason_REASON_UNKNOWN:
			billReason = unknownReason
		default:
			return "", "", fmt.Errorf("invalid reason for customer disinterest: %s", reason)
		}
	}

	return billInterested, billReason, nil
}
