package publisher

import (
	"context"
	"time"

	energycontracts "github.com/utilitywarehouse/energy-contracts"
	"github.com/uw-labs/substrate"
	"google.golang.org/protobuf/proto"
)

type syncPublisher struct {
	sink    substrate.SynchronousMessageSink
	appName string
}

func NewSyncPublisher(sink substrate.SynchronousMessageSink, appName string) SyncPublisher {
	return &syncPublisher{
		sink:    sink,
		appName: appName,
	}
}

func (p *syncPublisher) Sink(ctx context.Context, proto proto.Message, at time.Time) error {
	//nolint: staticcheck
	msg, err := energycontracts.
		NewMessage(proto).
		WithApplication(p.appName).
		WithOccurredAt(at).
		Build()
	if err != nil {
		return err
	}

	return p.sink.PublishMessage(ctx, msg)
}
