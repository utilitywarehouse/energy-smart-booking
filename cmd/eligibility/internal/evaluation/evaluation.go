package evaluation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
)

func (e *Evaluator) RunFull(ctx context.Context, occupancyID string) error {
	now := time.Now()

	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateCampaignability(occupancy)
	err = e.publishCampaignabilityIfChanged(ctx, occupancyID, occupancy.Account.ID, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate campaignability for occupancy %s: %w", occupancyID, err)
	}

	reasons = evaluateEligibility(occupancy)
	err = e.publishEligibilityIfChanged(ctx, occupancyID, occupancy.Account.ID, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate eligibility for occupancy %s: %w", occupancyID, err)
	}

	reasons, err = evaluateSuppliability(occupancy)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}
	err = e.publishSuppliabilityIfChanged(ctx, occupancyID, occupancy.Account.ID, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}

	logrus.Infof("time to full evaluate occupancy %s: %s", occupancyID, time.Since(now))

	return nil
}

func (e *Evaluator) RunCampaignability(ctx context.Context, occupancyID string) error {
	now := time.Now()

	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateCampaignability(occupancy)
	err = e.publishCampaignabilityIfChanged(ctx, occupancyID, occupancy.Account.ID, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate campaignability for occupancy %s: %w", occupancyID, err)
	}

	logrus.Infof("time to run campaignability for occupancy %s: %s", occupancyID, time.Since(now))

	return nil
}

func (e *Evaluator) RunSuppliability(ctx context.Context, occupancyID string) error {
	now := time.Now()

	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons, err := evaluateSuppliability(occupancy)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}
	err = e.publishSuppliabilityIfChanged(ctx, occupancyID, occupancy.Account.ID, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}

	logrus.Infof("time to run suppliability for occupancy %s: %s", occupancyID, time.Since(now))

	return nil
}

func (e *Evaluator) RunEligibility(ctx context.Context, occupancyID string) error {
	now := time.Now()

	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateEligibility(occupancy)
	err = e.publishEligibilityIfChanged(ctx, occupancyID, occupancy.Account.ID, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate eligibility for occupancy %s: %w", occupancyID, err)
	}

	logrus.Infof("time to run eligibility for occupancy %s: %s", occupancyID, time.Since(now))

	return nil
}

func (e *Evaluator) publishCampaignabilityIfChanged(ctx context.Context, occupancyID string, accountID string, reasons domain.IneligibleReasons) error {
	campaignability, err := e.campaignabilityStore.Get(ctx, occupancyID, accountID)
	if err != nil && !errors.Is(err, store.ErrCampaignabilityNotFound) {
		return fmt.Errorf("failed to check campaignability for occupancyID %s, accountID %s: %w", occupancyID, accountID, err)
	}

	if errors.Is(err, store.ErrCampaignabilityNotFound) ||
		!ineligibleReasonSlicesEqual(campaignability.Reasons, reasons) {
		if len(reasons) == 0 {
			if err := e.campaignabilitySync.Sink(ctx, &smart.CampaignableOccupancyAddedEvent{
				OccupancyId: occupancyID,
				AccountId:   accountID,
			}, time.Now().UTC()); err != nil {
				return fmt.Errorf("failed to publish campaignability changed event for occupancy %s, account %s: %w", occupancyID, accountID, err)
			}
			return nil
		}

		protoReasons, err := reasons.MapToProto()
		if err != nil {
			return err
		}

		if err := e.campaignabilitySync.Sink(ctx, &smart.CampaignableOccupancyRemovedEvent{
			OccupancyId: occupancyID,
			AccountId:   accountID,
			Reasons:     protoReasons,
		}, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to publish campaignability changed event for occupancy %s, account %s: %w", occupancyID, accountID, err)
		}
	}

	return nil
}

func (e *Evaluator) publishSuppliabilityIfChanged(ctx context.Context, occupancyID string, accountID string, reasons domain.IneligibleReasons) error {
	suppliability, err := e.suppliabilityStore.Get(ctx, occupancyID, accountID)
	if err != nil && !errors.Is(err, store.ErrSuppliabilityNotFound) {
		return fmt.Errorf("failed to check suppliability for occupancyID %s, accountID %s: %w", occupancyID, accountID, err)
	}

	if errors.Is(err, store.ErrSuppliabilityNotFound) ||
		!ineligibleReasonSlicesEqual(suppliability.Reasons, reasons) {
		if len(reasons) == 0 {
			if err := e.suppliabilitySync.Sink(ctx, &smart.SuppliableOccupancyAddedEvent{
				OccupancyId: occupancyID,
				AccountId:   accountID,
			}, time.Now().UTC()); err != nil {
				return fmt.Errorf("failed to publish suppliability changed event for occupancy %s, account %s: %w", occupancyID, accountID, err)
			}
			return nil
		}

		protoReasons, err := reasons.MapToProto()
		if err != nil {
			return err
		}

		if err := e.suppliabilitySync.Sink(ctx, &smart.SuppliableOccupancyRemovedEvent{
			OccupancyId: occupancyID,
			AccountId:   accountID,
			Reasons:     protoReasons,
		}, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to publish suppliability changed event for occupancy %s, account %s: %w", occupancyID, accountID, err)
		}
	}

	return nil
}

func (e *Evaluator) publishEligibilityIfChanged(ctx context.Context, occupancyID string, accountID string, reasons domain.IneligibleReasons) error {
	var err error

	eligibility, err := e.eligibilityStore.Get(ctx, occupancyID, accountID)
	if err != nil && !errors.Is(err, store.ErrEligibilityNotFound) {
		return fmt.Errorf("failed to check eligibility for occupancyID %s, accountID %s: %w", occupancyID, accountID, err)
	}

	if errors.Is(err, store.ErrEligibilityNotFound) ||
		!ineligibleReasonSlicesEqual(eligibility.Reasons, reasons) {
		if len(reasons) == 0 {
			if err := e.eligibilitySync.Sink(ctx, &smart.EligibleOccupancyAddedEvent{
				OccupancyId: occupancyID,
				AccountId:   accountID,
			}, time.Now().UTC()); err != nil {
				return fmt.Errorf("failed to publish eligibility changed event for occupancy %s, account %s: %w", occupancyID, accountID, err)
			}
			return nil
		}
		protoReasons, err := reasons.MapToProto()
		if err != nil {
			return err
		}

		if err := e.eligibilitySync.Sink(ctx, &smart.EligibleOccupancyRemovedEvent{
			OccupancyId: occupancyID,
			AccountId:   accountID,
			Reasons:     protoReasons,
		}, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to publish eligibility changed event for occupancy %s, account %s: %w", occupancyID, accountID, err)
		}
	}

	return nil
}

func ineligibleReasonSlicesEqual(x, y domain.IneligibleReasons) bool {
	less := func(a, b domain.IneligibleReason) bool { return a < b }
	return cmp.Equal(x, y, cmpopts.SortSlices(less))
}
