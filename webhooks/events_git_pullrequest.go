package webhooks

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// GitPullRequestEventHandleFunc is a callback for git.pullrequest.* events.
type GitPullRequestEventHandleFunc func(ctx context.Context, deliveryID string, eventType EventType, event *GitPullRequestEvent) error

// OnGitPullRequest registers callbacks for all git.pullrequest.* events
// (created, updated, merged).
func (h *EventHandler) OnGitPullRequest(callbacks ...GitPullRequestEventHandleFunc) {
	h.onGitPullRequest = append(h.onGitPullRequest, callbacks...)
}

// SetOnGitPullRequest replaces all git.pullrequest.* callbacks.
func (h *EventHandler) SetOnGitPullRequest(callbacks ...GitPullRequestEventHandleFunc) {
	h.onGitPullRequest = callbacks
}

func (h *EventHandler) dispatchGitPullRequest(ctx context.Context, deliveryID string, eventType EventType, event *GitPullRequestEvent) error {
	for _, fn := range h.onBeforeAny {
		if err := fn(ctx, deliveryID, eventType, event); err != nil {
			return h.handleError(ctx, deliveryID, eventType, event, err)
		}
	}

	if len(h.onGitPullRequest) > 0 {
		g, gCtx := errgroup.WithContext(ctx)
		for _, fn := range h.onGitPullRequest {
			fn := fn
			g.Go(func() (retErr error) {
				defer func() {
					if r := recover(); r != nil {
						retErr = fmt.Errorf("panic in %s handler: %v", eventType, r)
					}
				}()
				return fn(gCtx, deliveryID, eventType, event)
			})
		}
		if err := g.Wait(); err != nil {
			return h.handleError(ctx, deliveryID, eventType, event, err)
		}
	}

	for _, fn := range h.onAfterAny {
		if err := fn(ctx, deliveryID, eventType, event); err != nil {
			return h.handleError(ctx, deliveryID, eventType, event, err)
		}
	}

	return nil
}
