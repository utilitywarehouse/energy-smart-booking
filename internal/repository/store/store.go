package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/store/migrations"
)

func Setup(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := postgres.Setup(ctx, dsn, migrations.Source)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
