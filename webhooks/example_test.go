package webhooks_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/jensholdgaard/azuredevops-go/webhooks"
)

func Example() {
	handler := webhooks.New(webhooks.WithBasicAuth("webhook", "secret"))

	handler.OnGitPush(func(ctx context.Context, deliveryID string, event *webhooks.GitPushEvent) error {
		fmt.Printf("Push to %s/%s by %s\n",
			event.Resource.Repository.Project.Name,
			event.Resource.Repository.Name,
			event.Resource.PushedBy.DisplayName,
		)
		return nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if err := handler.HandleEventRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Simulate a request
	payload := `{
		"eventType": "git.push",
		"publisherId": "tfs",
		"message": {"text": "pushed"},
		"detailedMessage": {"text": "pushed"},
		"resource": {
			"commits": [{"commitId": "abc123", "author": {"name": "Jane", "email": "j@ex.com", "date": "2024-01-01T00:00:00Z"}, "committer": {"name": "Jane", "email": "j@ex.com", "date": "2024-01-01T00:00:00Z"}, "comment": "fix", "url": "https://example.com"}],
			"refUpdates": [{"name": "refs/heads/main", "oldObjectId": "000", "newObjectId": "abc"}],
			"repository": {"id": "1", "name": "my-app", "url": "https://example.com", "project": {"id": "2", "name": "MyProject", "state": "wellFormed", "visibility": "private"}, "defaultBranch": "refs/heads/main", "remoteUrl": "https://example.com"},
			"pushedBy": {"displayName": "Jane Smith", "uniqueName": "jane@ex.com", "id": "3"},
			"pushId": 1, "date": "2024-01-01T00:00:00Z", "url": "https://example.com"
		},
		"subscriptionId": "sub1", "notificationId": 1, "id": "evt1",
		"resourceVersion": "1.0",
		"resourceContainers": {"collection": {"id": "c1"}, "account": {"id": "a1"}, "project": {"id": "p1"}},
		"createdDate": "2024-01-01T00:00:00Z"
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(payload))
	req.SetBasicAuth("webhook", "secret")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Output: Push to MyProject/my-app by Jane Smith
}
