package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/generator"
)

type Handler struct {
	generator *generator.LinkProvider
}

func NewHandler(generator *generator.LinkProvider) *Handler {
	return &Handler{generator: generator}
}

const (
	endpointGenerate = "/accounts/{number}/generate"
)

func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.Handle(endpointGenerate, s.patch(ctx)).Methods(http.MethodPatch)
}

func (s *Handler) patch(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountNumber, ok := mux.Vars(r)["number"]
		if !ok {
			logrus.Error("account number not provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		link, err := s.generator.Generate(ctx, accountNumber)
		if err != nil {
			logrus.WithField("account_number", accountNumber).Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(link))
	})
}
