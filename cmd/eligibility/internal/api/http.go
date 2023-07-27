package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
)

type EvaluationResult struct {
	AccountID            string   `json:"account_id"`
	IneligibleReasons    []string `json:"ineligible_reasons"`
	SmartBookingEligible bool     `json:"smart_booking_eligible"`
}

type eligibilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Eligibility, error)
}

type suppliabilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Suppliability, error)
}

type campaignabilityStore interface {
	Get(ctx context.Context, occupancyID, accountID string) (store.Campaignability, error)
}

type occupancyStore interface {
	GetIDsByAccount(ctx context.Context, accountID string) ([]string, error)
	GetLiveOccupancies(ctx context.Context, records chan<- string) error
}

type evaluator interface {
	RunFull(ctx context.Context, occupancyID string) error
}

type Handler struct {
	eligibilityStore     eligibilityStore
	campaignabilityStore campaignabilityStore
	suppliabilityStore   suppliabilityStore
	occupancyStore       occupancyStore

	evaluator evaluator
}

func NewHandler(
	eligibilityStore eligibilityStore,
	campaignabilityStore campaignabilityStore,
	suppliabilityStore suppliabilityStore,
	occupancyStore occupancyStore,
	evaluator evaluator) *Handler {
	return &Handler{
		eligibilityStore:     eligibilityStore,
		campaignabilityStore: campaignabilityStore,
		suppliabilityStore:   suppliabilityStore,
		occupancyStore:       occupancyStore,
		evaluator:            evaluator,
	}
}

const (
	endpointAccountEvaluation = "/accounts/{ID}/evaluation"
	endpointFullEvaluation    = "/evaluation"
)

// Register registers the http handler in a http router.
func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.Handle(endpointAccountEvaluation, s.get(ctx)).Methods(http.MethodGet)
	router.Handle(endpointAccountEvaluation, s.patch(ctx)).Methods(http.MethodPatch)
	router.Handle(endpointFullEvaluation, s.runFullEvaluation(ctx)).Methods(http.MethodPatch)

}

func (s *Handler) get(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := mux.Vars(r)["ID"]
		if !ok {
			logrus.Error("accountID not provided")
			w.Write([]byte("accountID not provided"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		occupancyIDs, err := s.occupancyStore.GetIDsByAccount(ctx, accountID)
		if err != nil {
			logrus.Debugf("failed to get occupancies for account ID %s: %s", accountID, err.Error())
			w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var (
			isEligible, isSuppliable bool
			reasons                  domain.IneligibleReasons
		)
		for _, occupancyID := range occupancyIDs {
			eligibility, err := s.eligibilityStore.Get(ctx, occupancyID, accountID)
			if err != nil {
				if errors.Is(err, store.ErrEligibilityNotFound) {
					logrus.Debugf("eligibility not computed for account %s, occupancy %s", accountID, occupancyID)
					w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				logrus.Debugf("failed to get eligibility for account %s: %s", accountID, err.Error())
				w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			suppliability, err := s.suppliabilityStore.Get(ctx, occupancyID, accountID)
			if err != nil {
				if errors.Is(err, store.ErrSuppliabilityNotFound) {
					logrus.Debugf("suppliability not computed for account %s, occupancy %s", accountID, occupancyID)
					w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				logrus.Debugf("failed to get suppliability for account %s: %s", accountID, err.Error())
				w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(eligibility.Reasons) == 0 {
				isEligible = true
			}
			if len(suppliability.Reasons) == 0 {
				isSuppliable = true
			}

			if isEligible && isSuppliable {
				body, err := json.Marshal(
					EvaluationResult{
						AccountID:            accountID,
						SmartBookingEligible: true,
					})
				if err != nil {
					logrus.Errorf("failed to marshall response: %s", err.Error())
					w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(body)
				return
			}

			// If none of the occupancies is eligible, we want to return the reasons for the first occupancy which has active services
			if !eligibility.Reasons.Contains(domain.IneligibleReasonNoActiveService) &&
				!suppliability.Reasons.Contains(domain.IneligibleReasonNoActiveService) &&
				len(reasons) == 0 {
				reasons = append(reasons, eligibility.Reasons...)
				reasons = append(reasons, suppliability.Reasons...)
			}
		}

		// no eligible occupancy found and none of the occupancies had active service
		if len(reasons) == 0 {
			reasons = append(reasons, domain.IneligibleReasonNoActiveService)
		}

		body, err := json.Marshal(
			EvaluationResult{
				AccountID:            accountID,
				IneligibleReasons:    reasons.ToString(),
				SmartBookingEligible: false,
			})
		if err != nil {
			logrus.Errorf("failed to marshall response: %s", err.Error())
			w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
}

func (s *Handler) patch(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := mux.Vars(r)["ID"]
		if !ok {
			logrus.Error("accountID not provided")
			w.Write([]byte("accountID not provided"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		occupancyIDs, err := s.occupancyStore.GetIDsByAccount(ctx, accountID)
		if err != nil {
			logrus.Debugf("failed to get occupancies for account ID %s: %s", accountID, err.Error())
			w.Write([]byte(fmt.Sprintf("failed to get eligibility for account %s", accountID)))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, id := range occupancyIDs {
			logrus.Debugf("run full evaluation for account %s, occupancy %s", accountID, id)
			err := s.evaluator.RunFull(ctx, id)
			if err != nil {
				logrus.Errorf("failed to run full evaluation for account %s, occupancy %s", accountID, id)
				w.Write([]byte(fmt.Sprintf("failed to run evaluation account %s", accountID)))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})
}

func (s *Handler) runFullEvaluation(_ context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go func() {
			newContext := context.Background()
			liveOccupancies := make(chan string, 50)
			for i := 0; i < 20; i++ {
				go func() {
					for id := range liveOccupancies {
						err := s.evaluator.RunFull(newContext, id)
						if err != nil {
							logrus.Errorf("failed to run evaluation of occupancy ID %s", id)
						}
					}
				}()
			}
		}()
	})
}
