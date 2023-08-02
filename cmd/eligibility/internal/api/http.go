package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

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
	GetPendingEvaluationOccupancies(ctx context.Context) ([]string, error)
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		occupancyIDs, err := s.occupancyStore.GetIDsByAccount(ctx, accountID)
		if err != nil {
			logrus.Debugf("failed to get occupancies for account ID %s: %s", accountID, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(occupancyIDs) == 0 {
			logrus.Debugf("No occupancy found for account ID %s", accountID)
			w.WriteHeader(http.StatusNotFound)
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
					w.WriteHeader(http.StatusNotFound)
					return
				}
				logrus.Debugf("failed to get eligibility for account %s: %s", accountID, err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			suppliability, err := s.suppliabilityStore.Get(ctx, occupancyID, accountID)
			if err != nil {
				if errors.Is(err, store.ErrSuppliabilityNotFound) {
					logrus.Debugf("suppliability not computed for account %s, occupancy %s", accountID, occupancyID)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				logrus.Debugf("failed to get suppliability for account %s: %s", accountID, err.Error())
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		occupancyIDs, err := s.occupancyStore.GetIDsByAccount(ctx, accountID)
		if err != nil {
			logrus.Debugf("failed to get occupancies for account ID %s: %s", accountID, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(occupancyIDs) == 0 {
			logrus.Errorf("no occupancies found for account ID %s", accountID)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		for _, id := range occupancyIDs {
			logrus.Debugf("run full evaluation for account %s, occupancy %s", accountID, id)
			err := s.evaluator.RunFull(ctx, id)
			if err != nil {
				logrus.Errorf("failed to run full evaluation for account %s, occupancy %s", accountID, id)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})
}

func (s *Handler) runFullEvaluation(_ context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go func() {
			start := time.Now()
			jobCtx := context.Background()
			liveOccupancies, err := s.occupancyStore.GetPendingEvaluationOccupancies(jobCtx)
			if err != nil {
				logrus.WithError(err).Error("failed to get live occupancies to evaluate")
				return
			}
			channel := make(chan string, len(liveOccupancies))
			for _, o := range liveOccupancies {
				channel <- o
			}
			close(channel)

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for id := range channel {
						now := time.Now()
						err := s.evaluator.RunFull(jobCtx, id)
						if err != nil {
							logrus.Errorf("failed to run evaluation of occupancy ID %s", id)
						} else {
							logrus.WithField("occupancy", id).WithField("elapsed", time.Since(now).String()).Debug("evaluation successfully run")
						}
					}
				}()
			}
			wg.Wait()

			logrus.WithField("elapsed", time.Since(start).String()).Info("full evaluation process completed")
		}()

		w.Write([]byte("full evaluation started"))
	})
}
