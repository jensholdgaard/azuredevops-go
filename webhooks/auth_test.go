package webhooks

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateBasicAuth_Success(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.SetBasicAuth("user", "pass")

	err := validateBasicAuth(r, "user", "pass")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateBasicAuth_Missing(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	err := validateBasicAuth(r, "user", "pass")
	if err != ErrBasicAuthMissing {
		t.Fatalf("expected ErrBasicAuthMissing, got: %v", err)
	}
}

func TestValidateBasicAuth_WrongUsername(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.SetBasicAuth("wrong", "pass")

	err := validateBasicAuth(r, "user", "pass")
	if err != ErrBasicAuthFailed {
		t.Fatalf("expected ErrBasicAuthFailed, got: %v", err)
	}
}

func TestValidateBasicAuth_WrongPassword(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.SetBasicAuth("user", "wrong")

	err := validateBasicAuth(r, "user", "pass")
	if err != ErrBasicAuthFailed {
		t.Fatalf("expected ErrBasicAuthFailed, got: %v", err)
	}
}
