package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountStore struct {
	pool *pgxpool.Pool
}

type Account struct {
	ID       string
	PSRCodes []string
	OptOut   bool
}

func NewAccount(pool *pgxpool.Pool) *AccountStore {
	return &AccountStore{pool: pool}
}

func (s *AccountStore) AddPSRCodes(ctx context.Context, accountID string, codes []string) error {
	q := `
	INSERT INTO accounts(id, psr_codes)
	VALUES ($1, $2)
	ON CONFLICT (id)
	DO UPDATE 
	SET psr_codes = $2, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountID, pq.Array(codes))

	return err
}

func (s *AccountStore) AddOptOut(ctx context.Context, accountID string, optOut bool) error {
	q := `
	INSERT INTO accounts(id, opt_out)
	VALUES ($1, $2)
	ON CONFLICT (id)
	DO UPDATE 
	SET opt_out = $2, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountID, optOut)

	return err
}

func (s *AccountStore) GetAccount(ctx context.Context, accountID string) (Account, error) {
	var account Account

	if err := s.pool.QueryRow(ctx, `SELECT id, psr_codes, opt_out FROM accounts WHERE id = $1;`, accountID).
		Scan(&account.ID, &account.PSRCodes, &account.OptOut); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountNotFound
		}
		return Account{}, err
	}

	return account, nil
}
