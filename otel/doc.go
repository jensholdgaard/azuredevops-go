// Package otel provides opt-in OpenTelemetry instrumentation for the
// azuredevops-go webhooks package.
//
// Use WithTracing, WithMetrics, or WithTelemetry to instrument an EventHandler:
//
//	handler := webhooks.New(
//	    webhooks.WithBasicAuth("user", "pass"),
//	    adootel.WithTelemetry(nil, nil),
//	)
package otel
