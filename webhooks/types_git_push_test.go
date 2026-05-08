package webhooks

import (
	"bytes"
	"encoding/json"
	"testing"
)

// fullGitPushEvent is a struct covering ALL fields in the ADO git.push webhook payload.
// Used with DisallowUnknownFields to detect API drift — if ADO adds a field to the
// response that we don't have here, the test fails.
type fullGitPushEvent struct {
	SubscriptionID string `json:"subscriptionId"`
	NotificationID int    `json:"notificationId"`
	ID             string `json:"id"`
	EventType      string `json:"eventType"`
	PublisherID    string `json:"publisherId"`
	Message        struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"message"`
	DetailedMessage struct {
		Text string `json:"text"`
	} `json:"detailedMessage"`
	Resource struct {
		Commits []struct {
			CommitID string `json:"commitId"`
			Author   struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Date  string `json:"date"`
			} `json:"author"`
			Committer struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Date  string `json:"date"`
			} `json:"committer"`
			Comment string `json:"comment"`
			URL     string `json:"url"`
		} `json:"commits"`
		RefUpdates []struct {
			Name        string `json:"name"`
			OldObjectID string `json:"oldObjectId"`
			NewObjectID string `json:"newObjectId"`
		} `json:"refUpdates"`
		Repository struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Project struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				URL        string `json:"url"`
				State      string `json:"state"`
				Visibility string `json:"visibility"`
			} `json:"project"`
			DefaultBranch string `json:"defaultBranch"`
			RemoteURL     string `json:"remoteUrl"`
		} `json:"repository"`
		PushedBy struct {
			DisplayName string `json:"displayName"`
			UniqueName  string `json:"uniqueName"`
			ID          string `json:"id"`
			URL         string `json:"url"`
			ImageURL    string `json:"imageUrl"`
		} `json:"pushedBy"`
		PushID int    `json:"pushId"`
		Date   string `json:"date"`
		URL    string `json:"url"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"collection"`
		Account struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"account"`
		Project struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate string `json:"createdDate"`
}

func TestContract_GitPush_FullPayload(t *testing.T) {
	data := loadTestdata(t, "git.push.json")

	var event fullGitPushEvent
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&event); err != nil {
		t.Fatalf("strict unmarshal failed (unknown field in fixture?): %v", err)
	}

	if event.EventType != "git.push" {
		t.Errorf("expected eventType 'git.push', got %q", event.EventType)
	}
	if event.Resource.Repository.Name == "" {
		t.Error("resource.repository.name must not be empty")
	}
	if event.Resource.Repository.Project.Name == "" {
		t.Error("resource.repository.project.name must not be empty")
	}
	if len(event.Resource.Commits) == 0 {
		t.Error("resource.commits must not be empty")
	}
	if len(event.Resource.RefUpdates) == 0 {
		t.Error("resource.refUpdates must not be empty")
	}
	if event.Resource.PushedBy.DisplayName == "" {
		t.Error("resource.pushedBy.displayName must not be empty")
	}
}

func TestContract_GitPush_MinimalPayload(t *testing.T) {
	data := loadTestdata(t, "git.push.minimal.json")

	// Minimal fixture omits optional fields (html, markdown, baseUrl, url, imageUrl).
	// Standard unmarshal should still work fine for our typed structs.
	var event GitPushEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to unmarshal minimal payload: %v", err)
	}

	if event.EventType != EventTypeGitPush {
		t.Errorf("expected eventType %q, got %q", EventTypeGitPush, event.EventType)
	}
	if event.Resource.Repository.Name != "my-application" {
		t.Errorf("expected repo 'my-application', got %q", event.Resource.Repository.Name)
	}
}

func TestContract_GitPush_RealPayload(t *testing.T) {
	data := loadTestdata(t, "git.push.real.json")

	var event GitPushEvent
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&event); err != nil {
		t.Fatalf("strict decode of real ADO payload failed: %v", err)
	}

	if event.EventType != EventTypeGitPush {
		t.Errorf("eventType = %q, want %q", event.EventType, EventTypeGitPush)
	}
	if event.Resource.Repository.Name != "Fabrikam-Fiber-Git" {
		t.Errorf("repository.name = %q, want %q", event.Resource.Repository.Name, "Fabrikam-Fiber-Git")
	}
	if event.Resource.Repository.Project.LastUpdateTime != "0001-01-01T00:00:00" {
		t.Errorf("lastUpdateTime = %q, want %q", event.Resource.Repository.Project.LastUpdateTime, "0001-01-01T00:00:00")
	}
	if event.Resource.PushedBy.DisplayName != "Jamal Hartnett" {
		t.Errorf("pushedBy.displayName = %q, want %q", event.Resource.PushedBy.DisplayName, "Jamal Hartnett")
	}
	if len(event.Resource.Commits) != 1 {
		t.Fatalf("commits count = %d, want 1", len(event.Resource.Commits))
	}
	if event.Resource.Commits[0].Comment != "Fixed bug in web.config file" {
		t.Errorf("commit comment = %q, want %q", event.Resource.Commits[0].Comment, "Fixed bug in web.config file")
	}
	if event.Resource.PushID != 14 {
		t.Errorf("pushId = %d, want 14", event.Resource.PushID)
	}
}

func TestContract_GitPush_TypedDecode(t *testing.T) {
	data := loadTestdata(t, "git.push.json")

	// Verify the full fixture decodes through our strict two-pass decoder.
	event, err := decodeTyped(EventTypeGitPush, data)
	if err != nil {
		t.Fatalf("decodeTyped failed: %v", err)
	}

	pushEvent, ok := event.(*GitPushEvent)
	if !ok {
		t.Fatalf("expected *GitPushEvent, got %T", event)
	}

	if pushEvent.Resource.Repository.Project.Name != "MyProject" {
		t.Errorf("expected project 'MyProject', got %q", pushEvent.Resource.Repository.Project.Name)
	}
	if pushEvent.Resource.Commits[0].Author.Email != "jane.smith@example.com" {
		t.Errorf("expected author email 'jane.smith@example.com', got %q", pushEvent.Resource.Commits[0].Author.Email)
	}
}
