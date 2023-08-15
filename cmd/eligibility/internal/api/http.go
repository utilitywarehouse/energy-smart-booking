package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type occupancyStore interface {
	GetLiveOccupancies(ctx context.Context) ([]string, error)
}

type evaluator interface {
	RunFull(ctx context.Context, occupancyID string) error
}

type Handler struct {
	occupancyStore occupancyStore

	evaluator evaluator
}

func NewHandler(
	occupancyStore occupancyStore,
	evaluator evaluator) *Handler {
	return &Handler{
		occupancyStore: occupancyStore,
		evaluator:      evaluator,
	}
}

const (
	endpointFullEvaluation = "/evaluation"
)

// Register registers the http handler in a http router.
func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.Handle(endpointFullEvaluation, s.runFullEvaluation(ctx)).Methods(http.MethodPatch)
}

func (s *Handler) runFullEvaluation(_ context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go func() {
			start := time.Now()
			jobCtx := context.Background()
			liveOccupancies, err := s.occupancyStore.GetLiveOccupancies(jobCtx)
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
