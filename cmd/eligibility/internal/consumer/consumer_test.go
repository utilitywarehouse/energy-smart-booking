package consumer_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	envelope "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store/migrations"
	"github.com/uw-labs/substrate"
	"github.com/uw-labs/substrate-tools/message"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	dsn  string
	pool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.SetupTestContainer(ctx)
	if err != nil {
		logrus.WithError(err).Panic("unable to create postgres test container")
	}
	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			logrus.WithError(err).Panic("unable to terminate test container")
		}
	}()

	dsn, err = postgres.GetTestContainerDSN(container)
	if err != nil {
		logrus.WithError(err).Panic("unable to get dsn from test container")
	}

	pool, err = postgres.Setup(ctx, dsn, migrations.Source)
	if err != nil {
		logrus.WithError(err).Panic("unable to connect to/run migration on database")
	}
	defer func() {
		if err = postgres.Teardown(pool, migrations.Source); err != nil {
			logrus.WithError(err).Panic("unable to teardown database")
		}
	}()

	m.Run()
}

func truncateDB(t *testing.T) {
	pool.Exec(t.Context(), "TRUNCATE TABLE booking_reference, service, occupancy, site, booking, occupancy_eligible, partial_booking, smart_meter_interest CASCADE")
}

func makeMessage(msg proto.Message) (substrate.Message, error) {
	ts := timestamppb.Now()

	payload, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}

	env := &envelope.Envelope{
		Uuid:       uuid.New().String(),
		CreatedAt:  timestamppb.Now(),
		Message:    payload,
		OccurredAt: ts,
		Sender: &envelope.Sender{
			Application: "smart-scheduler",
		},
	}

	bytes, err := proto.Marshal(env)
	if err != nil {
		return nil, err
	}

	return message.NewMessage(bytes), nil
}
