package api

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	smart_booking "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
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

type Evaluator interface {
	RunSuppliability(ctx context.Context, occupancyID string) error
	RunEligibility(ctx context.Context, occupancyID string) error
}

type OccupancyStore interface {
	GetIDsByAccount(ctx context.Context, accountID string) ([]string, error)
}

type EligibilityGRPCApi struct {
	smart_booking.UnimplementedEligiblityAPIServer
	eligibilityStore   EligibilityStore
	suppliabilityStore SuppliabilityStore
	occupancyStore     OccupancyStore
}

func NewEligibilityGRPCApi(eligibilityStore EligibilityStore, suppliabilityStore SuppliabilityStore, occupancyStore OccupancyStore) *EligibilityGRPCApi {
	return &EligibilityGRPCApi{
		eligibilityStore:   eligibilityStore,
		suppliabilityStore: suppliabilityStore,
		occupancyStore:     occupancyStore,
	}
}

func (a *EligibilityGRPCApi) GetAccountEligibleForSmartBooking(ctx context.Context, req *smart_booking.GetAccountEligibilityForSmartBookingRequest) (*smart_booking.GetAccountEligibilityForSmartBookingResponse, error) {
	var isEligibile, isSuppliable bool

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
				return nil, status.Errorf(codes.NotFound, "eligibility not for account %s", req.AccountId)
			}
			logrus.Debugf("failed to get eligibility for account %s: %s", req.AccountId, err.Error())
			return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
		}

		suppliability, err := a.suppliabilityStore.Get(ctx, occupancyID, req.AccountId)
		if err != nil {
			if errors.Is(err, store.ErrSuppliabilityNotFound) {
				logrus.Debugf("suppliability not computed for account %s, occupancy %s", req.AccountId, occupancyID)
				return nil, status.Errorf(codes.NotFound, "suppliability not for account %s", req.AccountId)
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
	}

	return &smart_booking.GetAccountEligibilityForSmartBookingResponse{
		AccountId: req.AccountId,
		Eligible:  false,
	}, nil
}
