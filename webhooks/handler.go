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
	if h.username != "" || h.password != "" {
		if err := validateBasicAuth(r, h.username, h.password); err != nil {
			return err
		}
	}

	event, deliveryID, err := parse(r)
	if err != nil {
		return err
	}

	return h.dispatch(r.Context(), deliveryID, event)
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
