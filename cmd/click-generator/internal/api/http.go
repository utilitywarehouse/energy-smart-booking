package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type ClickLinkGateway interface {
	GenerateGenericLink(ctx context.Context, accountNo string) (string, error)
	GenerateAuthenticated(ctx context.Context, accountNo string, attributes map[string]string) (string, error)
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
	AccountNumber string            `json:"account_number"`
	Attributes    map[string]string `json:"attributes"`
}

func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.Handle(endpointGenerate, s.generate(ctx)).Methods(http.MethodPost)
}

func (s *Handler) generate(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GenerateLinkRequest
		b, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("failed to read body content", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(b, &req)
		if err != nil {
			slog.Error("failed to unmarshall body content", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.AccountNumber == "" {
			slog.Error("account number not provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		linkType := r.URL.Query().Get("type")
		if linkType == "" {
			slog.Error("link type not specified")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var link string
		switch linkType {
		default:
			slog.Error("unknown link type requested", "link_type", linkType)
			w.WriteHeader(http.StatusBadRequest)
			return
		case "auth":
			link, err = s.clickLinkGw.GenerateAuthenticated(ctx, req.AccountNumber, req.Attributes)
		case "generic":
			link, err = s.clickLinkGw.GenerateGenericLink(ctx, req.AccountNumber)
		}

		if err != nil {
			slog.Error(err.Error(), "account_number", req.AccountNumber)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(link))
	})
}
