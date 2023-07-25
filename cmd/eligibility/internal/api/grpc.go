package api

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	smart_booking "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EligibilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Eligibility, error)
}

type SuppliabilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Suppliability, error)
}

type OccupancyStore interface {
	GetIDsByAccount(ctx context.Context, accountID string) ([]string, error)
}

type AccountStore interface {
	GetAccount(ctx context.Context, accountID string) (store.Account, error)
}

type EligibilityGRPCApi struct {
	smart_booking.UnimplementedEligiblityAPIServer
	eligibilityStore   EligibilityStore
	suppliabilityStore SuppliabilityStore
	occupancyStore     OccupancyStore
	accountStore       AccountStore
}

func NewEligibilityGRPCApi(
	eligibilityStore EligibilityStore,
	suppliabilityStore SuppliabilityStore,
	occupancyStore OccupancyStore,
	accountStore AccountStore) *EligibilityGRPCApi {
	return &EligibilityGRPCApi{
		eligibilityStore:   eligibilityStore,
		suppliabilityStore: suppliabilityStore,
		occupancyStore:     occupancyStore,
		accountStore:       accountStore,
	}
}

func (a *EligibilityGRPCApi) GetAccountEligibleForSmartBooking(ctx context.Context, req *smart_booking.GetAccountEligibilityForSmartBookingRequest) (*smart_booking.GetAccountEligibilityForSmartBookingResponse, error) {
	account, err := a.accountStore.GetAccount(ctx, req.AccountId)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		logrus.Debugf("failed to get account for account ID %s: %s", req.GetAccountId(), err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account ID %s", req.AccountId)
	}

	// an account which has opted out of smart booking should not be considered eligible to go through the journey
	if account.OptOut {
		return &smart_booking.GetAccountEligibilityForSmartBookingResponse{
			AccountId: req.AccountId,
			Eligible:  false,
		}, nil
	}

	var (
		isEligibile, isSuppliable bool
		reasons                   domain.IneligibleReasons
	)

	occupancyIDs, err := a.occupancyStore.GetIDsByAccount(ctx, req.AccountId)
	if err != nil {
		logrus.Debugf("failed to get occupancies for account ID %s: %s", req.GetAccountId(), err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
	}

	for _, occupancyID := range occupancyIDs {
		eligibility, err := a.eligibilityStore.Get(ctx, occupancyID, req.AccountId)
		if err != nil {
			if errors.Is(err, store.ErrEligibilityNotFound) {
				logrus.Debugf("eligibility not computed for account %s, occupancy %s", req.AccountId, occupancyID)
				return nil, status.Errorf(codes.NotFound, "eligibility not found for account %s", req.AccountId)
			}
			logrus.Debugf("failed to get eligibility for account %s: %s", req.AccountId, err.Error())
			return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
		}

		suppliability, err := a.suppliabilityStore.Get(ctx, occupancyID, req.AccountId)
		if err != nil {
			if errors.Is(err, store.ErrSuppliabilityNotFound) {
				logrus.Debugf("suppliability not computed for account %s, occupancy %s", req.AccountId, occupancyID)
				return nil, status.Errorf(codes.NotFound, "suppliability not found for account %s", req.AccountId)
			}
			logrus.Debugf("failed to get suppliability for account %s: %s", req.AccountId, err.Error())
			return nil, status.Errorf(codes.Internal, "failed to get suppliability for account %s", req.AccountId)
		}

		if len(eligibility.Reasons) == 0 {
			isEligibile = true
		}
		if len(suppliability.Reasons) == 0 {
			isSuppliable = true
		}

		if isEligibile && isSuppliable {
			return &smart_booking.GetAccountEligibilityForSmartBookingResponse{AccountId: req.AccountId, Eligible: true}, nil
		}

		// If none of the occupancies is eligible, we want to return the reasons for the first occupancy which has active services
		if !eligibility.Reasons.Contains(domain.IneligibleReasonNoActiveService) &&
			!suppliability.Reasons.Contains(domain.IneligibleReasonNoActiveService) &&
			len(reasons) == 0 {
			reasons = append(reasons, eligibility.Reasons...)
			reasons = append(reasons, suppliability.Reasons...)
		}
	}

	protoReasons, err := reasons.MapToProto()
	if err != nil {
		logrus.Debugf("failed to get eligibility for account %s: %s", req.AccountId, err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
	}

	return &smart_booking.GetAccountEligibilityForSmartBookingResponse{
		AccountId:         req.AccountId,
		Eligible:          false,
		IneligibleReasons: protoReasons,
	}, nil
}

func (a *EligibilityGRPCApi) GetAccountOccupancyEligibilityForSmartBooking(ctx context.Context, req *smart_booking.GetAccountOccupancyEligibilityForSmartBookingRequest) (*smart_booking.GetAccountOccupancyEligibilityForSmartBookingResponse, error) {
	account, err := a.accountStore.GetAccount(ctx, req.AccountId)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		logrus.Debugf("failed to get account for account ID %s: %s", req.GetAccountId(), err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account ID %s", req.AccountId)
	}

	// an account which has opted out of smart booking should not be considered eligible to go through the journey
	if account.OptOut {
		return &smart_booking.GetAccountOccupancyEligibilityForSmartBookingResponse{
			AccountId:   req.AccountId,
			OccupancyId: req.OccupancyId,
			Eligible:    false,
		}, nil
	}

	eligibility, err := a.eligibilityStore.Get(ctx, req.OccupancyId, req.AccountId)
	if err != nil {
		if errors.Is(err, store.ErrEligibilityNotFound) {
			logrus.Debugf("eligibility not computed for account %s, occupancy %s", req.AccountId, req.OccupancyId)
			return nil, status.Errorf(codes.NotFound, "eligibility not found for account %s", req.AccountId)
		}
		logrus.Debugf("failed to get eligibility for account %s, occupancy %s: %s", req.AccountId, req.OccupancyId, err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
	}
	suppliability, err := a.suppliabilityStore.Get(ctx, req.OccupancyId, req.AccountId)
	if err != nil {
		if errors.Is(err, store.ErrEligibilityNotFound) {
			logrus.Debugf("suppliability not computed for account %s, occupancy %s", req.AccountId, req.OccupancyId)
			return nil, status.Errorf(codes.NotFound, "suppliability not found for account %s", req.AccountId)
		}
		logrus.Debugf("failed to get suppliability for account %s, occupancy %s: %s", req.AccountId, req.OccupancyId, err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get suppliability for account %s", req.AccountId)
	}

	eligible := len(eligibility.Reasons) == 0 && len(suppliability.Reasons) == 0

	return &smart_booking.GetAccountOccupancyEligibilityForSmartBookingResponse{
		AccountId:   req.AccountId,
		OccupancyId: req.OccupancyId,
		Eligible:    eligible,
	}, nil
}
