package consumer

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	envelope "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store/migrations"
	"github.com/utilitywarehouse/energy-smart-booking/internal/testcommon"
	"github.com/uw-labs/substrate"
	"github.com/uw-labs/substrate-tools/message"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestOptOutConsumer(t *testing.T) {
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
	accountRepo := &accountRepoMock{
		accountIDNumber: map[string]string{
			"accountId1": "accountNumber",
		},
	}
	handler := Handle(s, accountRepo)

	msgs := []substrate.Message{}
	optOutEv1, err := testcommon.MakeMessage(&smart.AccountBookingOptOutAddedEvent{
		AccountId: "accountId1",
	})
	assert.NoError(t, err)

	optOutEv2, err := testcommon.MakeMessage(&smart.AccountBookingOptOutAddedEvent{
		AccountId: "accountId2",
		AddedBy:   "user",
	})
	assert.NoError(t, err)

	err = handler(ctx, append(msgs, optOutEv1, optOutEv2))
	assert.NoError(t, err)

	optOutAccounts, err := s.List(ctx)
	assert.NoError(t, err, "failed to list opt out accounts")
	assert.Equal(t, 2, len(optOutAccounts))

	optOutRemovedEv, err := testcommon.MakeMessage(&smart.AccountBookingOptOutRemovedEvent{
		AccountId: "accountId1",
	})

	err = handler(ctx, []substrate.Message{optOutRemovedEv})
	assert.NoError(t, err, "failed to handle opt out removed event")
	optOutAccounts, err = s.List(ctx)
	assert.NoError(t, err, "failed to list opt out accounts")
	assert.Equal(t, 1, len(optOutAccounts))
}

func MakeMessage(msg proto.Message) (substrate.Message, error) {
	ts := timestamppb.Now()

	payload, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}

	env := &envelope.Envelope{
		Uuid:       uuid.New().String(),
		CreatedAt:  timestamppb.Now(),
		Message:    payload,
		OccurredAt: ts,
		Sender: &envelope.Sender{
			Application: "app",
		},
	}

	bytes, err := proto.Marshal(env)
	if err != nil {
		return nil, err
	}

	return message.NewMessage(bytes), nil
}

type accountRepoMock struct {
	accountIDNumber map[string]string
}

func (a *accountRepoMock) AccountNumber(_ context.Context, accountNumber string) (string, error) {
	return a.accountIDNumber[accountNumber], nil
}
