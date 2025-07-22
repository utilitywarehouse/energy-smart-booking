package store

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store/migrations"
)

var pgDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		slog.Error("failed to set up test container ", "error", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			slog.Error("failed to terminate container ", "error", err)
		}
	}()

	pgDSN, err = postgres.GetTestContainerDSN(container)
	if err != nil {
		slog.Error("failed to set up test container", "error", err)
	}

	pool, err := postgres.Setup(ctx, pgDSN, migrations.Source)
	if err != nil {
		slog.Error("failed to set up postgres", "error", err)
	}
	defer func() {
		if err := postgres.Teardown(pool, migrations.Source); err != nil {
			slog.Error("failed to teardown(migrate down)", "error", err)
		}
	}()

	m.Run()
}

func connect(ctx context.Context) *pgxpool.Pool {
	pool, err := postgres.Connect(ctx, pgDSN)
	if err != nil {
		slog.Error("failed to set up postgres connection pool", "error", err)
	}
	return pool
}
