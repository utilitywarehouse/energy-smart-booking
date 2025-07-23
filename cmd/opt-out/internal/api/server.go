package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/uwos-go/iam/identity"
	"github.com/utilitywarehouse/uwos-go/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/iam/principal"
)

type Account struct {
	ID      string    `json:"id"`
	Number  string    `json:"number"`
	AddedBy string    `json:"added_by"`
	AddedAt time.Time `json:"added_at"`
}

type AccountOptOutStore interface {
	Get(ctx context.Context, number string) (*store.Account, error)
	List(ctx context.Context) ([]store.Account, error)
}

type AccountsRepository interface {
	AccountID(ctx context.Context, accountNumber string) (string, error)
}

type IDClient interface {
	WhoAmI(ctx context.Context, in *principal.Model) (identity.WhoAmIResult, error)
}

type Handler struct {
	store        AccountOptOutStore
	publisher    publisher.SyncPublisher
	accountsRepo AccountsRepository
	idClient     IDClient
}

func NewHandler(store AccountOptOutStore, sink publisher.SyncPublisher, accountsRepo AccountsRepository, idClient IDClient) *Handler {
	return &Handler{
		store:        store,
		publisher:    sink,
		accountsRepo: accountsRepo,
		idClient:     idClient,
	}
}

const (
	endpointAccounts = "/accounts"
	endpointAccount  = "/accounts/{number}"
)

// Register registers the http handler in a http router.
func (s *Handler) Register(ctx context.Context, router *mux.Router) {
	router.HandleFunc(endpointAccounts, s.list).Methods(http.MethodGet)
	router.Handle(endpointAccount, s.add(ctx)).Methods(http.MethodPost)
	router.Handle(endpointAccount, s.get(ctx)).Methods(http.MethodGet)
	router.Handle(endpointAccount, s.remove(ctx)).Methods(http.MethodDelete)
}

// EnableCORS enables adding CORS headers.
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("enable cors middlware")
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Handler) add(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountNumber, ok := mux.Vars(r)["number"]
		if !ok {
			slog.Error("accountNumber not provided")
			w.Write([]byte("accountNumber not provided"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		accountID, err := s.accountsRepo.AccountID(ctx, accountNumber)
		if err != nil {
			slog.Error("failed to find account id for accountNumber", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// retrieve account from database to avoid sending duplicate events.
		_, err = s.store.Get(ctx, accountID)
		if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
			slog.Error("failed to check opt out status for account accountNumber", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if errors.Is(err, store.ErrAccountNotFound) {

			var addedBy string

			id, err := s.idClient.WhoAmI(ctx, pdp.PrincipalFromCtx(r.Context()))
			if err != nil {
				slog.Error("failed to check principal identity from context", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if id.Principal.Staff != nil {
				addedBy = id.Principal.Staff.Email
			}

			err = s.publisher.Sink(ctx, &smart.AccountBookingOptOutAddedEvent{
				AccountId: accountID,
				AddedBy:   addedBy,
			}, time.Now().UTC())
			if err != nil {
				slog.Error("failed to publish opt out added event for account", "account_number", accountNumber, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func (s *Handler) get(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountNumber, ok := mux.Vars(r)["number"]
		if !ok {
			slog.Error("accountNumber not provided")
			w.Write([]byte("accountNumber not provided"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		accountID, err := s.accountsRepo.AccountID(ctx, accountNumber)
		if err != nil {
			slog.Error("failed to find account id for accountNumber", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// retrieve account from database to avoid sending duplicate events.
		acc, err := s.store.Get(ctx, accountID)
		if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
			slog.Error("failed to check opt out status for accountNumber", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		account := &Account{
			ID:      acc.ID,
			Number:  acc.Number,
			AddedBy: acc.AddedBy,
			AddedAt: acc.AddedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		j, _ := json.Marshal(account)
		_, _ = w.Write(j)
	})
}

func (s *Handler) remove(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountNumber, ok := mux.Vars(r)["number"]
		if !ok {
			slog.Error("accountNumber not provided")
			w.Write([]byte("accountNumber not provided"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		accountID, err := s.accountsRepo.AccountID(ctx, accountNumber)
		if err != nil {
			slog.Error("failed to find account id for accountNumber", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// retrieve account from database to avoid sending duplicate events.
		_, err = s.store.Get(ctx, accountID)
		if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
			slog.Error("failed to check opt out status for accountNumber", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if errors.Is(err, store.ErrAccountNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var removedBy string

		id, err := s.idClient.WhoAmI(r.Context(), pdp.PrincipalFromCtx(r.Context()))
		if err != nil {
			slog.Error("failed to check principal identity from context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if id.Principal.Staff != nil {
			removedBy = id.Principal.Staff.Email
		}

		err = s.publisher.Sink(ctx, &smart.AccountBookingOptOutRemovedEvent{
			AccountId: accountID,
			RemovedBy: removedBy,
		}, time.Now())
		if err != nil {
			slog.Error("failed to publish opt out removed event for account", "account_number", accountNumber, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func (s *Handler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	list, err := s.store.List(ctx)
	if err != nil {
		slog.Error("failed to list all accounts", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	accounts := make([]Account, len(list))
	for i, a := range list {
		accounts[i] = Account{
			ID:      a.ID,
			Number:  a.Number,
			AddedBy: a.AddedBy,
			AddedAt: a.AddedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	j, _ := json.Marshal(accounts)
	_, _ = w.Write(j)
}
