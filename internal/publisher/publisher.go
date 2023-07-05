package publisher

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
)

type SyncPublisher interface {
	Sink(ctx context.Context, payload proto.Message, occurredAt time.Time) error
}
