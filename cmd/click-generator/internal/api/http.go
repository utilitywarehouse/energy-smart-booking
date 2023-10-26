package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ClickLinkGateway interface {
	GenerateGenericLink(ctx context.Context, accountNo string) (string, error)
	GenerateAuthenticated(ctx context.Context, accountNo string) (string, error)
}

type Handler struct {
	clickLinkGw ClickLinkGateway
}

func NewHandler(clickLinkGw ClickLinkGateway) *Handler {
	return &Handler{clickLinkGw}
}

const (
	endpointGenerate = "/generate"
)

type GenerateLinkRequest struct {
	AccountNumber string `json:"account_number"`
}

func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.Handle(endpointGenerate, s.generate(ctx)).Methods(http.MethodPost)
}

func (s *Handler) generate(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GenerateLinkRequest
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logrus.Errorf("Failed to read body content: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(b, &req)
		if err != nil {
			logrus.Errorf("Failed to unmarshall body content: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.AccountNumber == "" {
			logrus.Error("account number not provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		linkType := r.URL.Query().Get("type")
		if linkType == "" {
			logrus.Error("link type not specified")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var link string

		switch linkType {
		default:
			logrus.Errorf("unknown link type requested: %s", linkType)
			w.WriteHeader(http.StatusBadRequest)
			return
		case "auth":
			link, err = s.clickLinkGw.GenerateAuthenticated(ctx, req.AccountNumber)
		case "generic":
			link, err = s.clickLinkGw.GenerateGenericLink(ctx, req.AccountNumber)
		}

		if err != nil {
			logrus.WithField("account_number", req.AccountNumber).Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(link))
	})
}
