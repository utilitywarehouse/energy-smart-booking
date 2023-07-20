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

var ErrMeterNotFound = errors.New("meter not found")

type MeterStore struct {
	pool *pgxpool.Pool
}

type Meter struct {
	ID         string
	Mpxn       string
	Msn        string
	SupplyType domain.SupplyType
	Capacity   *float32
	MeterType  string
}

func NewMeter(pool *pgxpool.Pool) *MeterStore {
	return &MeterStore{pool: pool}
}

func (s *MeterStore) Add(ctx context.Context, meter *Meter) error {
	q := `
	INSERT INTO meters(id, mpxn, msn, supply_type, meter_type)
	VALUES ($1, $2, $3, $4, $5);`

	_, err := s.pool.Exec(ctx, q, meter.ID, meter.Mpxn, meter.Msn, meter.SupplyType.String(), meter.MeterType)

	return err
}

func (s *MeterStore) AddMeterCapacity(ctx context.Context, meterID string, cap float32) error {
	_, err := s.pool.Exec(ctx, `UPDATE meters SET capacity = $2 WHERE id = $1;`, meterID, cap)

	return err
}

func (s *MeterStore) AddMeterType(ctx context.Context, meterID string, meterType string) error {
	_, err := s.pool.Exec(ctx, `UPDATE meters SET meter_type = $2 WHERE id = $1;`, meterID, meterType)

	return err
}

func (s *MeterStore) InstallMeter(ctx context.Context, meterID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE meters
		SET installed_at = $2,
		uninstalled_at = NULL
		WHERE id = $1;`, meterID, at)

	return err
}

func (s *MeterStore) ReInstallMeter(ctx context.Context, meterID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE meters
		SET uninstalled_at = NULL
		WHERE id = $1;`, meterID)

	return err
}

func (s *MeterStore) UninstallMeter(ctx context.Context, meterID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE meters
		SET uninstalled_at = $2
		WHERE id = $1;`, meterID, at)

	return err
}

func (s *MeterStore) Get(ctx context.Context, mpxn string) (Meter, error) {
	var (
		meter    Meter
		capacity sql.NullFloat64
	)

	q := `
	SELECT id, mpxn, msn, supply_type, capacity, meter_type
	FROM meters 
	WHERE mpxn = $1
	AND installed_at IS NOT NULL
	AND uninstalled_at IS NULL
	ORDER BY installed_at DESC;`

	err := s.pool.QueryRow(ctx, q, mpxn).Scan(
		&meter.ID,
		&meter.Mpxn,
		&meter.Msn,
		&meter.SupplyType,
		&capacity,
		&meter.MeterType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Meter{}, ErrMeterNotFound
		}
		return Meter{}, err
	}
	if capacity.Valid {
		val := float32(capacity.Float64)
		meter.Capacity = &val
	}

	return meter, nil
}
