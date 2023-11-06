package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-pkg/domain"
)

var ErrMeterpointNotFound = errors.New("meterpoint not found")

type MeterpointStore struct {
	pool *pgxpool.Pool
}

type Meterpoint struct {
	Mpxn         string
	SupplyType   domain.SupplyType
	AltHan       bool
	ProfileClass platform.ProfileClass
	SSC          string
}

func NewMeterpoint(pool *pgxpool.Pool) *MeterpointStore {
	return &MeterpointStore{pool: pool}
}

func (s *MeterpointStore) AddProfileClass(ctx context.Context, mpxn string, supplyType domain.SupplyType, profileClass platform.ProfileClass) error {
	q := `
	INSERT INTO meterpoints(mpxn, supply_type, profile_class)
	VALUES ($1, $2, $3)
	ON CONFLICT (mpxn)
	DO UPDATE 
	SET profile_class = $3, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, mpxn, supplyType.String(), profileClass.String())

	return err
}

func (s *MeterpointStore) AddSsc(ctx context.Context, mpxn string, supplyType domain.SupplyType, ssc string) error {
	q := `
	INSERT INTO meterpoints(mpxn, supply_type, ssc)
	VALUES ($1, $2, $3)
	ON CONFLICT (mpxn)
	DO UPDATE 
	SET ssc = $3, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, mpxn, supplyType.String(), ssc)

	return err
}

func (s *MeterpointStore) AddAltHan(ctx context.Context, mpxn string, supplyType domain.SupplyType, altHan bool) error {
	q := `
	INSERT INTO meterpoints(mpxn, supply_type, alt_han)
	VALUES ($1, $2, $3)
	ON CONFLICT (mpxn)
	DO UPDATE
	SET alt_han = $3, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, mpxn, supplyType.String(), altHan)

	return err
}

func (s *MeterpointStore) Get(ctx context.Context, mpxn string) (Meterpoint, error) {
	var (
		meterpoint        Meterpoint
		profileClass, ssc sql.NullString
	)

	q := `
	SELECT mpxn, supply_type, profile_class, ssc, alt_han
	FROM meterpoints
	WHERE mpxn = $1;`

	err := s.pool.QueryRow(ctx, q, mpxn).Scan(
		&meterpoint.Mpxn,
		&meterpoint.SupplyType,
		&profileClass,
		&ssc,
		&meterpoint.AltHan,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Meterpoint{}, ErrMeterpointNotFound
		}
		return Meterpoint{}, err
	}
	if profileClass.Valid {
		pc, ok := platform.ProfileClass_value[profileClass.String]
		if !ok {
			return Meterpoint{}, fmt.Errorf("invalid profile class %s", profileClass.String)
		}
		meterpoint.ProfileClass = platform.ProfileClass(pc)
	}
	if ssc.Valid {
		meterpoint.SSC = ssc.String
	}

	return meterpoint, nil
}

func (s *MeterpointStore) GetAltHan(ctx context.Context, mpxn string) (bool, error) {
	q := `
		SELECT alt_han
		FROM meterpoints
		WHERE mpxn = $1;
	`
	isAltHan := false
	err := s.pool.QueryRow(ctx, q, mpxn).Scan(&isAltHan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrMeterpointNotFound
		}
		return false, err
	}
	return isAltHan, nil
}
