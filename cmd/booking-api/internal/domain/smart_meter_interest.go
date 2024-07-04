package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var ErrInvalidSmartMeterInterestRequest = errors.New("invalid smart meter interest request")

type RegisterInterestParams struct {
	AccountID  string
	Interested bool
	Reason     *bookingv1.Reason
}

type SmartMeterInterest struct {
	RegistrationID string
	AccountNumber  string
	Interested     bool
	Reason         *bookingv1.Reason
	CreatedAt      time.Time
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

func (s SmartMeterInterestDomain) RegisterInterest(ctx context.Context, params RegisterInterestParams) (*SmartMeterInterest, error) {
	accountNumber, err := s.accountNumber.Get(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	registrationID := uuid.New().String()
	createdAt := time.Now().UTC()

	var reason string
	if params.Reason != nil {
		reason = params.Reason.String()
	}

	if err := s.interestStore.Insert(ctx, models.SmartMeterInterest{
		RegistrationID: registrationID,
		AccountID:      params.AccountID,
		Interested:     params.Interested,
		Reason:         reason,
		CreatedAt:      createdAt,
	}); err != nil {
		return nil, err
	}

	// Return account number NOT ID
	return &SmartMeterInterest{
		RegistrationID: registrationID,
		AccountNumber:  accountNumber,
		Interested:     params.Interested,
		Reason:         params.Reason,
		CreatedAt:      createdAt,
	}, nil
}
