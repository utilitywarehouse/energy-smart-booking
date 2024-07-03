package bill

import (
	"fmt"

	"github.com/utilitywarehouse/bill-contracts/go/pkg/generated/bill_contracts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/uwos-go/x/bill/commentcode"
)

const unknownReason = "Wouldn't or didn't give a reason"

func BuildCommentCode(smartMeterInterest *domain.SmartMeterInterest) (*bill_contracts.InboundEvent, error) {
	billInterested, billReason, err := MapReason(smartMeterInterest.Interested, smartMeterInterest.Reason)
	if err != nil {
		return nil, err
	}

	r := commentcode.Record{
		AccountNumber:  smartMeterInterest.AccountNumber,
		Status:         commentcode.StatusOpenInternal,
		CommentCode:    commentcode.CommentCodeSmartMeterInterest,
		ComAdditional1: billInterested,
		ComAdditional2: billReason,
	}

	return r.Build(commentcode.WithID(smartMeterInterest.RegistrationID), commentcode.WithCreatedAt(smartMeterInterest.CreatedAt))
}

func MapReason(interested bool, reason *bookingv1.Reason) (string, string, error) {
	var billInterested string
	var billReason string
	if interested {
		billInterested = "Request smart meter"
		if reason == nil {
			billReason = unknownReason
		} else {
			switch *reason {
			case bookingv1.Reason_REASON_CONTROL:
				billReason = "I want an In-Home Display so I can be in control of my energy usage"
			case bookingv1.Reason_REASON_AUTOMATIC_SENDING:
				billReason = "I don't want to worry about sending my meter readings anymore"
			case bookingv1.Reason_REASON_ACCURACY:
				billReason = "I want accurate bills that reflect exactly what I have used"
			default:
				return "", "", fmt.Errorf("invalid reason for smart meter interest: %s", reason)
			}
		}
	} else {
		billInterested = "Decline smart meter"
		if reason == nil {
			billReason = unknownReason
		} else {
			switch *reason {
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
			default:
				return "", "", fmt.Errorf("invalid reason for smart meter disinterest: %s", reason)
			}
		}
	}

	return billInterested, billReason, nil
}
