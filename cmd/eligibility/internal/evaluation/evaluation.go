package evaluation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/metrics"
)

func (e *Evaluator) RunFull(ctx context.Context, occupancyID string) error {
	occupancy, err := e.LoadOccupancy(ctx, occupancyID)
	if err != nil {
		return fmt.Errorf("failed to load occupanncy for ID %s: %w", occupancyID, err)
	}

	cReasons := evaluateCampaignability(occupancy)
	err = e.publishCampaignabilityIfChanged(ctx, occupancy, cReasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate campaignability for occupancy %s: %w", occupancyID, err)
	}

	eReasons := evaluateEligibility(occupancy)
	err = e.publishEligibilityIfChanged(ctx, occupancy, eReasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate eligibility for occupancy %s: %w", occupancyID, err)
	}

	sReasons := evaluateSuppliability(occupancy)
	err = e.publishSuppliabilityIfChanged(ctx, occupancy, sReasons)
	if err != nil {
		return fmt.Errorf("failed to evaluate suppliability for occupancy %s: %w", occupancyID, err)
	}

	metrics.SmartBookingEvaluationCounter.WithLabelValues("full").Inc()

	occupancy.EvaluationResult.CampaignabilityEvaluated = true
	occupancy.EvaluationResult.Campaignability = cReasons
	occupancy.EvaluationResult.SuppliabilityEvaluated = true
	occupancy.EvaluationResult.Suppliability = sReasons
	occupancy.EvaluationResult.EligibilityEvaluated = true
	occupancy.EvaluationResult.Eligibility = eReasons

	err = e.publishSmartBookingJourneyEligibilityIfNeeded(ctx, occupancy)

	return err
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

	metrics.SmartBookingEvaluationCounter.WithLabelValues("campaignability").Inc()

	occupancy.EvaluationResult.CampaignabilityEvaluated = true
	occupancy.EvaluationResult.Campaignability = reasons

	err = e.publishSmartBookingJourneyEligibilityIfNeeded(ctx, occupancy)

	return err
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

	metrics.SmartBookingEvaluationCounter.WithLabelValues("suppliability").Inc()

	occupancy.EvaluationResult.SuppliabilityEvaluated = true
	occupancy.EvaluationResult.Suppliability = reasons

	err = e.publishSmartBookingJourneyEligibilityIfNeeded(ctx, occupancy)

	return err
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

	metrics.SmartBookingEvaluationCounter.WithLabelValues("eligibility").Inc()

	occupancy.EvaluationResult.EligibilityEvaluated = true
	occupancy.EvaluationResult.Eligibility = reasons

	err = e.publishSmartBookingJourneyEligibilityIfNeeded(ctx, occupancy)

	return err
}

func (e *Evaluator) publishCampaignabilityIfChanged(ctx context.Context, occupancy *domain.Occupancy, reasons domain.IneligibleReasons) error {

	if !occupancy.EvaluationResult.CampaignabilityEvaluated || !ineligibleReasonSlicesEqual(occupancy.EvaluationResult.Campaignability, reasons) {
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

	if !occupancy.EvaluationResult.SuppliabilityEvaluated || !ineligibleReasonSlicesEqual(occupancy.EvaluationResult.Suppliability, reasons) {
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

	if !occupancy.EvaluationResult.EligibilityEvaluated || !ineligibleReasonSlicesEqual(occupancy.EvaluationResult.Eligibility, reasons) {
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

func (e *Evaluator) publishSmartBookingJourneyEligibilityIfNeeded(ctx context.Context, occupancy *domain.Occupancy) error {

	if !(occupancy.EvaluationResult.CampaignabilityEvaluated &&
		occupancy.EvaluationResult.SuppliabilityEvaluated &&
		occupancy.EvaluationResult.EligibilityEvaluated) {
		return nil
	}

	serviceBookingRef, err := e.serviceStore.GetLiveServicesWithBookingRef(ctx, occupancy.ID)
	if err != nil {
		return fmt.Errorf("failed to check booking ref for services of occupancy ID %s: %w", occupancy.ID, err)
	}

	hasBookingRef := len(serviceBookingRef) > 0
	for _, s := range serviceBookingRef {
		if s.BookingRef == "" || s.DeletedAt != nil {
			hasBookingRef = false
			break
		}
	}

	eligible := len(occupancy.EvaluationResult.Campaignability) == 0 &&
		len(occupancy.EvaluationResult.Suppliability) == 0 &&
		len(occupancy.EvaluationResult.Eligibility) == 0

	if eligible && hasBookingRef {
		err = e.bookingEligibilitySync.Sink(ctx, &smart.SmartBookingJourneyOccupancyAddedEvent{
			OccupancyId: occupancy.ID,
			Reference:   serviceBookingRef[0].BookingRef,
		}, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to publish smart booking journey eligibility added event for occupancy ID %s: %w", occupancy.ID, err)
		}
	} else {
		err = e.bookingEligibilitySync.Sink(ctx, &smart.SmartBookingJourneyOccupancyRemovedEvent{
			OccupancyId: occupancy.ID,
		}, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to publish smart booking journey eligibility removed event for occupancy ID %s: %w", occupancy.ID, err)
		}
	}

	return nil
}

func ineligibleReasonSlicesEqual(x, y domain.IneligibleReasons) bool {
	less := func(a, b domain.IneligibleReason) bool { return a < b }
	return cmp.Equal(x, y, cmpopts.SortSlices(less))
}
