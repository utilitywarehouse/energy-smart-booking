package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountOptOutStore struct {
	pool *pgxpool.Pool
}

type Account struct {
	ID      string
	Number  string
	AddedBy string
	AddedAt time.Time
}

func NewAccountOptOut(pool *pgxpool.Pool) *AccountOptOutStore {
	return &AccountOptOutStore{pool: pool}
}

func (s *AccountOptOutStore) Add(ctx context.Context, id, number, addedBy string, at time.Time) error {
	q := `
	INSERT INTO opt_out_account (id, number, added_by, created_at) 
	VALUES ($1, $2, $3, $4);`
	_, err := s.pool.Exec(ctx, q, id, number, addedBy, at)
	return err
}

func (s *AccountOptOutStore) Get(ctx context.Context, id string) (*Account, error) {
	var account Account
	if err := s.pool.QueryRow(ctx, `SELECT * from opt_out_account WHERE id = $1;`, id).
		Scan(&account.ID, &account.Number, &account.AddedBy, &account.AddedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return &account, nil
}

func (s *AccountOptOutStore) Remove(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM opt_out_account WHERE id = $1`, id)
	return err
}

func (s *AccountOptOutStore) List(ctx context.Context) ([]Account, error) {
	rows, err := s.pool.Query(ctx, `SELECT * FROM opt_out_account ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)

	for rows.Next() {
		var account Account
		err = rows.Scan(&account.ID, &account.Number, &account.AddedBy, &account.AddedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
