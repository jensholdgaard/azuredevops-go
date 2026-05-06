package otel

import (
	"github.com/jensholdgaard/azuredevops-go/webhooks"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const scopeName = "github.com/jensholdgaard/azuredevops-go/otel"

// WithTracing instruments HandleEventRequest with distributed tracing.
// If tp is nil, the global TracerProvider is used.
func WithTracing(tp trace.TracerProvider) webhooks.Option {
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	obs := &observer{tracer: tp.Tracer(scopeName)}
	return webhooks.WithRequestObserver(obs)
}

// WithMetrics records event counters and processing duration histograms.
// If mp is nil, the global MeterProvider is used.
func WithMetrics(mp metric.MeterProvider) webhooks.Option {
	if mp == nil {
		mp = otel.GetMeterProvider()
	}
	m := mp.Meter(scopeName)
	eventCounter, _ := m.Int64Counter("azuredevops.webhook.events",
		metric.WithUnit("{event}"),
		metric.WithDescription("Total webhook events received"),
	)
	durationHist, _ := m.Float64Histogram("azuredevops.webhook.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Webhook event processing duration"),
	)
	obs := &observer{
		eventCounter: eventCounter,
		durationHist: durationHist,
	}
	return webhooks.WithRequestObserver(obs)
}

// WithTelemetry enables both tracing and metrics.
func WithTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) webhooks.Option {
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	if mp == nil {
		mp = otel.GetMeterProvider()
	}
	m := mp.Meter(scopeName)
	eventCounter, _ := m.Int64Counter("azuredevops.webhook.events",
		metric.WithUnit("{event}"),
		metric.WithDescription("Total webhook events received"),
	)
	durationHist, _ := m.Float64Histogram("azuredevops.webhook.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Webhook event processing duration"),
	)
	obs := &observer{
		tracer:       tp.Tracer(scopeName),
		eventCounter: eventCounter,
		durationHist: durationHist,
	}
	return webhooks.WithRequestObserver(obs)
}
