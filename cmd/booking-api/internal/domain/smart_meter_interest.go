package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/utilitywarehouse/bill-contracts/go/pkg/generated/bill_contracts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/bill"
)

var ErrInvalidSmartMeterInterestRequest = errors.New("invalid smart meter interest request")

type RegisterInterestParams struct {
	AccountID  string
	Interested bool
	Reason     bookingv1.Reason
}

// type SmartMeterInterest struct {
// 	RegistrationID string
// 	AccountNumber  string
// 	Interested     bool
// 	Reason         string
// 	CreatedAt      time.Time
// }

type SmartMeterInterestStore interface {
	Insert(ctx context.Context, requestID, accountID, reason string, interested bool, createdAt time.Time) error
}

type SmartMeterInterestDomain struct {
	accountNumber AccountNumberGateway
	interestStore SmartMeterInterestStore
}

func NewSmartMeterInterestDomain(accountNumberGateway AccountNumberGateway, interestStore SmartMeterInterestStore) SmartMeterInterestDomain {
	return SmartMeterInterestDomain{
		accountNumber: accountNumberGateway,
		interestStore: interestStore,
	}
}

func (s SmartMeterInterestDomain) RegisterInterest(ctx context.Context, params RegisterInterestParams) (*bill_contracts.InboundEvent, error) {
	accountNumber, err := s.accountNumber.Get(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	requestID := uuid.New().String()
	createdAt := time.Now().UTC()

	// Build comment code record into an InboundEvent.
	commentCodeEvent, err := bill.BuildCommentCode(requestID, accountNumber, params.Interested, params.Reason, createdAt)
	if err != nil {
		return nil, errors.Join(ErrInvalidSmartMeterInterestRequest, err)
	}

	s.interestStore.Insert(ctx, requestID, params.AccountID, params.Reason.String(), params.Interested, createdAt)

	return commentCodeEvent, nil
}
