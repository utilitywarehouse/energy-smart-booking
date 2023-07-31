package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

var ErrCampaignabilityNotFound = errors.New("Campaignability not found for occupancy")

type Campaignability struct {
	OccupancyID string
	AccountID   string
	Reasons     domain.IneligibleReasons
}

type CampaignabilityStore struct {
	pool *pgxpool.Pool
}

func NewCampaignability(pool *pgxpool.Pool) *CampaignabilityStore {
	return &CampaignabilityStore{pool: pool}
}

func (s *CampaignabilityStore) Add(ctx context.Context, occupancyID, accountID string, reasons domain.IneligibleReasons) error {
	q := `
	INSERT INTO campaignability(occupancy_id, account_id, reasons)
	VALUES ($1, $2, $3)
	ON CONFLICT (occupancy_id, account_id)
	DO UPDATE 
	SET reasons = $3, updated_at = now();`

	_, err := s.pool.Exec(ctx, q, occupancyID, accountID, reasons)

	return err
}

func (s *CampaignabilityStore) Get(ctx context.Context, occupancyID, accountID string) (Campaignability, error) {
	var campaignability Campaignability
	q := `
	SELECT occupancy_id, account_id, reasons 
	FROM campaignability 
	WHERE occupancy_id = $1
	AND account_id = $2;`
	if err := s.pool.QueryRow(ctx, q, occupancyID, accountID).
		Scan(&campaignability.OccupancyID, &campaignability.AccountID, &campaignability.Reasons); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Campaignability{}, ErrCampaignabilityNotFound
		}
		return Campaignability{}, err
	}

	return campaignability, nil
}
