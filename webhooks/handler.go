package webhooks

import (
	"context"
	"net/http"
)

// GitPushEventHandleFunc is a callback for git.push events.
type GitPushEventHandleFunc func(ctx context.Context, deliveryID string, event *GitPushEvent) error

// EventHandleFunc is a generic callback for any event (used for before/after hooks).
type EventHandleFunc func(ctx context.Context, deliveryID string, eventType EventType, event any) error

// ErrorEventHandleFunc is called when an event handler returns an error.
type ErrorEventHandleFunc func(ctx context.Context, deliveryID string, eventType EventType, event any, err error) error

// EventHandler manages webhook event callbacks and dispatches incoming requests.
type EventHandler struct {
	// Auth
	username string
	password string

	// Hooks
	onBeforeAny []EventHandleFunc
	onAfterAny  []EventHandleFunc
	onError     []ErrorEventHandleFunc

	// Observers
	observers []RequestObserver

	// Event handlers
	onGitPush []GitPushEventHandleFunc
}

// Option configures an EventHandler.
type Option func(*EventHandler)

// WithBasicAuth configures basic auth validation for incoming requests.
func WithBasicAuth(username, password string) Option {
	return func(h *EventHandler) {
		h.username = username
		h.password = password
	}
}

// New creates an EventHandler with the given options.
func New(opts ...Option) *EventHandler {
	h := &EventHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// OnBeforeAny registers callbacks that run before the typed event handler.
func (h *EventHandler) OnBeforeAny(callbacks ...EventHandleFunc) {
	h.onBeforeAny = append(h.onBeforeAny, callbacks...)
}

// OnAfterAny registers callbacks that run after the typed event handler.
func (h *EventHandler) OnAfterAny(callbacks ...EventHandleFunc) {
	h.onAfterAny = append(h.onAfterAny, callbacks...)
}

// OnError registers callbacks that are invoked when a handler returns an error.
func (h *EventHandler) OnError(callbacks ...ErrorEventHandleFunc) {
	h.onError = append(h.onError, callbacks...)
}

// HandleEventRequest processes an incoming HTTP webhook request.
// It validates auth, parses the payload, and dispatches to registered callbacks.
func (h *EventHandler) HandleEventRequest(r *http.Request) error {
	ctx := r.Context()

	// Start observers
	ends := make([]func(EventType, string, any, error), 0, len(h.observers))
	for _, obs := range h.observers {
		var end func(EventType, string, any, error)
		ctx, end = obs.ObserveRequest(ctx, r)
		ends = append(ends, end)
	}
	r = r.WithContext(ctx)

	// Process request
	eventType, deliveryID, event, err := h.processRequest(r)

	// End observers (reverse order)
	for i := len(ends) - 1; i >= 0; i-- {
		ends[i](eventType, deliveryID, event, err)
	}
	return err
}

// processRequest contains the core auth → parse → dispatch logic.
func (h *EventHandler) processRequest(r *http.Request) (EventType, string, any, error) {
	if h.username != "" || h.password != "" {
		if err := validateBasicAuth(r, h.username, h.password); err != nil {
			return "", "", nil, err
		}
	}

	event, deliveryID, err := parse(r)
	if err != nil {
		return "", deliveryID, nil, err
	}

	var eventType EventType
	switch event.(type) {
	case *GitPushEvent:
		eventType = EventTypeGitPush
	}

	return eventType, deliveryID, event, h.dispatch(r.Context(), deliveryID, event)
}

// dispatch routes the parsed event to the appropriate typed handler.
func (h *EventHandler) dispatch(ctx context.Context, deliveryID string, event any) error {
	switch e := event.(type) {
	case *GitPushEvent:
		return h.dispatchGitPush(ctx, deliveryID, e)
	default:
		return ErrUnknownEventType
	}
}
