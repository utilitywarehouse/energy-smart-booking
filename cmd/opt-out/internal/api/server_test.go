package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store/migrations"
	"github.com/utilitywarehouse/uwos-go/iam/identity"
	"github.com/utilitywarehouse/uwos-go/iam/principal"
	"google.golang.org/protobuf/proto"
)

func TestServer(t *testing.T) {
	testAccountNumber := "accountNo"
	testAccountID := "accountID"

	ctx := context.Background()
	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)

	postgresURL, err := postgres.GetTestContainerDSN(container)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := postgres.Setup(ctx, postgresURL, migrations.Source)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = postgres.Teardown(pool, migrations.Source); err != nil {
			t.Fatal(err)
		}
	}()

	s := store.NewAccountOptOut(pool)
	mockPublisher := MockSink{}
	mockAccountsRepo := accountRepoMock{
		accountNumberID: map[string]string{
			testAccountNumber: testAccountID,
		},
	}
	identityClient := identityClientMock{}
	router := mux.NewRouter()
	httpHandler := NewHandler(s, &mockPublisher, &mockAccountsRepo, &identityClient)
	httpHandler.Register(ctx, router)

	err = s.Add(ctx, testAccountID, testAccountNumber, "user", time.Now())
	assert.NoError(t, err, "failed to add account")

	// test get account
	path := strings.Replace(endpointAccount, "{number}", testAccountNumber, -1)
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Header.Add("authorization", "Bearer token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	var account Account
	bytes, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bytes, &account)
	assert.Equal(t, testAccountNumber, account.Number)

	//test list accounts
	r = httptest.NewRequest(http.MethodGet, endpointAccounts, nil)
	r.Header.Add("authorization", "Bearer token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	var accounts []Account
	bytes, err = io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bytes, &accounts)
	assert.Equal(t, 1, len(accounts))

	// test create opt out event when account already exists in store
	r = httptest.NewRequest(http.MethodPost, path, nil)
	r.Header.Add("authorization", "Bearer token")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
	// no msg should be published as account is already opt out
	assert.Equal(t, 0, len(mockPublisher.Msgs))

	// test remove opt out
	r = httptest.NewRequest(http.MethodDelete, path, nil)
	r.Header.Add("authorization", "Bearer token")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	expectedEv := &smart.AccountBookingOptOutRemovedEvent{
		AccountId: testAccountID,
		RemovedBy: "email",
	}
	assert.Equal(t, 1, len(mockPublisher.Msgs))
	assert.Equal(t, expectedEv, mockPublisher.Msgs[0])

	mockPublisher.Msgs = mockPublisher.Msgs[:0]

	// test opt out when account doesn't exist in store
	r = httptest.NewRequest(http.MethodPost, path, nil)
	r.Header.Add("authorization", "Bearer token")
	w = httptest.NewRecorder()

	err = s.Remove(ctx, testAccountID)
	assert.NoError(t, err)

	router.ServeHTTP(w, r)

	expectedOptOutEv := &smart.AccountBookingOptOutAddedEvent{
		AccountId: testAccountID,
		AddedBy:   "email",
	}
	assert.Equal(t, 1, len(mockPublisher.Msgs))
	assert.Equal(t, expectedOptOutEv, mockPublisher.Msgs[0])
}

type accountRepoMock struct {
	accountNumberID map[string]string
}

func (a *accountRepoMock) AccountID(_ context.Context, accountNumber string) (string, error) {
	return a.accountNumberID[accountNumber], nil
}

type identityClientMock struct {
}

func (i *identityClientMock) WhoAmI(_ context.Context, _ *principal.Model) (identity.WhoAmIResult, error) {
	staff := identity.StaffPrincipal{
		ID:    "id",
		Email: "email",
	}
	principalResult := identity.PrincipalResult{Staff: &staff}

	return identity.WhoAmIResult{Principal: &principalResult}, nil
}

type MockSink struct {
	Msgs []proto.Message
}

func (m *MockSink) Sink(_ context.Context, payload proto.Message, _ time.Time) error {
	m.Msgs = append(m.Msgs, payload)
	return nil
}
