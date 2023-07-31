package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSiteNotFound = errors.New("site not found")

type SiteStore struct {
	pool *pgxpool.Pool
}

type Site struct {
	ID       string
	PostCode string
}

func NewSite(pool *pgxpool.Pool) *SiteStore {
	return &SiteStore{pool: pool}
}

func (s *SiteStore) Add(ctx context.Context, id, postCode string, at time.Time) error {
	q := `
	INSERT INTO sites (id, post_code, created_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (id)
	DO UPDATE 
	SET post_code = $2, updated_at = now()	
	`
	_, err := s.pool.Exec(ctx, q, id, postCode, at)

	return err
}

func (s *SiteStore) Get(ctx context.Context, id string) (Site, error) {
	var site Site
	if err := s.pool.QueryRow(ctx, `SELECT id, post_code FROM sites WHERE id = $1`, id).
		Scan(&site.ID, &site.PostCode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Site{}, ErrSiteNotFound
		}
		return Site{}, err
	}

	return site, nil
}
