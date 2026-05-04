package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Provider sets authentication headers on outbound HTTP requests to Azure DevOps APIs.
type Provider interface {
	// SetAuthHeaders adds the appropriate Authorization header to the request.
	SetAuthHeaders(ctx context.Context, req *http.Request) error
}

// PAT returns a Provider that authenticates using a Personal Access Token.
// ADO PATs use Basic auth with an empty username and the token as password.
func PAT(token string) Provider {
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + token))
	return &patProvider{encoded: encoded}
}

type patProvider struct {
	encoded string
}

func (p *patProvider) SetAuthHeaders(_ context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Basic "+p.encoded)
	return nil
}

// ServicePrincipalConfig holds the credentials for Azure AD service principal authentication.
type ServicePrincipalConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

// ServicePrincipal returns a Provider that authenticates using OAuth2 client credentials
// (Azure AD service principal). Tokens are cached and refreshed automatically.
func ServicePrincipal(cfg ServicePrincipalConfig) Provider {
	oauthCfg := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", cfg.TenantID),
		Scopes:       []string{"499b84ac-1321-427f-aa17-267ca6975798/.default"}, // Azure DevOps resource ID
	}

	return &servicePrincipalProvider{
		cfg: oauthCfg,
	}
}

type servicePrincipalProvider struct {
	cfg   *clientcredentials.Config
	mu    sync.Mutex
	token *oauth2.Token
}

func (p *servicePrincipalProvider) SetAuthHeaders(ctx context.Context, req *http.Request) error {
	token, err := p.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain access token: %w", err)
	}
	token.SetAuthHeader(req)
	return nil
}

func (p *servicePrincipalProvider) getToken(ctx context.Context) (*oauth2.Token, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Return cached token if still valid (with 60s buffer)
	if p.token != nil && p.token.Expiry.After(time.Now().Add(60*time.Second)) {
		return p.token, nil
	}

	token, err := p.cfg.Token(ctx)
	if err != nil {
		return nil, err
	}

	p.token = token
	return token, nil
}
