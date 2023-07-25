package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOccupancyNotFound = errors.New("occupancy not found")

type OccupancyStore struct {
	pool *pgxpool.Pool
}

type Occupancy struct {
	OccupancyID string
	SiteID      string
	AccountID   string
}

func NewOccupancy(pool *pgxpool.Pool) *OccupancyStore {
	return &OccupancyStore{pool: pool}
}

func (s *OccupancyStore) Add(ctx context.Context, occupancyID, siteID, accountID string) error {
	q := `
	INSERT INTO occupancy (occupancy_id, site_id, account_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (id)
	DO NOTHING;`
	_, err := s.pool.Exec(ctx, q, occupancyID, siteID, accountID)

	return err
}

func (s *OccupancyStore) UpdateSite(ctx context.Context, occupancyID, siteID string) error {
	q := `UPDATE occupancy SET site_id = $2 WHERE occupancy_id = $1;`
	_, err := s.pool.Exec(ctx, q, occupancyID, siteID)

	return err
}

func (s *OccupancyStore) GetOccupancyByID(ctx context.Context, occupancyID string) (*Occupancy, error) {
	var occ Occupancy
	q := `SELECT occupancy_id, site_id, account_id FROM occupancy WHERE id = $1;`
	if err := s.pool.QueryRow(ctx, q, occupancyID).
		Scan(&occ.OccupancyID, &occ.SiteID, &occ.AccountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOccupancyNotFound
		}
		return nil, err
	}

	return &occ, nil
}
