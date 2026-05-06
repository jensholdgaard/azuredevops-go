package webhooks

import (
	"context"
	"net/http"
)

// RequestObserver observes the full HandleEventRequest lifecycle without
// coupling to any specific telemetry library.
type RequestObserver interface {
	// ObserveRequest is called at the start of HandleEventRequest.
	// It returns a (possibly enriched) context and an end function.
	// The end function is called when processing completes with the outcome.
	ObserveRequest(ctx context.Context, r *http.Request) (context.Context, func(eventType EventType, deliveryID string, event any, err error))
}

// WithRequestObserver registers observers for the request lifecycle.
func WithRequestObserver(obs ...RequestObserver) Option {
	return func(h *EventHandler) {
		h.observers = append(h.observers, obs...)
	}
}
