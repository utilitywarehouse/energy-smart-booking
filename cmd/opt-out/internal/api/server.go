package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/publisher"
	"github.com/utilitywarehouse/uwos-go/v1/iam"
	"github.com/utilitywarehouse/uwos-go/v1/iam/identity"
	"github.com/utilitywarehouse/uwos-go/v1/iam/pdp"
	"github.com/utilitywarehouse/uwos-go/v1/iam/principal"
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

type OptOutServer struct {
	store        AccountOptOutStore
	publisher    publisher.SyncPublisher
	accountsRepo AccountsRepository
	idClient     IDClient
}

func New(port int, store AccountOptOutStore, sink publisher.SyncPublisher, accountsRepo AccountsRepository, idClient IDClient) *http.Server {
	router := mux.NewRouter()
	handler := &OptOutServer{
		store:        store,
		publisher:    sink,
		accountsRepo: accountsRepo,
		idClient:     idClient,
	}
	handler.Register(router)

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
}

const (
	endpointAccounts = "/accounts"
	endpointAccount  = "/accounts/{number}"
)

// Register registers the http handler in a http router.
func (s *OptOutServer) Register(router *mux.Router) {
	router.Use(iam.HTTPHandler(true))

	router.HandleFunc(endpointAccounts, s.list).Methods(http.MethodGet)
	router.HandleFunc(endpointAccount, s.add).Methods(http.MethodPost)
	router.HandleFunc(endpointAccount, s.get).Methods(http.MethodGet)
	router.HandleFunc(endpointAccount, s.remove).Methods(http.MethodDelete)
}

func (s *OptOutServer) add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accountNumber, ok := mux.Vars(r)["number"]
	if !ok {
		log.Error("accountNumber not provided")
		w.Write([]byte("accountNumber not provided"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accountId, err := s.accountsRepo.AccountID(ctx, accountNumber)
	if err != nil {
		log.WithError(err).Errorf("failed to find account id for accountNumber %s", accountNumber)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// retrieve account from database to avoid sending duplicate events.
	_, err = s.store.Get(ctx, accountId)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		log.WithError(err).Errorf("failed to check opt out status for account accountNumber %s", accountNumber)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if errors.Is(err, store.ErrAccountNotFound) {

		var addedBy string

		id, err := s.idClient.WhoAmI(ctx, pdp.PrincipalFromCtx(ctx))
		if err != nil {
			log.WithError(err).Error("failed to check principal identity from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if id.Principal.Staff != nil {
			addedBy = id.Principal.Staff.Email
		}

		err = s.publisher.Sink(ctx, &smart.AccountBookingOptOutAddedEvent{
			AccountId:     accountId,
			AccountNumber: accountNumber,
			AddedBy:       addedBy,
		}, time.Now().UTC())
		if err != nil {
			log.WithError(err).Errorf("failed to publish opt out added event for account %s", accountNumber)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *OptOutServer) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountNumber, ok := mux.Vars(r)["number"]
	if !ok {
		log.Error("accountNumber not provided")
		w.Write([]byte("accountNumber not provided"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accountID, err := s.accountsRepo.AccountID(ctx, accountNumber)
	if err != nil {
		log.WithError(err).Errorf("failed to find account id for accountNumber %s", accountNumber)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// retrieve account from database to avoid sending duplicate events.
	acc, err := s.store.Get(ctx, accountID)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		log.WithError(err).Errorf("failed to check opt out status for account accountNumber %s", accountNumber)
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
}

func (s *OptOutServer) remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountNumber, ok := mux.Vars(r)["number"]
	if !ok {
		log.Error("accountNumber not provided")
		w.Write([]byte("accountNumber not provided"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accountID, err := s.accountsRepo.AccountID(ctx, accountNumber)
	if err != nil {
		log.WithError(err).Errorf("failed to find account id for accountNumber %s", accountNumber)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// retrieve account from database to avoid sending duplicate events.
	_, err = s.store.Get(ctx, accountID)
	if err != nil && !errors.Is(err, store.ErrAccountNotFound) {
		log.WithError(err).Errorf("failed to check opt out status for account accountNumber %s", accountNumber)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, store.ErrAccountNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var removedBy string

	id, err := s.idClient.WhoAmI(ctx, pdp.PrincipalFromCtx(ctx))
	if err != nil {
		log.WithError(err).Error("failed to check principal identity from context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id.Principal.Staff != nil {
		removedBy = id.Principal.Staff.Email
	}

	err = s.publisher.Sink(ctx, &smart.AccountBookingOptOutRemovedEvent{
		AccountId:     accountID,
		AccountNumber: accountNumber,
		RemovedBy:     removedBy,
	}, time.Now())
	if err != nil {
		log.WithError(err).Errorf("failed to publish opt out removed event for account %s", accountNumber)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *OptOutServer) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	list, err := s.store.List(ctx)
	if err != nil {
		log.WithError(err).Error("failed to list all accounts")
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
