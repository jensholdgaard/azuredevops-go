package webhooks

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// OnGitPush registers one or more callbacks for git.push events.
// Callbacks are appended to any previously registered handlers.
func (h *EventHandler) OnGitPush(callbacks ...GitPushEventHandleFunc) {
	h.onGitPush = append(h.onGitPush, callbacks...)
}

// SetOnGitPush replaces all git.push callbacks with the provided ones.
func (h *EventHandler) SetOnGitPush(callbacks ...GitPushEventHandleFunc) {
	h.onGitPush = callbacks
}

// dispatchGitPush runs before hooks, typed handlers (in parallel), after hooks,
// and error hooks as needed. Each callback gets panic recovery.
func (h *EventHandler) dispatchGitPush(ctx context.Context, deliveryID string, event *GitPushEvent) error {
	// Before hooks (sequential)
	for _, fn := range h.onBeforeAny {
		if err := fn(ctx, deliveryID, EventTypeGitPush, event); err != nil {
			return h.handleError(ctx, deliveryID, EventTypeGitPush, event, err)
		}
	}

	// Typed handlers (parallel with panic recovery)
	if len(h.onGitPush) > 0 {
		g, gCtx := errgroup.WithContext(ctx)
		for _, fn := range h.onGitPush {
			fn := fn
			g.Go(func() (retErr error) {
				defer func() {
					if r := recover(); r != nil {
						retErr = fmt.Errorf("panic in git.push handler: %v", r)
					}
				}()
				return fn(gCtx, deliveryID, event)
			})
		}
		if err := g.Wait(); err != nil {
			return h.handleError(ctx, deliveryID, EventTypeGitPush, event, err)
		}
	}

	// After hooks (sequential)
	for _, fn := range h.onAfterAny {
		if err := fn(ctx, deliveryID, EventTypeGitPush, event); err != nil {
			return h.handleError(ctx, deliveryID, EventTypeGitPush, event, err)
		}
	}

	return nil
}

// handleError invokes all registered error callbacks. Returns the original error
// if no error callback overrides it.
func (h *EventHandler) handleError(ctx context.Context, deliveryID string, eventType EventType, event any, err error) error {
	for _, fn := range h.onError {
		if cbErr := fn(ctx, deliveryID, eventType, event, err); cbErr != nil {
			return cbErr
		}
	}
	return err
}
