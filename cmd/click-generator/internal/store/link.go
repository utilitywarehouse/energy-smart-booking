package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountLinkNotFound = errors.New("account link not found")

type Link struct {
	pool *pgxpool.Pool
}

func NewLink(pool *pgxpool.Pool) *Link {
	return &Link{pool: pool}
}

func (s *Link) Add(ctx context.Context, accountID, occupancyID, link string) error {
	q := `
	INSERT INTO account_links(account_id, occupancy_id, link)
	VALUES ($1, $2, $3)
	ON CONFLICT (account_id, occupancy_id)
	DO UPDATE
	SET link = $3,
	    updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountID, occupancyID, link)

	return err
}

func (s *Link) Remove(ctx context.Context, accountID, occupancyID string) error {
	q := `
	DELETE FROM account_links
	WHERE account_id = $1
	AND occupancy_id = $2;`

	_, err := s.pool.Exec(ctx, q, accountID, occupancyID)

	return err
}

func (s *Link) Get(ctx context.Context, accountID, occupancyID string) (string, error) {
	q := `
	SELECT link from account_links
	WHERE account_id = $1
	AND occupancy_id = $2;`

	var link string
	if err := s.pool.QueryRow(ctx, q, accountID, occupancyID).Scan(&link); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrAccountLinkNotFound
		}
		return "", err
	}

	return link, nil
}
