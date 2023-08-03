package generator

import (
	"context"
	"fmt"
)

type AccountLinkStore interface {
	Add(ctx context.Context, accountID, link string) error
}

type MachineAuthInjector interface {
	ToCtx(context.Context) context.Context
}

type Link struct {
	mai              MachineAuthInjector
	provider         *LinkProvider
	accountLinkStore AccountLinkStore
}

func NewLink(
	mai MachineAuthInjector,
	provider *LinkProvider,
	linkStore AccountLinkStore) *Link {
	return &Link{
		mai:              mai,
		provider:         provider,
		accountLinkStore: linkStore,
	}
}
func (l *Link) Generate(ctx context.Context, accountNumber string) error {
	link, err := l.provider.GenerateAuthenticated(ctx, accountNumber)
	if err != nil {
		return fmt.Errorf("failed to generate link for account number %s: %w", accountNumber, err)
	}

	err = l.accountLinkStore.Add(ctx, accountNumber, link)
	if err != nil {
		return fmt.Errorf("failed to persist link for account number %s: %w", accountNumber, err)
	}

	return nil
}
