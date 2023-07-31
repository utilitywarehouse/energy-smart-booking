package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPostCodeNotFound = errors.New("post code not found")

type PostCodeStore struct {
	pool *pgxpool.Pool
}

func NewPostCode(pool *pgxpool.Pool) *PostCodeStore {
	return &PostCodeStore{pool: pool}
}

func (s *PostCodeStore) AddWanCoverage(ctx context.Context, postCode string, covered bool) error {
	q := `
	INSERT INTO postcodes (post_code, wan_coverage)
	VALUES ($1, $2)
	ON CONFLICT(post_code)
	DO UPDATE 
	SET wan_coverage = $2;`

	_, err := s.pool.Exec(ctx, q, postCode, covered)

	return err
}

func (s *PostCodeStore) GetWanCoverage(ctx context.Context, postCode string) (bool, error) {
	var covered bool
	if err := s.pool.QueryRow(ctx, `SELECT wan_coverage FROM postcodes WHERE post_code = $1`, postCode).
		Scan(&covered); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrPostCodeNotFound
		}
		return false, err
	}

	return covered, nil
}
