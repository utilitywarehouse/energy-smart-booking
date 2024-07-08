package publisher

import (
	"context"
	"time"

	"github.com/uw-labs/substrate"
	"github.com/uw-labs/substrate-tools/message"
	"google.golang.org/protobuf/proto"
)

type billPublisher struct {
	sink substrate.SynchronousMessageSink
}

func NewBillPublisher(sink substrate.SynchronousMessageSink) SyncPublisher {
	return &billPublisher{
		sink: sink,
	}
}

func (b *billPublisher) Sink(ctx context.Context, evt proto.Message, at time.Time) error {
	msg, err := proto.Marshal(evt)
	if err != nil {
		return err
	}

	return b.sink.PublishMessage(ctx, message.NewMessage(msg))
}
