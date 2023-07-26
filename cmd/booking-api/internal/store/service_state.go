package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-pkg/domain"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceStore struct {
	pool *pgxpool.Pool
}

type Service struct {
	ServiceID   string
	Mpxn        string
	OccupancyID string
	SupplyType  domain.SupplyType
	AccountID   string
	StartDate   *time.Time
	EndDate     *time.Time
	IsLive      bool
}

func NewService(pool *pgxpool.Pool) *ServiceStore {
	return &ServiceStore{pool: pool}
}

func defaultIfNull(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func (s *ServiceStore) Upsert(ctx context.Context, service *Service) error {
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

	_, err := s.pool.Exec(ctx, q,
		service.ServiceID,
		service.Mpxn,
		service.OccupancyID,
		service.SupplyType.String(),
		service.AccountID,
		sql.NullTime{Time: defaultIfNull(service.StartDate), Valid: service.StartDate != nil},
		sql.NullTime{Time: defaultIfNull(service.EndDate), Valid: service.EndDate != nil},
		service.IsLive,
	)

	return err
}
