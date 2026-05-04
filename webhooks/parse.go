package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// maxBodySize is the maximum webhook payload size (10 MB).
	maxBodySize = 10 * 1024 * 1024

	// headerActivityID is the ADO correlation header (equivalent to GitHub's X-GitHub-Delivery).
	headerActivityID = "X-Vss-Activityid"
)

// parse reads and validates the HTTP request, returning the parsed event and delivery ID.
func parse(r *http.Request) (any, string, error) {
	if r.Method != http.MethodPost {
		return nil, "", ErrInvalidHTTPMethod
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrParsingPayload, err)
	}
	if len(body) == 0 {
		return nil, "", ErrEmptyBody
	}

	// First pass: peek at eventType with lenient decode.
	var base baseEvent
	if err := json.Unmarshal(body, &base); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrParsingPayload, err)
	}

	// Second pass: strict decode into typed struct.
	event, err := decodeTyped(base.EventType, body)
	if err != nil {
		return nil, "", err
	}

	deliveryID := r.Header.Get(headerActivityID)

	return event, deliveryID, nil
}

// decodeTyped performs strict JSON decoding into the appropriate typed struct
// based on the event type. DisallowUnknownFields ensures we catch API drift.
func decodeTyped(eventType EventType, body []byte) (any, error) {
	switch eventType {
	case EventTypeGitPush:
		var event GitPushEvent
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&event); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrParsingPayload, err)
		}
		return &event, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownEventType, eventType)
	}
}
