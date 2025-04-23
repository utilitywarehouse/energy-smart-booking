package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	smart_booking "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/eligibility/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/evaluation"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/auth"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/helpers"
	"github.com/utilitywarehouse/uwos-go/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const resourceID = "booking-eligibility-grpc-api"

var (
	ErrUserUnauthorised = errors.New("user does not have required access")
)

type EligibilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Eligibility, error)
}

type Auth interface {
	Authorize(ctx context.Context, params *auth.PolicyParams) (bool, error)
}

type SuppliabilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Suppliability, error)
}

type OccupancyStore interface {
	GetLiveOccupanciesIDsByAccountID(ctx context.Context, accountID string) ([]string, error)
}

type AccountStore interface {
	GetAccount(ctx context.Context, accountID string) (store.Account, error)
}

type ServiceStore interface {
	GetLiveServicesWithBookingRef(ctx context.Context, occupancyID string) ([]store.ServiceBookingRef, error)
}

type EligibilityGRPCApi struct {
	smart_booking.UnimplementedEligiblityAPIServer
	eligibilityStore    EligibilityStore
	suppliabilityStore  SuppliabilityStore
	occupancyStore      OccupancyStore
	accountStore        AccountStore
	serviceStore        ServiceStore
	auth                Auth
	meterpointEvaluator *evaluation.MeterpointEvaluator
}

func NewEligibilityGRPCApi(
	eligibilityStore EligibilityStore,
	suppliabilityStore SuppliabilityStore,
	occupancyStore OccupancyStore,
	accountStore AccountStore,
	serviceStore ServiceStore,
	auth Auth,
	meterpointEvaluator *evaluation.MeterpointEvaluator,
) *EligibilityGRPCApi {
	return &EligibilityGRPCApi{
		eligibilityStore:    eligibilityStore,
		suppliabilityStore:  suppliabilityStore,
		occupancyStore:      occupancyStore,
		accountStore:        accountStore,
		serviceStore:        serviceStore,
		auth:                auth,
		meterpointEvaluator: meterpointEvaluator,
	}
}

func (a *EligibilityGRPCApi) GetAccountEligibleForSmartBooking(ctx context.Context, req *smart_booking.GetAccountEligibleForSmartBookingRequest) (_ *smart_booking.GetAccountEligibleForSmartBookingResponse, err error) {
	ctx, span := tracing.Start(ctx, "EligibilityAPI.GetAccountEligibleForSmartBooking",
		trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	err = a.validateCredentials(ctx, auth.GetAction, auth.EligibilityResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	account, err := a.accountStore.GetAccount(ctx, req.AccountId)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		logrus.Debugf("failed to get account for account ID %s: %s", req.GetAccountId(), err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account ID %s", req.AccountId)
	}

	span.AddEvent("get-account", trace.WithAttributes(attribute.Bool("opt.out", account.OptOut), attribute.String("psr.codes", fmt.Sprintf("%v", account.PSRCodes))))

	// an account which has opted out of smart booking should not be considered eligible to go through the journey
	if account.OptOut {
		return &smart_booking.GetAccountEligibleForSmartBookingResponse{
			AccountId: req.AccountId,
			Eligible:  false,
		}, nil
	}

	var (
		eligible bool
		reasons  domain.IneligibleReasons
	)

	occupancyIDs, err := a.occupancyStore.GetLiveOccupanciesIDsByAccountID(ctx, req.AccountId)
	if err != nil {
		logrus.Debugf("failed to get live occupancies for account ID %s: %s", req.GetAccountId(), err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
	}
	span.AddEvent("get-occupancy", trace.WithAttributes(attribute.String("ids", fmt.Sprintf("%v", occupancyIDs))))

	// if there are no live occupancies for given account
	// customer is not eligible to go through smart booking journey
	if len(occupancyIDs) == 0 {
		reasons = domain.IneligibleReasons{domain.IneligibleReasonNoActiveService}
	} else {
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

			eligible = len(eligibility.Reasons) == 0 && len(suppliability.Reasons) == 0
			if eligible {
				// check it has booking references assigned
				serviceBookingRef, err := a.serviceStore.GetLiveServicesWithBookingRef(ctx, occupancyID)
				if err != nil {
					logrus.Debugf("failed to get service booking references for account %s, occupancy %s: %s", req.AccountId, occupancyID, err.Error())
					return nil, status.Errorf(codes.Internal, "failed to check service booking references for account %s", req.AccountId)
				}
				serviceBookingRefAttr := helpers.CreateSpanAttribute(serviceBookingRef, "get-live-services", span)
				span.AddEvent("service-booking-references", trace.WithAttributes(serviceBookingRefAttr))

				hasBookingRef := len(serviceBookingRef) > 0
				for _, s := range serviceBookingRef {
					if s.BookingRef == "" || s.DeletedAt != nil {
						hasBookingRef = false
						break
					}
				}
				if hasBookingRef {
					span.AddEvent("get-eligibility", trace.WithAttributes(
						attribute.String("reasons", fmt.Sprintf("%v", reasons)),
						attribute.Bool("eligible", true)))
					return &smart_booking.GetAccountEligibleForSmartBookingResponse{AccountId: req.AccountId, Eligible: true}, nil
				}
			}

			// If none of the occupancies is eligible, we want to return the reasons for the first occupancy
			if len(reasons) == 0 {
				reasons = append(reasons, eligibility.Reasons...)
				reasons = append(reasons, suppliability.Reasons...)
			}
		}
	}

	span.AddEvent("get-eligibility", trace.WithAttributes(
		attribute.String("reasons", fmt.Sprintf("%v", reasons)),
		attribute.Bool("eligible", false)))

	protoReasons, err := reasons.MapToProto()
	if err != nil {
		logrus.Debugf("failed to get eligibility for account %s: %s", req.AccountId, err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account %s", req.AccountId)
	}

	return &smart_booking.GetAccountEligibleForSmartBookingResponse{
		AccountId:         req.AccountId,
		Eligible:          false,
		IneligibleReasons: protoReasons,
	}, nil
}

func (a *EligibilityGRPCApi) GetAccountOccupancyEligibleForSmartBooking(ctx context.Context, req *smart_booking.GetAccountOccupancyEligibleForSmartBookingRequest) (_ *smart_booking.GetAccountOccupancyEligibleForSmartBookingResponse, err error) {
	ctx, span := tracing.Start(ctx, "EligibilityAPI.GetAccountOccupancyEligibleForSmartBooking",
		trace.WithAttributes(attribute.String("account.id", req.GetAccountId())),
		trace.WithAttributes(attribute.String("occupancy.id", req.GetOccupancyId())),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	err = a.validateCredentials(ctx, auth.GetAction, auth.EligibilityResource, req.AccountId)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	account, err := a.accountStore.GetAccount(ctx, req.AccountId)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		logrus.Debugf("failed to get account for account ID %s: %s", req.GetAccountId(), err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get eligibility for account ID %s", req.AccountId)
	}

	span.AddEvent("get-account", trace.WithAttributes(attribute.Bool("opt.out", account.OptOut), attribute.String("psr.codes", fmt.Sprintf("%v", account.PSRCodes))))

	// an account which has opted out of smart booking should not be considered eligible to go through the journey
	if account.OptOut {
		return &smart_booking.GetAccountOccupancyEligibleForSmartBookingResponse{
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
	if eligible {
		// check it has booking references assigned
		serviceBookingRef, err := a.serviceStore.GetLiveServicesWithBookingRef(ctx, req.OccupancyId)
		if err != nil {
			logrus.Debugf("failed to get service booking references for account %s, occupancy %s: %s", req.AccountId, req.OccupancyId, err.Error())
			return nil, status.Errorf(codes.Internal, "failed to check service booking references for account %s", req.AccountId)
		}
		serviceBookingRefAttr := helpers.CreateSpanAttribute(serviceBookingRef, "get-live-services", span)
		span.AddEvent("service-booking-references", trace.WithAttributes(serviceBookingRefAttr))

		for _, s := range serviceBookingRef {
			if s.BookingRef == "" || s.DeletedAt != nil {
				eligible = false
				break
			}
		}
	}

	span.AddEvent("get-eligibility", trace.WithAttributes(
		attribute.String("eligibility.reasons", fmt.Sprintf("%v", eligibility.Reasons)),
		attribute.String("suppliability.reasons", fmt.Sprintf("%v", suppliability.Reasons)),
		attribute.Bool("eligible", eligible)))

	return &smart_booking.GetAccountOccupancyEligibleForSmartBookingResponse{
		AccountId:   req.AccountId,
		OccupancyId: req.OccupancyId,
		Eligible:    eligible,
	}, nil
}

func (a *EligibilityGRPCApi) GetMeterpointEligibility(ctx context.Context, req *smart_booking.GetMeterpointEligibilityRequest) (_ *smart_booking.GetMeterpointEligibilityResponse, err error) {
	ctx, span := tracing.Start(ctx, "EligibilityAPI.GetMeterpointEligibility",
		trace.WithAttributes(attribute.String("electricity.meterpoint.number", req.GetMpan())),
		trace.WithAttributes(attribute.String("gas.meterpoint.number", req.GetMprn())),
		trace.WithAttributes(attribute.String("customer.postcode", req.GetPostcode())))
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	if req.GetMpan() == "" {
		return nil, status.Error(codes.InvalidArgument, "no mpan provided")
	}
	if req.GetPostcode() == "" {
		return nil, status.Error(codes.InvalidArgument, "no postcode provided")
	}

	err = a.validateCredentials(ctx, auth.GetAction, auth.EligibilityResource, resourceID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserUnauthorised):
			return nil, status.Errorf(codes.PermissionDenied, "user does not have access to this action, %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate credentials")
		}
	}

	if req.GetMprn() != "" {
		result, err := a.meterpointEvaluator.GetGasMeterpointEligibility(ctx, req.GetMprn())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !result.Eligible {
			span.AddEvent("response", trace.WithAttributes(attribute.Bool("eligible", result.Eligible), attribute.String("reason", string(result.Reason))))
			return &smart_booking.GetMeterpointEligibilityResponse{
				Eligible: false,
			}, nil
		}
	}

	result, err := a.meterpointEvaluator.GetElectricityMeterpointEligibility(ctx, req.Mpan, req.GetPostcode())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	span.AddEvent("response", trace.WithAttributes(attribute.Bool("eligible", result.Eligible), attribute.String("reason", string(result.Reason))))

	return &smart_booking.GetMeterpointEligibilityResponse{
		Eligible: result.Eligible,
	}, nil
}

func (a *EligibilityGRPCApi) validateCredentials(ctx context.Context, action, resource, requestAccountID string) error {

	authorised, err := a.auth.Authorize(ctx, &auth.PolicyParams{
		Action:     action,
		Resource:   resource,
		ResourceID: requestAccountID,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"action":   action,
			"resource": resource,
		}).Error("Authorize error: ", err)
		return err
	}
	if !authorised {
		return ErrUserUnauthorised
	}

	return nil
}
