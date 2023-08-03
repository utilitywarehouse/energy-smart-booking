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

func (s *Link) Add(ctx context.Context, accountNumber, link string) error {
	q := `
	INSERT INTO account_links(account_number, link)
	VALUES ($1, $2)
	ON CONFLICT (account_number)
	DO UPDATE
	SET link = $2,
	    updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountNumber, link)

	return err
}

func (s *Link) Remove(ctx context.Context, accountNumber string) error {
	q := `
	DELETE FROM account_links
	WHERE account_number = $1;`

	_, err := s.pool.Exec(ctx, q, accountNumber)

	return err
}

func (s *Link) Get(ctx context.Context, accountNumber string) (string, error) {
	q := `
	SELECT link from account_links
	WHERE account_number = $1;`

	var link string
	if err := s.pool.QueryRow(ctx, q, accountNumber).Scan(&link); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrAccountLinkNotFound
		}
		return "", err
	}

	return link, nil
}
