package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSiteNotFound = errors.New("site not found")

type SiteStore struct {
	pool *pgxpool.Pool
}

type Site struct {
	SiteID   string
	Postcode string
}

func NewSite(pool *pgxpool.Pool) *SiteStore {
	return &SiteStore{pool: pool}
}

func (s *SiteStore) Add(ctx context.Context, siteID, postcode string) error {
	q := `
	INSERT INTO site (site_id, postcode)
	VALUES ($1, $2)
	ON CONFLICT (site_id)
	DO UPDATE 
	SET postcode = $2, updated_at = now();
	`
	_, err := s.pool.Exec(ctx, q, siteID, postcode)

	return err
}

func (s *SiteStore) GePostcodeByID(ctx context.Context, siteID string) (string, error) {
	var postcode sql.NullString
	q := `SELECT postcode FROM site WHERE site_id = $1;`
	if err := s.pool.QueryRow(ctx, q, siteID).
		Scan(&postcode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrSiteNotFound
		}
		return "", err
	}

	return postcode.String, nil
}
