package tracer

import (
	"encoding/json"

	"github.com/opentracing/opentracing-go"
)

func ExtractTracerContext(value []byte) (opentracing.SpanContext, error) {
	var m map[string]string
	err := json.Unmarshal(value, &m)
	if err != nil {
		return nil, err
	}

	incomingTrace, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.TextMapCarrier(m),
	)
	if err != nil {
		return nil, err
	}

	return incomingTrace, err
}

func InjectTracerContext(span opentracing.Span) ([]byte, error) {
	traceContext := make(map[string]string)
	err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(traceContext),
	)
	if err != nil {
		return nil, err
	}

	encodedTraceContext, err := json.Marshal(traceContext)
	if err != nil {
		return nil, err
	}

	return encodedTraceContext, nil
}
