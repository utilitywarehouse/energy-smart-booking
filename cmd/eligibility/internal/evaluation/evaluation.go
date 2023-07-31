package evaluation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

func (e *Evaluator) RunFull(ctx context.Context, occupancyID string) error {
	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateCampaignability(occupancy)
	err = e.publishCampaignabilityIfChanged(ctx, occupancy, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate campaignability for occupancy %s: %w", occupancyID, err)
	}

	reasons = evaluateEligibility(occupancy)
	err = e.publishEligibilityIfChanged(ctx, occupancy, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate eligibility for occupancy %s: %w", occupancyID, err)
	}

	reasons = evaluateSuppliability(occupancy)
	err = e.publishSuppliabilityIfChanged(ctx, occupancy, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}

	return nil
}

func (e *Evaluator) RunCampaignability(ctx context.Context, occupancyID string) error {
	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateCampaignability(occupancy)
	err = e.publishCampaignabilityIfChanged(ctx, occupancy, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate campaignability for occupancy %s: %w", occupancyID, err)
	}

	return nil
}

func (e *Evaluator) RunSuppliability(ctx context.Context, occupancyID string) error {
	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateSuppliability(occupancy)
	err = e.publishSuppliabilityIfChanged(ctx, occupancy, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}

	return nil
}

func (e *Evaluator) RunEligibility(ctx context.Context, occupancyID string) error {
	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	reasons := evaluateEligibility(occupancy)
	err = e.publishEligibilityIfChanged(ctx, occupancy, reasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate eligibility for occupancy %s: %w", occupancyID, err)
	}

	return nil
}

func (e *Evaluator) publishCampaignabilityIfChanged(ctx context.Context, occupancy *domain.Occupancy, reasons domain.IneligibleReasons) error {

	if !ineligibleReasonSlicesEqual(occupancy.EvaluationResult.Campaignability, reasons) {
		if len(reasons) == 0 {
			if err := e.campaignabilitySync.Sink(ctx, &smart.CampaignableOccupancyAddedEvent{
				OccupancyId: occupancy.ID,
				AccountId:   occupancy.Account.ID,
			}, time.Now().UTC()); err != nil {
				return fmt.Errorf("failed to publish campaignability changed event for occupancy %s, account %s: %w", occupancy.ID, occupancy.Account.ID, err)
			}
			return nil
		}

		protoReasons, err := reasons.MapToProto()
		if err != nil {
			return err
		}

		if err := e.campaignabilitySync.Sink(ctx, &smart.CampaignableOccupancyRemovedEvent{
			OccupancyId: occupancy.ID,
			AccountId:   occupancy.Account.ID,
			Reasons:     protoReasons,
		}, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to publish campaignability changed event for occupancy %s, account %s: %w", occupancy.ID, occupancy.Account.ID, err)
		}
	}

	return nil
}

func (e *Evaluator) publishSuppliabilityIfChanged(ctx context.Context, occupancy *domain.Occupancy, reasons domain.IneligibleReasons) error {

	if !ineligibleReasonSlicesEqual(occupancy.EvaluationResult.Suppliability, reasons) {
		if len(reasons) == 0 {
			if err := e.suppliabilitySync.Sink(ctx, &smart.SuppliableOccupancyAddedEvent{
				OccupancyId: occupancy.ID,
				AccountId:   occupancy.Account.ID,
			}, time.Now().UTC()); err != nil {
				return fmt.Errorf("failed to publish suppliability changed event for occupancy %s, account %s: %w", occupancy.ID, occupancy.Account.ID, err)
			}
			return nil
		}

		protoReasons, err := reasons.MapToProto()
		if err != nil {
			return err
		}

		if err := e.suppliabilitySync.Sink(ctx, &smart.SuppliableOccupancyRemovedEvent{
			OccupancyId: occupancy.ID,
			AccountId:   occupancy.Account.ID,
			Reasons:     protoReasons,
		}, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to publish suppliability changed event for occupancy %s, account %s: %w", occupancy.ID, occupancy.Account.ID, err)
		}
	}

	return nil
}

func (e *Evaluator) publishEligibilityIfChanged(ctx context.Context, occupancy *domain.Occupancy, reasons domain.IneligibleReasons) error {

	if !ineligibleReasonSlicesEqual(occupancy.EvaluationResult.Eligibility, reasons) {
		if len(reasons) == 0 {
			if err := e.eligibilitySync.Sink(ctx, &smart.EligibleOccupancyAddedEvent{
				OccupancyId: occupancy.ID,
				AccountId:   occupancy.Account.ID,
			}, time.Now().UTC()); err != nil {
				return fmt.Errorf("failed to publish eligibility changed event for occupancy %s, account %s: %w", occupancy.ID, occupancy.Account.ID, err)
			}
			return nil
		}
		protoReasons, err := reasons.MapToProto()
		if err != nil {
			return err
		}

		if err := e.eligibilitySync.Sink(ctx, &smart.EligibleOccupancyRemovedEvent{
			OccupancyId: occupancy.ID,
			AccountId:   occupancy.Account.ID,
			Reasons:     protoReasons,
		}, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to publish eligibility changed event for occupancy %s, account %s: %w", occupancy.ID, occupancy.Account.ID, err)
		}
	}

	return nil
}

func ineligibleReasonSlicesEqual(x, y domain.IneligibleReasons) bool {
	less := func(a, b domain.IneligibleReason) bool { return a < b }
	return cmp.Equal(x, y, cmpopts.SortSlices(less))
}
