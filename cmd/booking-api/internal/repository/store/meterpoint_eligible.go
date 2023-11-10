package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MeterpointEligibleStore struct {
	pool *pgxpool.Pool
	ttl  time.Duration
}

func NewMeterpointEligible(pool *pgxpool.Pool, ttl time.Duration) *MeterpointEligibleStore {
	return &MeterpointEligibleStore{pool: pool, ttl: ttl}
}

func (s *MeterpointEligibleStore) CacheEligibility(ctx context.Context, mpxn string, eligible bool) error {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	var eligibleTime *time.Time = nil
	if eligible {
		eligibleTime = &now
	}
	return s.Upsert(ctx, mpxn, eligibleTime, expiresAt)
}

func (s *MeterpointEligibleStore) Upsert(ctx context.Context, mpxn string, eligible *time.Time, expiresAt time.Time) error {
	q := `
	INSERT INTO meterpoint_eligible (mpxn, eligible, expires_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (mpxn)
	DO UPDATE SET expires_at = $2, updated_at = NOW();
	`
	_, err := s.pool.Exec(ctx, q, mpxn, eligible, expiresAt)
	return err
}

func (s *MeterpointEligibleStore) GetEligibilityForMpxn(ctx context.Context, mpxn string) (bool, bool, error) {
	q := `
	SELECT eligible, expires_at
	FROM meterpoint_eligible
	WHERE mpxn = $1;
	`
	var (
		eligible  *time.Time
		expiresAt time.Time
	)
	err := s.pool.QueryRow(ctx, q, mpxn).Scan(&eligible, &expiresAt)
	if err != nil {
		return false, false, err
	}

	now := time.Now()
	if now.After(expiresAt) {
		return false, false, nil
	}
	return (eligible != nil), true, nil
}
