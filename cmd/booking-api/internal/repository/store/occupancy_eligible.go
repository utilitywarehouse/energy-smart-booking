package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type OccupancyEligibleStore struct {
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

func NewOccupancyEligible(pool *pgxpool.Pool) *OccupancyEligibleStore {
	return &OccupancyEligibleStore{pool: pool}
}

func (s *OccupancyEligibleStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *OccupancyEligibleStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)

	s.batch = nil
	return res.Close()
}

func (s *OccupancyEligibleStore) Upsert(occupancy models.OccupancyEligibility) {
	q := `
	INSERT INTO occupancy_eligible (occupancy_id, reference)
	VALUES ($1, $2)
	ON CONFLICT (occupancy_id)
	DO UPDATE SET updated_at = NOW();`

	s.batch.Queue(q, occupancy.OccupancyID, occupancy.Reference)
}

func (s *OccupancyEligibleStore) Delete(occupancy models.OccupancyEligibility) {
	q := `UPDATE occupancy_eligible SET deleted_at = NOW() WHERE occupancy_id = $1;`

	s.batch.Queue(q, occupancy.OccupancyID)
}
