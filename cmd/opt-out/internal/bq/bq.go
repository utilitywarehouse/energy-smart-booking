package bq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/bigquery"
	energy_contracts "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	"github.com/utilitywarehouse/energy-pkg/metrics"
	"github.com/utilitywarehouse/energy-smart-booking/internal/indexer"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type OptOutAdded struct {
	ID      string
	Number  string
	AddedBy string
	AddedAt time.Time
}

type OptOutRemoved struct {
	ID        string
	Number    string
	RemovedBy string
	RemovedAt time.Time
}

// Save returns the BQ query to save an optOut added event.
func (a *OptOutAdded) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"account_id":     a.ID,
		"account_number": a.Number,
		"added_by":       a.AddedBy,
		"added_at":       a.AddedAt,
	}, a.ID, nil
}

// Save returns the BQ query to save an optOut removed event.
func (a *OptOutRemoved) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"account_id":     a.ID,
		"account_number": a.Number,
		"removed_by":     a.RemovedBy,
		"removed_at":     a.RemovedAt,
	}, a.ID, nil
}

type AccountsRepository interface {
	AccountNumber(ctx context.Context, accountID string) (string, error)
}

type BigQueryIndexer struct {
	OptOutAdded   indexer.BigQuery
	OptOutRemoved indexer.BigQuery
	AccountsRepo  AccountsRepository
}

func (i *BigQueryIndexer) PreHandle(_ context.Context) error {
	i.OptOutAdded.Begin()
	i.OptOutRemoved.Begin()

	return nil
}

func (i *BigQueryIndexer) PostHandle(ctx context.Context) error {
	if err := i.OptOutAdded.Commit(ctx); err != nil {
		return err
	}

	return i.OptOutRemoved.Commit(ctx)
}

func (i *BigQueryIndexer) Handle(ctx context.Context, message substrate.Message) error {
	var env energy_contracts.Envelope
	if err := proto.Unmarshal(message.Data(), &env); err != nil {
		return err
	}

	if env.Message == nil {
		slog.Info("skipping empty message")
		metrics.SkippedMessageCounter.WithLabelValues("empty_message").Inc()
		return nil
	}

	inner, err := env.Message.UnmarshalNew()
	if err != nil {
		return fmt.Errorf("error unmarshaling event [%s] %s: %w", env.GetUuid(), env.GetMessage().GetTypeUrl(), err)
	}
	switch x := inner.(type) {
	case *smart.AccountBookingOptOutAddedEvent:
		accountNumber, err := i.AccountsRepo.AccountNumber(ctx, x.GetAccountId())
		if err != nil {
			return fmt.Errorf("failed to get account number for account id %s: %w", x.GetAccountId(), err)
		}

		i.OptOutAdded.Queue(&OptOutAdded{
			ID:      x.GetAccountId(),
			Number:  accountNumber,
			AddedBy: x.GetAddedBy(),
			AddedAt: env.OccurredAt.AsTime(),
		})
	case *smart.AccountBookingOptOutRemovedEvent:
		accountNumber, err := i.AccountsRepo.AccountNumber(ctx, x.GetAccountId())
		if err != nil {
			return fmt.Errorf("failed to get account number for account id %s: %w", x.GetAccountId(), err)
		}

		i.OptOutRemoved.Queue(&OptOutRemoved{
			ID:        x.GetAccountId(),
			Number:    accountNumber,
			RemovedBy: x.GetRemovedBy(),
			RemovedAt: env.OccurredAt.AsTime(),
		})
	default:
		// no nothing
	}
	return nil
}
