package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDispatchGitPullRequestCreated(t *testing.T) {
	data, err := os.ReadFile("testdata/git.pullrequest.created.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	handler := New()
	var called atomic.Bool
	var gotEventType EventType
	handler.OnGitPullRequest(func(_ context.Context, deliveryID string, eventType EventType, event *GitPullRequestEvent) error {
		called.Store(true)
		gotEventType = eventType
		if event.Resource.PullRequestID != 42 {
			t.Errorf("pullRequestId = %d, want 42", event.Resource.PullRequestID)
		}
		if deliveryID != "test-delivery" {
			t.Errorf("deliveryID = %q, want %q", deliveryID, "test-delivery")
		}
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(string(data)))
	req.Header.Set("X-Vss-Activityid", "test-delivery")

	if err := handler.HandleEventRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called.Load() {
		t.Fatal("handler was not called")
	}
	if gotEventType != EventTypeGitPullRequestCreated {
		t.Errorf("eventType = %q, want %q", gotEventType, EventTypeGitPullRequestCreated)
	}
}

func TestDispatchGitPullRequestMerged(t *testing.T) {
	data, err := os.ReadFile("testdata/git.pullrequest.merged.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	handler := New()
	var gotEventType EventType
	handler.OnGitPullRequest(func(_ context.Context, _ string, eventType EventType, event *GitPullRequestEvent) error {
		gotEventType = eventType
		if event.Resource.Status != "completed" {
			t.Errorf("status = %q, want %q", event.Resource.Status, "completed")
		}
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(string(data)))
	if err := handler.HandleEventRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEventType != EventTypeGitPullRequestMerged {
		t.Errorf("eventType = %q, want %q", gotEventType, EventTypeGitPullRequestMerged)
	}
}

func TestSetOnGitPullRequest_Replaces(t *testing.T) {
	data, err := os.ReadFile("testdata/git.pullrequest.created.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	handler := New()
	var firstCalled, secondCalled atomic.Bool

	handler.OnGitPullRequest(func(_ context.Context, _ string, _ EventType, _ *GitPullRequestEvent) error {
		firstCalled.Store(true)
		return nil
	})
	handler.SetOnGitPullRequest(func(_ context.Context, _ string, _ EventType, _ *GitPullRequestEvent) error {
		secondCalled.Store(true)
		return nil
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/webhook", strings.NewReader(string(data)))
	if err := handler.HandleEventRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstCalled.Load() {
		t.Error("first handler should not have been called after SetOnGitPullRequest")
	}
	if !secondCalled.Load() {
		t.Error("second handler was not called")
	}
}

func TestServeHTTP_Success(t *testing.T) {
	data, err := os.ReadFile("testdata/git.push.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	handler := New()
	handler.OnGitPush(func(_ context.Context, _ string, _ *GitPushEvent) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(data)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestServeHTTP_BadMethod(t *testing.T) {
	handler := New()

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_AuthFailure(t *testing.T) {
	handler := New(WithBasicAuth("user", "pass"))

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_HandlerError_Returns500(t *testing.T) {
	data, err := os.ReadFile("testdata/git.push.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	handler := New()
	handler.OnGitPush(func(_ context.Context, _ string, _ *GitPushEvent) error {
		return context.DeadlineExceeded
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(data)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
