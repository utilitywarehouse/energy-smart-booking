package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEvaluationNotFound = errors.New("evaluation not found")

type Evaluation struct {
	AccountID   string
	OccupancyID string
	Eligible    bool
	Suppliable  bool
}

type EligibleAccountOccupancy struct {
	AccountID   string
	OccupancyID string
}

type SmartBookingEvaluation struct {
	pool *pgxpool.Pool
}

func NewSmartBookingEvaluation(pool *pgxpool.Pool) *SmartBookingEvaluation {
	return &SmartBookingEvaluation{pool: pool}
}

func (s *SmartBookingEvaluation) UpsertEligibility(ctx context.Context, accountID, occupancyID string, eligible bool) error {
	q := `
	INSERT INTO smart_booking_evaluation(account_id, occupancy_id, eligible)
	VALUES ($1, $2, $3)
	ON CONFLICT (account_id, occupancy_id)
	DO UPDATE 
	SET eligible = $3,
	    updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountID, occupancyID, eligible)

	return err
}

func (s *SmartBookingEvaluation) UpsertSuppliability(ctx context.Context, accountID, occupancyID string, suppliable bool) error {
	q := `
	INSERT INTO smart_booking_evaluation(account_id, occupancy_id, suppliable)
	VALUES ($1, $2, $3)
	ON CONFLICT (account_id, occupancy_id)
	DO UPDATE 
	SET suppliable = $3,
	    updated_at = now();`

	_, err := s.pool.Exec(ctx, q, accountID, occupancyID, suppliable)

	return err
}

func (s *SmartBookingEvaluation) Get(ctx context.Context, accountID, occupancyID string) (Evaluation, error) {
	var result Evaluation

	q := `
	SELECT account_id, occupancy_id, eligible, suppliable 
	FROM smart_booking_evaluation
	WHERE account_id = $1
	AND occupancy_id = $2;`

	if err := s.pool.QueryRow(ctx, q, accountID, occupancyID).Scan(
		&result.AccountID,
		&result.OccupancyID,
		&result.Eligible,
		&result.Suppliable); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Evaluation{}, ErrEvaluationNotFound
		}
		return Evaluation{}, err
	}

	return result, nil
}

func (s *SmartBookingEvaluation) GetEligible(ctx context.Context) ([]EligibleAccountOccupancy, error) {
	q := `
	SELECT account_id, occupancy_id
	FROM smart_booking_evaluation
	WHERE eligible IS TRUE
	AND suppliable IS TRUE;`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make([]EligibleAccountOccupancy, 0)
	for rows.Next() {
		var res EligibleAccountOccupancy
		if err := rows.Scan(&res.AccountID, &res.OccupancyID); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return results, nil
}
