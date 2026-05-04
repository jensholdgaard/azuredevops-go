package webhooks

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestEventHandler_HandleEventRequest_GitPush(t *testing.T) {
	var called atomic.Bool
	h := New()
	h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		called.Store(true)
		if event.Resource.Repository.Name != "my-application" {
			t.Errorf("expected repo 'my-application', got %q", event.Resource.Repository.Name)
		}
		if deliveryID != "delivery-abc" {
			t.Errorf("expected deliveryID 'delivery-abc', got %q", deliveryID)
		}
		return nil
	})

	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.Header.Set(headerActivityID, "delivery-abc")

	if err := h.HandleEventRequest(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called.Load() {
		t.Error("handler was not called")
	}
}

func TestEventHandler_BasicAuth_Required(t *testing.T) {
	h := New(WithBasicAuth("admin", "secret"))
	h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		return nil
	})

	body := loadTestdata(t, "git.push.json")

	// No auth
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	err := h.HandleEventRequest(r)
	if err != ErrBasicAuthMissing {
		t.Fatalf("expected ErrBasicAuthMissing, got: %v", err)
	}

	// Wrong auth
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.SetBasicAuth("admin", "wrong")
	err = h.HandleEventRequest(r)
	if err != ErrBasicAuthFailed {
		t.Fatalf("expected ErrBasicAuthFailed, got: %v", err)
	}

	// Correct auth
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.SetBasicAuth("admin", "secret")
	err = h.HandleEventRequest(r)
	if err != nil {
		t.Fatalf("unexpected error with correct auth: %v", err)
	}
}

func TestEventHandler_BeforeAfterHooks(t *testing.T) {
	var order []string

	h := New()
	h.OnBeforeAny(func(ctx context.Context, deliveryID string, eventType EventType, event any) error {
		order = append(order, "before")
		return nil
	})
	h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		order = append(order, "handler")
		return nil
	})
	h.OnAfterAny(func(ctx context.Context, deliveryID string, eventType EventType, event any) error {
		order = append(order, "after")
		return nil
	})

	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	if err := h.HandleEventRequest(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"before", "handler", "after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], order[i])
		}
	}
}

func TestEventHandler_ErrorHook(t *testing.T) {
	handlerErr := errors.New("handler failed")
	var errorHookCalled bool

	h := New()
	h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		return handlerErr
	})
	h.OnError(func(ctx context.Context, deliveryID string, eventType EventType, event any, err error) error {
		errorHookCalled = true
		if !errors.Is(err, handlerErr) {
			t.Errorf("expected handlerErr, got: %v", err)
		}
		return err
	})

	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	err := h.HandleEventRequest(r)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handlerErr, got: %v", err)
	}
	if !errorHookCalled {
		t.Error("error hook was not called")
	}
}

func TestEventHandler_PanicRecovery(t *testing.T) {
	h := New()
	h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		panic("something went wrong")
	})

	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))

	err := h.HandleEventRequest(r)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("expected panic-related error, got: %v", err)
	}
}

func TestEventHandler_ParallelCallbacks(t *testing.T) {
	var count atomic.Int32

	h := New()
	for i := 0; i < 5; i++ {
		h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
			count.Add(1)
			return nil
		})
	}

	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	if err := h.HandleEventRequest(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count.Load() != 5 {
		t.Errorf("expected 5 callbacks to run, got %d", count.Load())
	}
}

func TestEventHandler_SetOnGitPush_Replaces(t *testing.T) {
	var called string

	h := New()
	h.OnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		called = "first"
		return nil
	})
	h.SetOnGitPush(func(ctx context.Context, deliveryID string, event *GitPushEvent) error {
		called = "second"
		return nil
	})

	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	if err := h.HandleEventRequest(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called != "second" {
		t.Errorf("expected 'second', got %q", called)
	}
}
