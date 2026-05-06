package otel

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jensholdgaard/azuredevops-go/webhooks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type observer struct {
	tracer       trace.Tracer
	eventCounter metric.Int64Counter
	durationHist metric.Float64Histogram
}

func (o *observer) ObserveRequest(ctx context.Context, r *http.Request) (context.Context, func(webhooks.EventType, string, any, error)) {
	start := time.Now()

	var span trace.Span
	if o.tracer != nil {
		ctx, span = o.tracer.Start(ctx, "azuredevops.webhook.handle",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(r.Method),
			),
		)
	}

	return ctx, func(eventType webhooks.EventType, deliveryID string, _ any, err error) {
		duration := time.Since(start).Seconds()
		outcome := "success"
		if err != nil {
			outcome = "error"
		}

		attrs := []attribute.KeyValue{
			semconv.MessagingOperationName(string(eventType)),
		}
		if deliveryID != "" {
			attrs = append(attrs, semconv.MessagingMessageID(deliveryID))
		}

		if span != nil {
			span.SetAttributes(attrs...)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.SetAttributes(attribute.String("error.type", classifyError(err)))
			} else {
				span.SetStatus(codes.Ok, "")
			}
			span.End()
		}

		metricAttrs := metric.WithAttributes(
			semconv.MessagingOperationName(string(eventType)),
			attribute.String("outcome", outcome),
		)
		if o.eventCounter != nil {
			o.eventCounter.Add(ctx, 1, metricAttrs)
		}
		if o.durationHist != nil {
			o.durationHist.Record(ctx, duration, metricAttrs)
		}
	}
}

func classifyError(err error) string {
	switch {
	case errors.Is(err, webhooks.ErrBasicAuthMissing), errors.Is(err, webhooks.ErrBasicAuthFailed):
		return "auth"
	case errors.Is(err, webhooks.ErrInvalidHTTPMethod),
		errors.Is(err, webhooks.ErrParsingPayload),
		errors.Is(err, webhooks.ErrEmptyBody),
		errors.Is(err, webhooks.ErrUnknownEventType):
		return "parse"
	default:
		return "handler"
	}
}
