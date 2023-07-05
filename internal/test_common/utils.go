//go:build testing

package test_common

import (
	"context"
	"time"

	"github.com/google/uuid"
	envelope "github.com/utilitywarehouse/energy-contracts/pkg/generated"
	"github.com/uw-labs/substrate"
	"github.com/uw-labs/substrate-tools/message"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockSink struct {
	Msgs []proto.Message
}

func (m *MockSink) Sink(ctx context.Context, payload proto.Message, occurredAt time.Time) error {
	m.Msgs = append(m.Msgs, payload)
	return nil
}

func MakeMessage(msg proto.Message) (substrate.Message, error) {
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
