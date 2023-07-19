package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-pkg/domain"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceStore struct {
	pool *pgxpool.Pool
}

type Service struct {
	ID          string
	Mpxn        string
	OccupancyID string
	SupplyType  domain.SupplyType
	IsLive      bool
	StartDate   *time.Time
	EndDate     *time.Time
}

func NewService(pool *pgxpool.Pool) *ServiceStore {
	return &ServiceStore{pool: pool}
}

func (s *ServiceStore) Add(ctx context.Context, service *Service) error {
	q := `
	INSERT INTO services (id, mpxn, occupancy_id, supply_type, is_live)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id)
	DO UPDATE 
	SET mpxn = $2, occupancy_id = $3, is_live = $5;`

	_, err := s.pool.Exec(ctx, q,
		service.ID,
		service.Mpxn,
		service.OccupancyID,
		service.SupplyType.String(),
		service.IsLive,
	)

	return err
}

func (s *ServiceStore) AddStatDate(ctx context.Context, serviceID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE services set start_date = $2 where id = $1`, serviceID, at)

	return err
}

func (s *ServiceStore) AddEndDate(ctx context.Context, serviceID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE services set end_date = $2 where id = $1`, serviceID, at)

	return err
}

func (s *ServiceStore) Get(ctx context.Context, serviceID string) (Service, error) {
	q := `
	SELECT id, mpxn, occupancy_id, supply_type, is_live, start_date, end_date
	FROM services
	WHERE id = $1;`

	row := s.pool.QueryRow(ctx, q, serviceID)
	service, err := rowIntoService(row)

	return service, err
}

func (s *ServiceStore) GetLiveServicesByOccupancyID(ctx context.Context, occupancyID string) ([]Service, error) {
	q := `
	SELECT id, mpxn, occupancy_id, supply_type, is_live, start_date, end_date
	FROM services
	WHERE occupancy_id = $1
	AND is_live is TRUE;`

	rows, err := s.pool.Query(ctx, q, occupancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]Service, 0)

	for rows.Next() {
		var service Service
		service, err = rowIntoService(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func rowIntoService(row pgx.Row) (Service, error) {
	var (
		service            Service
		startDate, endDate sql.NullTime
	)

	if err := row.Scan(
		&service.ID,
		&service.Mpxn,
		&service.OccupancyID,
		&service.SupplyType,
		&service.IsLive,
		&startDate,
		&endDate,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Service{}, ErrServiceNotFound
		}
		return Service{}, err
	}

	if startDate.Valid {
		tm := startDate.Time
		service.StartDate = &tm
	}
	if endDate.Valid {
		tm := endDate.Time
		service.EndDate = &tm
	}

	return service, nil
}
