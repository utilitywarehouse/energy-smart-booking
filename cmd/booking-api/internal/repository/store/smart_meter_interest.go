package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SmartMeterInterestStore struct {
	pool *pgxpool.Pool
}

func NewSmartMeterInterestStore(pool *pgxpool.Pool) *SmartMeterInterestStore {
	return &SmartMeterInterestStore{pool: pool}
}

func (s *SmartMeterInterestStore) Insert(ctx context.Context, registrationID, accountID, reason string, interested bool, createdAt time.Time) error {

	q := `
	INSERT INTO smart_meter_interest (registration_id, account_id, interested, reason, created_at)
    VALUES ($1, $2, $3, $4, %5)
	ON CONFLICT (registration_id)
	DO NOTHING;`

	if _, err := s.pool.Exec(ctx, q, registrationID, accountID, interested, reason, createdAt); err != nil {
		return fmt.Errorf("failed to insert smart meter interest for account ID %s: %w", accountID, err)
	}

	return nil
}
