package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOccupancyNotFound = errors.New("occupancy not found")

type OccupancyStore struct {
	pool *pgxpool.Pool
}

type Occupancy struct {
	ID        string
	SiteID    string
	AccountID string
}

func NewOccupancy(pool *pgxpool.Pool) *OccupancyStore {
	return &OccupancyStore{pool: pool}
}

func (s *OccupancyStore) Add(ctx context.Context, id, siteID, accountID string, at time.Time) error {
	q := `
	INSERT INTO occupancies (id, site_id, account_id, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id)
	DO NOTHING;`
	_, err := s.pool.Exec(ctx, q, id, siteID, accountID, at)

	return err
}

func (s *OccupancyStore) AddSite(ctx context.Context, occupancyID, siteID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE occupancies SET site_id = $2 WHERE id = $1;`, occupancyID, siteID)

	return err
}

func (s *OccupancyStore) Get(ctx context.Context, id string) (Occupancy, error) {
	var occ Occupancy
	if err := s.pool.QueryRow(ctx, `SELECT id, site_id, account_id FROM occupancies WHERE id = $1`, id).
		Scan(&occ.ID, &occ.SiteID, &occ.AccountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Occupancy{}, ErrOccupancyNotFound
		}
		return Occupancy{}, err
	}

	return occ, nil
}

func (s *OccupancyStore) GetIDsByAccount(ctx context.Context, accountID string) ([]string, error) {
	q := `SELECT id FROM occupancies WHERE account_id = $1`

	return s.queryOccupanciesByIdentifier(ctx, q, accountID)
}

func (s *OccupancyStore) GetIDsBySite(ctx context.Context, siteID string) ([]string, error) {
	q := `SELECT id FROM occupancies WHERE site_id = $1`

	return s.queryOccupanciesByIdentifier(ctx, q, siteID)
}

// GetIDsByPostcode gets occupancies by postcode.
func (s *OccupancyStore) GetIDsByPostcode(ctx context.Context, postCode string) ([]string, error) {
	q := `SELECT id FROM occupancies WHERE site_id IN (SELECT id FROM sites WHERE post_code = $1)`

	return s.queryOccupanciesByIdentifier(ctx, q, postCode)
}

func (s *OccupancyStore) queryOccupanciesByIdentifier(ctx context.Context, query, id string) ([]string, error) {
	var ids = make([]string, 0)

	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var occID string
		err = rows.Scan(&occID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, occID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
