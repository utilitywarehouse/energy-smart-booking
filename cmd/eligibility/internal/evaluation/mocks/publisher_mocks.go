package mock_evaluation

import (
	context "context"
	time "time"

	proto "google.golang.org/protobuf/proto"
)

type MockSink struct {
	Msgs []proto.Message
}

func (m *MockSink) Sink(_ context.Context, payload proto.Message, _ time.Time) error {
	m.Msgs = append(m.Msgs, payload)
	return nil
}
