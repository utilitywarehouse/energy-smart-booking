package api

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type occupancyStore interface {
	GetLiveOccupanciesPendingEvaluation(ctx context.Context) ([]string, error)
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
	endpointFullEvaluation  = "/evaluation"
	endpointRerunEvaluation = "/rerunEvaluation"
)

// Register registers the http handler in a http router.
func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.Handle(endpointFullEvaluation, s.runFullEvaluation(ctx)).Methods(http.MethodPatch)
	router.Handle(endpointRerunEvaluation, s.rerunFullEvaluation(ctx)).Methods(http.MethodPatch)
}

func (s *Handler) runFullEvaluation(_ context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		go func() {
			start := time.Now()
			jobCtx := context.Background()
			occupancies, err := s.occupancyStore.GetLiveOccupanciesPendingEvaluation(jobCtx)
			if err != nil {
				slog.Error("failed to get live occupancies to evaluate", "error", err)
				return
			}
			s.runEvaluation(jobCtx, occupancies)
			slog.Info("full evaluation process completed", "elapsed", time.Since(start).String())
		}()

		w.Write([]byte("full evaluation started"))
	})
}

func (s *Handler) rerunFullEvaluation(_ context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		go func() {
			start := time.Now()
			jobCtx := context.Background()
			occupancies, err := s.occupancyStore.GetLiveOccupancies(jobCtx)
			if err != nil {
				slog.Error("failed to get live occupancies to evaluate", "error", err)
				return
			}
			s.runEvaluation(jobCtx, occupancies)
			slog.Info("full evaluation process completed", "elapsed", time.Since(start).String())
		}()

		w.Write([]byte("full evaluation started"))
	})
}

func (s *Handler) runEvaluation(ctx context.Context, ids []string) {
	channel := make(chan string, len(ids))
	for _, o := range ids {
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
				err := s.evaluator.RunFull(ctx, id)
				if err != nil {
					slog.Error("failed to run evaluation", "occupancy_id", id)
				} else {
					slog.Debug("evaluation successfully ran", "occupancy_id", id, "elapsed", time.Since(now).String())
				}
			}
		}()
	}
	wg.Wait()
}
