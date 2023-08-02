package generator

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/click-generator/internal/store"
)

type AccountsLookup interface {
	AccountNumber(ctx context.Context, accountID string) (string, error)
}

type BookingEvaluationStore interface {
	GetEligible(ctx context.Context) ([]store.EligibleAccountOccupancy, error)
}

type AccountLinkStore interface {
	Add(ctx context.Context, accountID, occupancyID, link string) error
}

type MachineAuthInjector interface {
	ToCtx(context.Context) context.Context
}

type Link struct {
	mai              MachineAuthInjector
	provider         *LinkProvider
	accounts         AccountsLookup
	evaluationStore  BookingEvaluationStore
	accountLinkStore AccountLinkStore
}

func NewLink(
	mai MachineAuthInjector,
	provider *LinkProvider,
	client AccountsLookup,
	evaluationStore BookingEvaluationStore,
	linkStore AccountLinkStore) *Link {
	return &Link{
		mai:              mai,
		provider:         provider,
		accounts:         client,
		evaluationStore:  evaluationStore,
		accountLinkStore: linkStore,
	}
}
func (l *Link) Run(ctx context.Context) error {
	log.Info("Start generating links ...")

	now := time.Now()

	// get all eligible accounts
	accountOccupancies, err := l.evaluationStore.GetEligible(ctx)
	if err != nil {
		return fmt.Errorf("failed to get eligible account occupancies: %w", err)
	}
	// group occupancies under account id, to generate the link only once, if multiple occupancies
	accountIDOccupanciesMap := make(map[string][]string)
	for _, ao := range accountOccupancies {
		if _, ok := accountIDOccupanciesMap[ao.AccountID]; !ok {
			accountIDOccupanciesMap[ao.AccountID] = make([]string, 0)
		}
		accountIDOccupanciesMap[ao.AccountID] = append(accountIDOccupanciesMap[ao.AccountID], ao.OccupancyID)
	}

	for id, occ := range accountIDOccupanciesMap {
		// lookup account number
		accountNumber, err := l.accounts.AccountNumber(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to lookup account number for account id %s: %w", id, err)
		}
		// generate link
		link, err := l.provider.Generate(ctx, accountNumber)
		if err != nil {
			return fmt.Errorf("failed to generate link for account id %s: %w", id, err)
		}
		// persist link for each account id - occupancy id pair
		for _, oID := range occ {
			err = l.accountLinkStore.Add(ctx, id, oID, link)
			if err != nil {
				return fmt.Errorf("failed to persist link for account id %s, occupancy id %s: %w", id, oID, err)
			}
		}
	}
	log.WithField("elapsed", time.Since(now).String()).WithField("total", len(accountIDOccupanciesMap)).Info("Successfully generated links")

	return nil
}
