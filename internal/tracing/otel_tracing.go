package tracing

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey struct{}

func FromContext(ctx context.Context) trace.Tracer {
	t, _ := ctx.Value(ctxKey{}).(trace.Tracer)
	return t
}

func NewContext(parent context.Context, t trace.Tracer) context.Context {
	return context.WithValue(parent, ctxKey{}, t)
}

// MarshalContext marshals a parent context
func MarshalContext(ctx context.Context) []byte {
	if ctx == nil {
		return nil
	}

	meta := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, meta)

	res, err := json.Marshal(meta)
	if err != nil {
		log.Warnf("unable to marshal context: %s", err.Error())
		return nil
	}
	return res
}

// UnmarshalContext unmarshals a remote context
func UnmarshalContext(ctx context.Context, b []byte) context.Context {
	if b == nil {
		return ctx
	}

	md := map[string]string{}

	err := json.Unmarshal(b, &md)
	if err != nil {
		log.Warnf("unable to unmarshal context: %s", err.Error())
		return ctx
	}

	meta := propagation.MapCarrier(md)
	return otel.GetTextMapPropagator().Extract(ctx, meta)
}
