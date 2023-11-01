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
	ErrPointOfSaleCustomerDetailsNotFound = errors.New("point of sale customer details not found")
)

type PointOfSaleSerialiser interface {
	Serialize(models.PointOfSaleCustomerDetails) ([]byte, error)
	Deserialize([]byte) (models.PointOfSaleCustomerDetails, error)
}

type PointOfSaleCustomerDetails struct {
	pool       *pgxpool.Pool
	serialiser PointOfSaleSerialiser
}

func NewPointOfSaleCustomerDetails(pool *pgxpool.Pool, serialiser PointOfSaleSerialiser) *PointOfSaleCustomerDetails {
	return &PointOfSaleCustomerDetails{pool, serialiser}
}

func (s *PointOfSaleCustomerDetails) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.PointOfSaleCustomerDetails, error) {

	var accNumber string
	var details []byte

	err := s.pool.QueryRow(ctx, `
		SELECT account_number, details FROM point_of_sale_customer_details WHERE account_number = $1 AND deleted_at IS NULL;
	`, accountNumber).Scan(&accNumber, &details)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w, %w", ErrPointOfSaleCustomerDetailsNotFound, err)
		}
		return nil, fmt.Errorf("failed to query customer details by account number %w", err)
	}

	unserialized, err := s.serialiser.Deserialize(details)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize details, %w", err)
	}

	return &unserialized, nil
}

func (s *PointOfSaleCustomerDetails) Upsert(ctx context.Context, accountNumber string, details models.PointOfSaleCustomerDetails) error {

	serializedDetails, err := s.serialiser.Serialize(details)
	if err != nil {
		return fmt.Errorf("failed to serialize model, %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO point_of_sale_customer_details
		VALUES($1, $2)
		ON CONFLICT(account_number)
		DO UPDATE SET details = $2;`, accountNumber, serializedDetails)

	if err != nil {
		return fmt.Errorf("failed to upsert customer details, %w", err)
	}

	return nil
}

func (s *PointOfSaleCustomerDetails) Delete(ctx context.Context, accountNumber string) error {

	_, err := s.pool.Exec(ctx, `
		UPDATE point_of_sale_customer_details
		SET deleted_at = NOW()
		WHERE account_number = $1;`, accountNumber)
	if err != nil {
		return fmt.Errorf("failed to upsert customer details, %w", err)
	}

	return nil
}
