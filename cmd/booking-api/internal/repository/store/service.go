package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrNoMPXNFound     = errors.New("mpxn not found")
)

type ServiceStore struct {
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

func NewService(pool *pgxpool.Pool) *ServiceStore {
	return &ServiceStore{pool: pool}
}

func (s *ServiceStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *ServiceStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)
	s.batch = nil
	return res.Close()
}

func (s *ServiceStore) Upsert(service models.Service) {
	q := `
	INSERT INTO service (
		service_id,
		mpxn,
		occupancy_id,
		supply_type,
		account_id,
		start_date,
		end_date,
		is_live
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
	) ON CONFLICT (service_id)
	DO UPDATE SET
		mpxn = $2,
		occupancy_id = $3,
		account_id = $5,
		start_date = $6,
		end_date = $7,
		is_live = $8;
	`

	s.batch.Queue(q,
		service.ServiceID,
		service.Mpxn,
		service.OccupancyID,
		service.SupplyType.String(),
		service.AccountID,
		sql.NullTime{Time: defaultIfNull(service.StartDate), Valid: service.StartDate != nil},
		sql.NullTime{Time: defaultIfNull(service.EndDate), Valid: service.EndDate != nil},
		service.IsLive)
}

func (s *ServiceStore) GetServiceMPXNByOccupancyID(ctx context.Context, occupancyID string) (string, error) {
	q := `
	SELECT DISTINCT(mpxn)
	FROM service
	WHERE occupancy_id = $1
	AND is_live IS TRUE;
	`

	var mpxn string

	err := s.pool.QueryRow(ctx, q, occupancyID).Scan(&mpxn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNoMPXNFound
		}
		return "", fmt.Errorf("failed to scan row, %w", err)
	}

	return mpxn, nil
}

func defaultIfNull(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
