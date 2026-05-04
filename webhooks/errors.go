package webhooks

import "errors"

var (
	// ErrInvalidHTTPMethod is returned when the request method is not POST.
	ErrInvalidHTTPMethod = errors.New("invalid HTTP method: only POST is accepted")

	// ErrParsingPayload is returned when the request body cannot be parsed as JSON.
	ErrParsingPayload = errors.New("failed to parse webhook payload")

	// ErrUnknownEventType is returned when the eventType field is not recognized.
	ErrUnknownEventType = errors.New("unknown event type")

	// ErrBasicAuthMissing is returned when basic auth is configured but not provided.
	ErrBasicAuthMissing = errors.New("basic auth credentials missing from request")

	// ErrBasicAuthFailed is returned when the provided credentials don't match.
	ErrBasicAuthFailed = errors.New("basic auth credentials invalid")

	// ErrEmptyBody is returned when the request body is empty.
	ErrEmptyBody = errors.New("request body is empty")
)
