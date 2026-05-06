package webhooks

import (
	"context"
	"errors"
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
	onGitPush        []GitPushEventHandleFunc
	onGitPullRequest []GitPullRequestEventHandleFunc
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

	// Start observers, propagating context through each
	ends := make([]func(EventType, string, any, error), 0, len(h.observers))
	for _, obs := range h.observers {
		var end func(EventType, string, any, error)
		ctx, end = obs.ObserveRequest(ctx, r)
		if end != nil {
			ends = append(ends, end)
		}
		r = r.WithContext(ctx)
	}

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
	// Extract deliveryID early so observers always get it, even on auth/parse failures.
	deliveryID := r.Header.Get(headerActivityID)

	if h.username != "" || h.password != "" {
		if err := validateBasicAuth(r, h.username, h.password); err != nil {
			return "", deliveryID, nil, err
		}
	}

	event, parsedDeliveryID, err := parse(r)
	if parsedDeliveryID != "" {
		deliveryID = parsedDeliveryID
	}
	if err != nil {
		return "", deliveryID, nil, err
	}

	var eventType EventType
	switch e := event.(type) {
	case *GitPushEvent:
		eventType = EventTypeGitPush
	case *GitPullRequestEvent:
		eventType = e.EventType
	}

	return eventType, deliveryID, event, h.dispatch(r.Context(), deliveryID, eventType, event)
}

// ServeHTTP implements http.Handler, allowing EventHandler to be mounted directly.
// Returns 200 on success, 400 on client errors (bad payload, auth failure),
// and 500 on handler errors.
func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.HandleEventRequest(r); err != nil {
		switch {
		case isClientError(err):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func isClientError(err error) bool {
	return errors.Is(err, ErrInvalidHTTPMethod) ||
		errors.Is(err, ErrParsingPayload) ||
		errors.Is(err, ErrEmptyBody) ||
		errors.Is(err, ErrUnknownEventType) ||
		errors.Is(err, ErrBasicAuthMissing) ||
		errors.Is(err, ErrBasicAuthFailed)
}
func (h *EventHandler) dispatch(ctx context.Context, deliveryID string, eventType EventType, event any) error {
	switch e := event.(type) {
	case *GitPushEvent:
		return h.dispatchGitPush(ctx, deliveryID, e)
	case *GitPullRequestEvent:
		return h.dispatchGitPullRequest(ctx, deliveryID, eventType, e)
	default:
		return ErrUnknownEventType
	}
}
