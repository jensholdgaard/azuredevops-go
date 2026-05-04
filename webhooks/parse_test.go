package webhooks

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_ValidGitPush(t *testing.T) {
	body := loadTestdata(t, "git.push.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.Header.Set(headerActivityID, "test-delivery-123")

	event, deliveryID, err := parse(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deliveryID != "test-delivery-123" {
		t.Errorf("expected deliveryID 'test-delivery-123', got %q", deliveryID)
	}

	pushEvent, ok := event.(*GitPushEvent)
	if !ok {
		t.Fatalf("expected *GitPushEvent, got %T", event)
	}

	if pushEvent.EventType != EventTypeGitPush {
		t.Errorf("expected eventType %q, got %q", EventTypeGitPush, pushEvent.EventType)
	}
	if pushEvent.Resource.Repository.Name != "my-application" {
		t.Errorf("expected repo name 'my-application', got %q", pushEvent.Resource.Repository.Name)
	}
}

func TestParse_MinimalGitPush(t *testing.T) {
	body := loadTestdata(t, "git.push.minimal.json")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))

	event, _, err := parse(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pushEvent, ok := event.(*GitPushEvent)
	if !ok {
		t.Fatalf("expected *GitPushEvent, got %T", event)
	}

	if len(pushEvent.Resource.Commits) != 1 {
		t.Errorf("expected 1 commit, got %d", len(pushEvent.Resource.Commits))
	}
}

func TestParse_InvalidMethod(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	_, _, err := parse(r)
	if err != ErrInvalidHTTPMethod {
		t.Fatalf("expected ErrInvalidHTTPMethod, got: %v", err)
	}
}

func TestParse_EmptyBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	_, _, err := parse(r)
	if err != ErrEmptyBody {
		t.Fatalf("expected ErrEmptyBody, got: %v", err)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))

	_, _, err := parse(r)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("expected parsing error, got: %v", err)
	}
}

func TestParse_UnknownEventType(t *testing.T) {
	payload := `{"eventType": "build.complete"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))

	_, _, err := parse(r)
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
	if !strings.Contains(err.Error(), "unknown event type") {
		t.Errorf("expected unknown event type error, got: %v", err)
	}
}

func loadTestdata(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", filename, err)
	}
	return data
}
