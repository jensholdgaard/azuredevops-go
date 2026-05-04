package webhooks

import (
	"crypto/subtle"
	"net/http"
)

// validateBasicAuth checks the request's basic auth credentials against the
// configured username and password using constant-time comparison to prevent
// timing attacks.
func validateBasicAuth(r *http.Request, username, password string) error {
	u, p, ok := r.BasicAuth()
	if !ok {
		return ErrBasicAuthMissing
	}

	usernameMatch := subtle.ConstantTimeCompare([]byte(u), []byte(username))
	passwordMatch := subtle.ConstantTimeCompare([]byte(p), []byte(password))

	if usernameMatch&passwordMatch != 1 {
		return ErrBasicAuthFailed
	}

	return nil
}
