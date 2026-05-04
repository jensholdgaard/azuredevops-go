package webhooks

// GitPushEvent represents a git.push webhook event from Azure DevOps.
//
// See: https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops#git.push
type GitPushEvent struct {
	SubscriptionID     string             `json:"subscriptionId"`
	NotificationID     int                `json:"notificationId"`
	ID                 string             `json:"id"`
	EventType          EventType          `json:"eventType"`
	PublisherID        string             `json:"publisherId"`
	Message            Message            `json:"message"`
	DetailedMessage    Message            `json:"detailedMessage"`
	Resource           GitPushResource    `json:"resource"`
	ResourceVersion    string             `json:"resourceVersion"`
	ResourceContainers ResourceContainers `json:"resourceContainers"`
	CreatedDate        string             `json:"createdDate"`
}

// GitPushResource contains the push details.
type GitPushResource struct {
	Commits    []Commit    `json:"commits"`
	RefUpdates []RefUpdate `json:"refUpdates"`
	Repository Repository  `json:"repository"`
	PushedBy   Identity    `json:"pushedBy"`
	PushID     int         `json:"pushId"`
	Date       string      `json:"date"`
	URL        string      `json:"url"`
}

// Commit represents a single commit in a push event.
type Commit struct {
	CommitID  string      `json:"commitId"`
	Author    GitUserDate `json:"author"`
	Committer GitUserDate `json:"committer"`
	Comment   string      `json:"comment"`
	URL       string      `json:"url"`
}

// GitUserDate represents a git user with a timestamp.
type GitUserDate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// RefUpdate represents a reference update in a push event.
type RefUpdate struct {
	Name        string `json:"name"`
	OldObjectID string `json:"oldObjectId"`
	NewObjectID string `json:"newObjectId"`
}
