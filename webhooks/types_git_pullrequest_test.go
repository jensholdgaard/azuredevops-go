package webhooks

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestGitPullRequestCreatedEvent_StrictDecode(t *testing.T) {
	data, err := os.ReadFile("testdata/git.pullrequest.created.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var event GitPullRequestEvent
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&event); err != nil {
		t.Fatalf("strict decode failed: %v", err)
	}

	if event.EventType != EventTypeGitPullRequestCreated {
		t.Errorf("eventType = %q, want %q", event.EventType, EventTypeGitPullRequestCreated)
	}
	if event.Resource.PullRequestID != 42 {
		t.Errorf("pullRequestId = %d, want 42", event.Resource.PullRequestID)
	}
	if event.Resource.Status != "active" {
		t.Errorf("status = %q, want %q", event.Resource.Status, "active")
	}
	if event.Resource.Title != "Add login validation" {
		t.Errorf("title = %q, want %q", event.Resource.Title, "Add login validation")
	}
	if event.Resource.SourceRefName != "refs/heads/feature/login-fix" {
		t.Errorf("sourceRefName = %q, want %q", event.Resource.SourceRefName, "refs/heads/feature/login-fix")
	}
	if event.Resource.TargetRefName != "refs/heads/main" {
		t.Errorf("targetRefName = %q, want %q", event.Resource.TargetRefName, "refs/heads/main")
	}
	if event.Resource.CreatedBy.DisplayName != "Jane Smith" {
		t.Errorf("createdBy.displayName = %q, want %q", event.Resource.CreatedBy.DisplayName, "Jane Smith")
	}
	if len(event.Resource.Reviewers) != 1 {
		t.Fatalf("reviewers count = %d, want 1", len(event.Resource.Reviewers))
	}
	if event.Resource.Reviewers[0].DisplayName != "Bob Jones" {
		t.Errorf("reviewer displayName = %q, want %q", event.Resource.Reviewers[0].DisplayName, "Bob Jones")
	}
	if event.Resource.Repository.Name != "my-application" {
		t.Errorf("repository.name = %q, want %q", event.Resource.Repository.Name, "my-application")
	}
}

func TestGitPullRequestCreatedEvent_RealPayload_StrictDecode(t *testing.T) {
	data, err := os.ReadFile("testdata/git.pullrequest.created.real.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var event GitPullRequestEvent
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&event); err != nil {
		t.Fatalf("strict decode failed: %v", err)
	}

	if event.EventType != EventTypeGitPullRequestCreated {
		t.Errorf("eventType = %q, want %q", event.EventType, EventTypeGitPullRequestCreated)
	}
	if event.Resource.PullRequestID != 1 {
		t.Errorf("pullRequestId = %d, want 1", event.Resource.PullRequestID)
	}
	if event.Resource.Title != "my first pull request" {
		t.Errorf("title = %q, want %q", event.Resource.Title, "my first pull request")
	}
	if event.Resource.CreatedBy.DisplayName != "Jamal Hartnett" {
		t.Errorf("createdBy.displayName = %q, want %q", event.Resource.CreatedBy.DisplayName, "Jamal Hartnett")
	}
	if len(event.Resource.Reviewers) != 1 {
		t.Fatalf("reviewers count = %d, want 1", len(event.Resource.Reviewers))
	}
	if !event.Resource.Reviewers[0].IsContainer {
		t.Error("expected reviewer to be a container (team)")
	}
	if event.Resource.Repository.Project.LastUpdateTime != "0001-01-01T00:00:00" {
		t.Errorf("lastUpdateTime = %q, want %q", event.Resource.Repository.Project.LastUpdateTime, "0001-01-01T00:00:00")
	}
	if len(event.Resource.Commits) != 1 {
		t.Fatalf("commits count = %d, want 1", len(event.Resource.Commits))
	}
	if event.Resource.Links["web"].Href == "" {
		t.Error("expected _links.web.href to be set")
	}
}

func TestGitPullRequestMergedEvent_StrictDecode(t *testing.T) {
	data, err := os.ReadFile("testdata/git.pullrequest.merged.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var event GitPullRequestEvent
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&event); err != nil {
		t.Fatalf("strict decode failed: %v", err)
	}

	if event.EventType != EventTypeGitPullRequestMerged {
		t.Errorf("eventType = %q, want %q", event.EventType, EventTypeGitPullRequestMerged)
	}
	if event.Resource.Status != "completed" {
		t.Errorf("status = %q, want %q", event.Resource.Status, "completed")
	}
	if event.Resource.ClosedDate != "2024-01-15T12:00:00Z" {
		t.Errorf("closedDate = %q, want %q", event.Resource.ClosedDate, "2024-01-15T12:00:00Z")
	}
	if len(event.Resource.Commits) != 1 {
		t.Fatalf("commits count = %d, want 1", len(event.Resource.Commits))
	}
	if event.Resource.Links == nil {
		t.Fatal("expected _links to be present")
	}
	if event.Resource.Links["web"].Href == "" {
		t.Error("expected _links.web.href to be set")
	}
}
