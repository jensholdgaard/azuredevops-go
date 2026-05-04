package auth

import (
	"context"
	"net/http"
	"testing"
)

func TestPAT_SetsBasicAuthHeader(t *testing.T) {
	provider := PAT("my-secret-token")

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://dev.azure.com/org/_apis/projects", nil)
	if err := provider.SetAuthHeaders(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Fatal("expected Authorization header to be set")
	}

	// ADO PATs use Basic auth with empty username: base64(":token")
	expected := "Basic Om15LXNlY3JldC10b2tlbg=="
	if auth != expected {
		t.Errorf("expected %q, got %q", expected, auth)
	}
}

func TestPAT_EmptyUsername(t *testing.T) {
	// Verify the PAT provider encodes with empty username prefix
	provider := PAT("abc123")

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	if err := provider.SetAuthHeaders(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode and verify format is ":token"
	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Fatal("Authorization header missing")
	}
	// "Basic " + base64(":abc123") = "Basic OmFiYzEyMw=="
	expected := "Basic OmFiYzEyMw=="
	if auth != expected {
		t.Errorf("expected %q, got %q", expected, auth)
	}
}

func TestServicePrincipal_RequiresAllFields(t *testing.T) {
	// ServicePrincipal returns a provider that will fail on token fetch
	// if credentials are invalid. We can at least verify it constructs without panic.
	provider := ServicePrincipal(ServicePrincipalConfig{
		TenantID:     "tenant-id",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestServicePrincipal_TokenFetchFails_WithBadCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network-dependent test in short mode")
	}

	provider := ServicePrincipal(ServicePrincipalConfig{
		TenantID:     "fake-tenant",
		ClientID:     "fake-client",
		ClientSecret: "fake-secret",
	})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)

	// This will fail because the credentials are invalid, but it should return
	// an error rather than panic.
	err := provider.SetAuthHeaders(context.Background(), req)
	if err == nil {
		t.Fatal("expected error with fake credentials")
	}
}
