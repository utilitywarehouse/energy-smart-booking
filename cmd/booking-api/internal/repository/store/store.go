package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store/migrations"
)

func Setup(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := postgres.Setup(ctx, dsn, migrations.Source)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func GetPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool with dsn: %s, %w", dsn, err)
	}
	return pool, nil
}
