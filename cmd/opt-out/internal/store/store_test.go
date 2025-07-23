package store

import (
	"context"
	"log/slog"
	"os"
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
		os.Exit(1)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			slog.Error("failed to terminate container ", "error", err)
			os.Exit(1)
		}
	}()

	pgDSN, err = postgres.GetTestContainerDSN(container)
	if err != nil {
		slog.Error("failed to set up test container", "error", err)
		os.Exit(1)
	}

	pool, err := postgres.Setup(ctx, pgDSN, migrations.Source)
	if err != nil {
		slog.Error("failed to set up postgres", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := postgres.Teardown(pool, migrations.Source); err != nil {
			slog.Error("failed to teardown(migrate down)", "error", err)
			os.Exit(1)
		}
	}()

	m.Run()
}

func connect(ctx context.Context) *pgxpool.Pool {
	pool, err := postgres.Connect(ctx, pgDSN)
	if err != nil {
		slog.Error("failed to set up postgres connection pool", "error", err)
		os.Exit(1)
	}
	return pool
}
