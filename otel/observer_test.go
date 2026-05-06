package otel

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jensholdgaard/azuredevops-go/webhooks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const validPayload = `{
	"subscriptionId": "sub-1",
	"notificationId": 1,
	"id": "event-1",
	"eventType": "git.push",
	"publisherId": "tfs",
	"message": {"text": "pushed"},
	"detailedMessage": {"text": "pushed details"},
	"resource": {
		"commits": [],
		"refUpdates": [{"name": "refs/heads/main", "oldObjectId": "000", "newObjectId": "111"}],
		"repository": {"id": "repo-1", "name": "my-repo", "url": "https://dev.azure.com", "project": {"id": "proj-1", "name": "MyProject", "state": "wellFormed"}, "defaultBranch": "refs/heads/main", "remoteUrl": "https://dev.azure.com/org/repo"},
		"pushedBy": {"displayName": "User", "uniqueName": "user@example.com", "id": "user-1"},
		"pushId": 1,
		"date": "2024-01-01T00:00:00Z",
		"url": "https://dev.azure.com"
	},
	"resourceVersion": "1.0",
	"resourceContainers": {
		"collection": {"id": "col-1"},
		"account": {"id": "acc-1"},
		"project": {"id": "proj-1"}
	},
	"createdDate": "2024-01-01T00:00:00Z"
}`

func TestWithTracing_CreatesSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	handler := webhooks.New(WithTracing(tp))
	handler.OnGitPush(func(_ context.Context, _ string, _ *webhooks.GitPushEvent) error {
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(validPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vss-Activityid", "delivery-123")

	err := handler.HandleEventRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name != "azuredevops.webhook.handle" {
		t.Errorf("expected span name %q, got %q", "azuredevops.webhook.handle", span.Name)
	}
	if span.Status.Code != codes.Ok {
		t.Errorf("expected status Ok, got %v", span.Status.Code)
	}

	// Check messaging.message.id attribute
	found := false
	for _, attr := range span.Attributes {
		if attr.Key == "messaging.message.id" && attr.Value.AsString() == "delivery-123" {
			found = true
		}
	}
	if !found {
		t.Error("expected messaging.message.id attribute with value delivery-123")
	}
}

func TestWithTracing_RecordsError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	handler := webhooks.New(WithTracing(tp))
	handler.OnGitPush(func(_ context.Context, _ string, _ *webhooks.GitPushEvent) error {
		return errors.New("handler failed")
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(validPayload))
	req.Header.Set("Content-Type", "application/json")

	_ = handler.HandleEventRequest(req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Status.Code != codes.Error {
		t.Errorf("expected status Error, got %v", span.Status.Code)
	}

	// Check error.type attribute
	found := false
	for _, attr := range span.Attributes {
		if attr.Key == "error.type" && attr.Value.AsString() == "handler" {
			found = true
		}
	}
	if !found {
		t.Error("expected error.type=handler attribute")
	}
}

func TestWithTracing_AuthError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	handler := webhooks.New(
		webhooks.WithBasicAuth("user", "pass"),
		WithTracing(tp),
	)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(validPayload))

	_ = handler.HandleEventRequest(req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Status.Code != codes.Error {
		t.Errorf("expected status Error, got %v", span.Status.Code)
	}

	found := false
	for _, attr := range span.Attributes {
		if attr.Key == "error.type" && attr.Value.AsString() == "auth" {
			found = true
		}
	}
	if !found {
		t.Error("expected error.type=auth attribute")
	}
}

func TestWithMetrics_CountsEvents(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	handler := webhooks.New(WithMetrics(mp))
	handler.OnGitPush(func(_ context.Context, _ string, _ *webhooks.GitPushEvent) error {
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(validPayload))
	req.Header.Set("Content-Type", "application/json")

	if err := handler.HandleEventRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	foundCounter := false
	foundHist := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "azuredevops.webhook.events":
				foundCounter = true
				data := m.Data.(metricdata.Sum[int64])
				if len(data.DataPoints) == 0 {
					t.Fatal("no data points for counter")
				}
				dp := data.DataPoints[0]
				if dp.Value != 1 {
					t.Errorf("expected counter value 1, got %d", dp.Value)
				}
				assertAttribute(t, dp.Attributes, "outcome", "success")
			case "azuredevops.webhook.duration":
				foundHist = true
				data := m.Data.(metricdata.Histogram[float64])
				if len(data.DataPoints) == 0 {
					t.Fatal("no data points for histogram")
				}
				if data.DataPoints[0].Count != 1 {
					t.Errorf("expected histogram count 1, got %d", data.DataPoints[0].Count)
				}
			}
		}
	}
	if !foundCounter {
		t.Error("azuredevops.webhook.events counter not found")
	}
	if !foundHist {
		t.Error("azuredevops.webhook.duration histogram not found")
	}
}

func TestWithTelemetry_BothTracingAndMetrics(t *testing.T) {
	traceExporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(traceExporter))
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	handler := webhooks.New(WithTelemetry(tp, mp))
	handler.OnGitPush(func(_ context.Context, _ string, _ *webhooks.GitPushEvent) error {
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(validPayload))
	req.Header.Set("Content-Type", "application/json")

	if err := handler.HandleEventRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(traceExporter.GetSpans()) != 1 {
		t.Errorf("expected 1 span, got %d", len(traceExporter.GetSpans()))
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	hasMetrics := false
	for _, sm := range rm.ScopeMetrics {
		if len(sm.Metrics) > 0 {
			hasMetrics = true
		}
	}
	if !hasMetrics {
		t.Error("expected metrics to be recorded")
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{webhooks.ErrBasicAuthMissing, "auth"},
		{webhooks.ErrBasicAuthFailed, "auth"},
		{webhooks.ErrInvalidHTTPMethod, "parse"},
		{webhooks.ErrParsingPayload, "parse"},
		{webhooks.ErrEmptyBody, "parse"},
		{webhooks.ErrUnknownEventType, "parse"},
		{errors.New("something else"), "handler"},
	}
	for _, tt := range tests {
		got := classifyError(tt.err)
		if got != tt.want {
			t.Errorf("classifyError(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func assertAttribute(t *testing.T, attrs attribute.Set, key, want string) {
	t.Helper()
	v, ok := attrs.Value(attribute.Key(key))
	if !ok {
		t.Errorf("attribute %q not found", key)
		return
	}
	if v.AsString() != want {
		t.Errorf("attribute %q = %q, want %q", key, v.AsString(), want)
	}
}
