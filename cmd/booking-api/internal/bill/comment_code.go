package bill

import (
	"fmt"

	"github.com/utilitywarehouse/bill-contracts/go/pkg/generated/bill_contracts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/domain"
	"github.com/utilitywarehouse/uwos-go/x/bill/commentcode"
)

const requestSmartMeter = "Request smart meter"
const declineSmartMeter = "Decline smart meter"
const noReasonGiven = "Wouldn't or didn't give a reason"

func BuildCommentCode(smartMeterInterest *domain.SmartMeterInterest) (*bill_contracts.InboundEvent, error) {
	billInterested := MapInterested(smartMeterInterest.Interested)

	billReason, err := MapReason(smartMeterInterest.Interested, smartMeterInterest.Reason)
	if err != nil {
		return nil, err
	}

	r := commentcode.Record{
		AccountNumber:  smartMeterInterest.AccountNumber,
		Status:         commentcode.StatusOpenInternal,
		CommentCode:    commentcode.CommentCodeSmartMeterInterest,
		CreatedAt:      smartMeterInterest.CreatedAt.Local(),
		ComAdditional1: billInterested,
		ComAdditional2: billReason,
	}

	return r.Build(commentcode.WithID(smartMeterInterest.RegistrationID), commentcode.WithCreatedAt(smartMeterInterest.CreatedAt))
}

func MapInterested(interested bool) string {
	if interested {
		return requestSmartMeter
	}
	return declineSmartMeter
}

func MapReason(interested bool, reason *bookingv1.Reason) (string, error) {
	if reason == nil {
		return noReasonGiven, nil
	}

	if interested {
		switch *reason {
		case bookingv1.Reason_REASON_CONTROL:
			return "I want an In-Home Display so I can be in control of my energy usage", nil
		case bookingv1.Reason_REASON_AUTOMATIC_SENDING:
			return "I don't want to worry about sending my meter readings anymore", nil
		case bookingv1.Reason_REASON_ACCURACY:
			return "I want accurate bills that reflect exactly what I have used", nil
		default:
			return "", fmt.Errorf("invalid reason for smart meter interest: %s", reason)
		}
	} else {
		switch *reason {
		case bookingv1.Reason_REASON_HEALTH:
			return "I have health concerns", nil
		case bookingv1.Reason_REASON_COST:
			return "I have cost concerns", nil
		case bookingv1.Reason_REASON_TECHNOLOGY:
			return "I have technology concerns", nil
		case bookingv1.Reason_REASON_SECURITY_AND_PRIVACY:
			return "I have security and privacy concerns", nil
		case bookingv1.Reason_REASON_INCONVENIENCE:
			return "Having my meter changed is inconvenient", nil
		default:
			return "", fmt.Errorf("invalid reason for smart meter disinterest: %s", reason)
		}
	}
}
