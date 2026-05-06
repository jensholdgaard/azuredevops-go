package webhooks

// EventType identifies an Azure DevOps Service Hook event.
type EventType string

const (
	// EventTypeGitPush is fired when commits are pushed to a repository.
	EventTypeGitPush EventType = "git.push"

	// EventTypeGitPullRequestCreated is fired when a pull request is created.
	EventTypeGitPullRequestCreated EventType = "git.pullrequest.created"

	// EventTypeGitPullRequestUpdated is fired when a pull request is updated.
	EventTypeGitPullRequestUpdated EventType = "git.pullrequest.updated"

	// EventTypeGitPullRequestMerged is fired when a pull request merge commit is created.
	EventTypeGitPullRequestMerged EventType = "git.pullrequest.merged"
)

// baseEvent is used internally to peek at eventType before full deserialization.
type baseEvent struct {
	EventType EventType `json:"eventType"`
}

// Message represents a notification message from Azure DevOps.
type Message struct {
	Text     string `json:"text"`
	HTML     string `json:"html,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// ResourceContainers holds references to the collection, account, and project.
type ResourceContainers struct {
	Collection ResourceContainer `json:"collection"`
	Account    ResourceContainer `json:"account"`
	Project    ResourceContainer `json:"project"`
}

// ResourceContainer is a reference to an Azure DevOps resource container.
type ResourceContainer struct {
	ID      string `json:"id"`
	BaseURL string `json:"baseUrl,omitempty"`
}

// Identity represents an Azure DevOps user identity.
type Identity struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	ID          string `json:"id"`
	URL         string `json:"url,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
}

// Repository represents an Azure DevOps Git repository.
type Repository struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Project       Project `json:"project"`
	DefaultBranch string  `json:"defaultBranch"`
	RemoteURL     string  `json:"remoteUrl"`
}

// Project represents an Azure DevOps project.
type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
	State      string `json:"state"`
	Visibility string `json:"visibility,omitempty"`
}
