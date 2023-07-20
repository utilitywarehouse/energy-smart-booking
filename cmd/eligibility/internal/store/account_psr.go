package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

var ErrAccountPSRCodesNotFound = errors.New("account psr codes not found")

type AccountPSRStore struct {
	pool *pgxpool.Pool
}

func NewAccountPSR(pool *pgxpool.Pool) *AccountPSRStore {
	return &AccountPSRStore{pool: pool}
}

func (s *AccountPSRStore) Add(ctx context.Context, accountID string, codes []string) error {
	q := `
	INSERT INTO account_psr(id, psr_codes)
	VALUES ($1, $2)
	ON CONFLICT (id)
	DO UPDATE 
	SET psr_codes = $2, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountID, pq.Array(codes))

	return err
}

func (s *AccountPSRStore) Remove(ctx context.Context, accountID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM account_psr WHERE id = $1;`, accountID)

	return err
}
func (s *AccountPSRStore) GetPSRCodes(ctx context.Context, accountID string) ([]string, error) {
	var codes []string

	if err := s.pool.QueryRow(ctx, `SELECT psr_codes FROM account_psr WHERE id = $1;`, accountID).
		Scan(&codes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountPSRCodesNotFound
		}
		return nil, err
	}

	return codes, nil
}
