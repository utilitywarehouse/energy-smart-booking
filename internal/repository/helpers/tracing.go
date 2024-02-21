package helpers

import (
	"encoding/json"

	"github.com/utilitywarehouse/uwos-go/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func CreateSpanAttribute(v any, kind string, span trace.Span) attribute.KeyValue {
	bytes, err := json.Marshal(v)
	if err != nil {
		tracing.RecordSpanError(span, err)
		return attribute.KeyValue{}
	}
	return attribute.String(kind, string(bytes))
}
