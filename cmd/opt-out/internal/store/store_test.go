package store

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/opt-out/internal/store/migrations"
)

var pgDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			logrus.Fatal(err)
		}
	}()

	pgDSN, err = postgres.GetTestContainerDSN(container)
	if err != nil {
		logrus.Fatal(err)
	}

	pool, err := postgres.Setup(ctx, pgDSN, migrations.Source)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := postgres.Teardown(pool, migrations.Source); err != nil {
			logrus.Fatal(err)
		}
	}()

	m.Run()
}

func connect(ctx context.Context) *pgxpool.Pool {
	pool, err := postgres.Connect(ctx, pgDSN)
	if err != nil {
		logrus.Fatal(err)
	}
	return pool
}
