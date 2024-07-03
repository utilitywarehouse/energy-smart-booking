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
	ErrRegistrationNotFound = errors.New("smart meter interest registration was not found")
)

type SmartMeterInterestStore struct {
	pool *pgxpool.Pool
}

func NewSmartMeterInterestStore(pool *pgxpool.Pool) *SmartMeterInterestStore {
	return &SmartMeterInterestStore{pool: pool}
}

func (s *SmartMeterInterestStore) Insert(ctx context.Context, smi models.SmartMeterInterest) error {

	q := `
	INSERT INTO smart_meter_interest (registration_id, account_id, interested, reason, created_at)
    VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (registration_id)
	DO NOTHING;`

	if _, err := s.pool.Exec(ctx, q, smi.RegistrationID, smi.AccountID, smi.Interested, smi.Reason, smi.CreatedAt); err != nil {
		return fmt.Errorf("failed to insert smart meter interest for account ID %s: %w", smi.AccountID, err)
	}

	return nil
}

func (s *SmartMeterInterestStore) Get(ctx context.Context, registrationID string) (*models.SmartMeterInterest, error) {

	var smi models.SmartMeterInterest
	q := `
	SELECT registration_id, account_id, interested, reason, created_at
	FROM smart_meter_interest
	WHERE registration_id = $1;`

	err := s.pool.QueryRow(ctx, q, registrationID).Scan(&smi.RegistrationID, &smi.AccountID, &smi.Interested, &smi.Reason, &smi.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRegistrationNotFound
		}
		return nil, fmt.Errorf("failed to get smart meter interest, %w", err)
	}

	return &smi, nil
}
