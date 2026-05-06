package webhooks

// GitPullRequestEvent represents a git.pullrequest.* webhook event from Azure DevOps.
// The same struct is used for created, updated, and merged events — the EventType
// field distinguishes them.
//
// See: https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops
type GitPullRequestEvent struct {
	SubscriptionID     string                 `json:"subscriptionId,omitempty"`
	NotificationID     int                    `json:"notificationId,omitempty"`
	ID                 string                 `json:"id"`
	EventType          EventType              `json:"eventType"`
	PublisherID        string                 `json:"publisherId"`
	Scope              string                 `json:"scope,omitempty"`
	Message            Message                `json:"message"`
	DetailedMessage    Message                `json:"detailedMessage"`
	Resource           GitPullRequestResource `json:"resource"`
	ResourceVersion    string                 `json:"resourceVersion"`
	ResourceContainers ResourceContainers     `json:"resourceContainers"`
	CreatedDate        string                 `json:"createdDate"`
}

// GitPullRequestResource contains the pull request details.
type GitPullRequestResource struct {
	Repository            Repository      `json:"repository"`
	PullRequestID         int             `json:"pullRequestId"`
	Status                string          `json:"status"`
	CreatedBy             Identity        `json:"createdBy"`
	CreationDate          string          `json:"creationDate"`
	ClosedDate            string          `json:"closedDate,omitempty"`
	Title                 string          `json:"title"`
	Description           string          `json:"description,omitempty"`
	SourceRefName         string          `json:"sourceRefName"`
	TargetRefName         string          `json:"targetRefName"`
	MergeStatus           string          `json:"mergeStatus"`
	MergeID               string          `json:"mergeId"`
	LastMergeSourceCommit CommitRef       `json:"lastMergeSourceCommit"`
	LastMergeTargetCommit CommitRef       `json:"lastMergeTargetCommit"`
	LastMergeCommit       CommitRef       `json:"lastMergeCommit"`
	Reviewers             []Reviewer      `json:"reviewers"`
	Commits               []CommitRef     `json:"commits,omitempty"`
	URL                   string          `json:"url"`
	Links                 map[string]Link `json:"_links,omitempty"`
}

// CommitRef is a lightweight commit reference (ID + URL).
type CommitRef struct {
	CommitID string `json:"commitId"`
	URL      string `json:"url"`
}

// Reviewer represents a pull request reviewer.
type Reviewer struct {
	ReviewerURL string `json:"reviewerUrl"`
	Vote        int    `json:"vote"`
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	URL         string `json:"url"`
	ImageURL    string `json:"imageUrl,omitempty"`
	IsContainer bool   `json:"isContainer,omitempty"`
}

// Link is a HAL-style link.
type Link struct {
	Href string `json:"href"`
}
