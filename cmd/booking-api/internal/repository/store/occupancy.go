package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrOccupancyNotFound = errors.New("occupancy not found")
	ErrFailedUpdate      = errors.New("no rows were affected by update statement")
)

type OccupancyStore struct {
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

func NewOccupancy(pool *pgxpool.Pool) *OccupancyStore {
	return &OccupancyStore{pool: pool}
}

func (s *OccupancyStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *OccupancyStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)

	s.batch = nil
	return res.Close()
}

func (s *OccupancyStore) Insert(occupancy models.Occupancy) {
	q := `
	INSERT INTO occupancy (occupancy_id, site_id, account_id, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (occupancy_id)
	DO NOTHING;`

	s.batch.Queue(q, occupancy.OccupancyID, occupancy.SiteID, occupancy.AccountID, occupancy.CreatedAt)
}

func (s *OccupancyStore) UpdateSiteID(occupancyID, siteID string) {
	q := `UPDATE occupancy SET site_id = $2 WHERE occupancy_id = $1;`

	s.batch.Queue(q, occupancyID, siteID)
}

func (s *OccupancyStore) GetOccupancyByID(ctx context.Context, occupancyID string) (*models.Occupancy, error) {
	var occ models.Occupancy
	q := `SELECT occupancy_id, site_id, account_id FROM occupancy WHERE occupancy_id = $1;`
	if err := s.pool.QueryRow(ctx, q, occupancyID).
		Scan(&occ.OccupancyID, &occ.SiteID, &occ.AccountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOccupancyNotFound
		}
		return nil, err
	}

	return &occ, nil
}

func (s *OccupancyStore) GetLiveOccupanciesByAccountID(ctx context.Context, accountID string) ([]models.Occupancy, error) {
	occupancies := make([]models.Occupancy, 0)

	q := `SELECT 
		o.occupancy_id,
		o.site_id,
		o.account_id,
		o.created_at
	FROM 
		occupancy o
	INNER JOIN service s 
		ON o.occupancy_id = s.occupancy_id
	WHERE
		o.account_id = $1
	AND
		s.is_live IS TRUE
	ORDER BY
		o.created_at DESC;`

	rows, err := s.pool.Query(ctx, q, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query get occupancies by account id, %w", err)
	}

	for rows.Next() {
		var occ models.Occupancy

		err := rows.Scan(&occ.OccupancyID, &occ.SiteID, &occ.AccountID, &occ.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row, %w", err)

		}

		occupancies = append(occupancies, occ)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error found in rows, %w", err)
	}

	return occupancies, nil
}
