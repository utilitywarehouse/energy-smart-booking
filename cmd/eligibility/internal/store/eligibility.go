package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

var ErrEligibilityNotFound = errors.New("eligibility not found for occupancy")

type Eligibility struct {
	OccupancyID string
	AccountID   string
	Reasons     domain.IneligibleReasons
}

type EligibilityStore struct {
	pool *pgxpool.Pool
}

func NewEligibility(pool *pgxpool.Pool) *EligibilityStore {
	return &EligibilityStore{pool: pool}
}

func (s *EligibilityStore) Add(ctx context.Context, occupancyID, accountID string, reasons domain.IneligibleReasons) error {
	q := `
	INSERT INTO eligibility(occupancy_id, account_id, reasons)
	VALUES ($1, $2, $3)
	ON CONFLICT (occupancy_id, account_id)
	DO UPDATE 
	SET reasons = $3, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, occupancyID, accountID, reasons)

	return err
}

func (s *EligibilityStore) Get(ctx context.Context, occupancyID, accountID string) (Eligibility, error) {
	var eligibility Eligibility

	q := `
	SELECT occupancy_id, account_id, reasons 
	FROM eligibility 
	WHERE occupancy_id = $1 
	AND account_id = $2;`
	if err := s.pool.QueryRow(ctx, q, occupancyID, accountID).
		Scan(&eligibility.OccupancyID, &eligibility.AccountID, &eligibility.Reasons); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Eligibility{}, ErrEligibilityNotFound
		}
		return Eligibility{}, err
	}

	return eligibility, nil
}
