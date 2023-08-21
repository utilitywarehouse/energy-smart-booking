package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

var ErrServiceNotFound = errors.New("service not found")

type ServiceStore struct {
	pool *pgxpool.Pool
}

type Service struct {
	ID          string
	Mpxn        string
	OccupancyID string
	SupplyType  energy_domain.SupplyType
	IsLive      bool
	StartDate   *time.Time
	EndDate     *time.Time
}

type ServiceBookingRef struct {
	ServiceID  string
	BookingRef string
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

func (s *ServiceStore) AddStartDate(ctx context.Context, serviceID string, at time.Time) error {
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

func (s *ServiceStore) LoadLiveServicesByOccupancyID(ctx context.Context, occupancyID string) ([]domain.Service, error) {
	q := `
	SELECT 
	    s.id, s.mpxn, s.supply_type,
	    m.mpxn, m.profile_class, m.ssc, m.alt_han
	FROM services s 
	LEFT JOIN meterpoints m
	ON s.mpxn = m.mpxn
	WHERE s.occupancy_id = $1
	AND s.is_live is TRUE;`

	rows, err := s.pool.Query(ctx, q, occupancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]domain.Service, 0)

	for rows.Next() {
		var (
			service                           domain.Service
			meterpointMpxn, profileClass, ssc sql.NullString
			altHan                            sql.NullBool
		)
		if err = rows.Scan(
			&service.ID,
			&service.Mpxn,
			&service.SupplyType,
			&meterpointMpxn,
			&profileClass,
			&ssc,
			&altHan); err != nil {
			return nil, err
		}

		if meterpointMpxn.Valid {
			service.Meterpoint = &domain.Meterpoint{
				Mpxn:   meterpointMpxn.String,
				AltHan: altHan.Bool,
				SSC:    ssc.String,
			}
			if profileClass.Valid {
				pc, ok := platform.ProfileClass_value[profileClass.String]
				if !ok {
					return nil, fmt.Errorf("invalid profile class %s", profileClass.String)
				}
				service.Meterpoint.ProfileClass = platform.ProfileClass(pc)
			}
		}

		services = append(services, service)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (s *ServiceStore) GetServicesWithBookingRef(ctx context.Context, occupancyID string) ([]ServiceBookingRef, error) {
	q := `
	SELECT s.id, b.reference
	FROM services s 
	LEFT JOIN booking_references b
	ON s.mpxn = b.mpxn
	WHERE s.occupancy_id = $1
	AND s.is_live is true 
	AND b.deleted_at IS NULL;`

	rows, err := s.pool.Query(ctx, q, occupancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := make([]ServiceBookingRef, 0)

	for rows.Next() {
		var (
			serviceBookingRef ServiceBookingRef
			bookingRef        sql.NullString
		)
		if err = rows.Scan(
			&serviceBookingRef.ServiceID,
			&bookingRef); err != nil {
			return nil, err
		}
		if bookingRef.Valid {
			serviceBookingRef.BookingRef = bookingRef.String
		}
		refs = append(refs, serviceBookingRef)
	}

	return refs, nil
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
