package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

var ErrSuppliabilityNotFound = errors.New("suppliability not found for occupancy")

type Suppliability struct {
	OccupancyID string
	AccountID   string
	Reasons     domain.IneligibleReasons
}

type SuppliabilityStore struct {
	pool *pgxpool.Pool
}

func NewSuppliability(pool *pgxpool.Pool) *SuppliabilityStore {
	return &SuppliabilityStore{pool: pool}
}

func (s *SuppliabilityStore) Add(ctx context.Context, occupancyID, accountID string, reasons domain.IneligibleReasons) error {
	q := `
	INSERT INTO suppliability(occupancy_id, account_id, reasons)
	VALUES ($1, $2, $3)
	ON CONFLICT (occupancy_id, account_id)
	DO UPDATE 
	SET reasons = $3, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, occupancyID, accountID, reasons)

	return err
}

func (s *SuppliabilityStore) Get(ctx context.Context, occupancyID, accountID string) (Suppliability, error) {
	var suppliability Suppliability

	q := `
	SELECT occupancy_id, account_id, reasons 
	FROM suppliability
	WHERE occupancy_id = $1
	AND account_id = $2;`
	if err := s.pool.QueryRow(ctx, q, occupancyID, accountID).
		Scan(&suppliability.OccupancyID, &suppliability.AccountID, &suppliability.Reasons); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Suppliability{}, ErrSuppliabilityNotFound
		}
		return Suppliability{}, err
	}

	return suppliability, nil
}
