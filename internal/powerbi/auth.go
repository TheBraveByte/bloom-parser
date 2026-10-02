package powerbi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultAuthURL = "https://login.microsoftonline.com"

const powerBIScope = "https://analysis.windows.net/powerbi/api/.default"

type tokenSource struct {
	cfg    Config
	client *http.Client

	mu    sync.Mutex
	token string
	exp   time.Time
}

func (t *tokenSource) tokenURL() string {
	base := strings.TrimRight(t.cfg.AuthURL, "/")
	if base == "" {
		base = defaultAuthURL
	}
	return base + "/" + t.cfg.TenantID + "/oauth2/v2.0/token"
}

func (t *tokenSource) Token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.token != "" && time.Now().Before(t.exp) {
		return t.token, nil
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {t.cfg.ClientID},
		"client_secret": {t.cfg.ClientSecret},
		"scope":         {powerBIScope},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.tokenURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: token endpoint returned %s", ErrUnauthenticated, resp.Status)
	}
	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.AccessToken == "" {
		return "", fmt.Errorf("%w: malformed token response", ErrUnauthenticated)
	}
	t.token = body.AccessToken
	t.exp = time.Now().Add(time.Duration(body.ExpiresIn)*time.Second - time.Minute)
	return t.token, nil
}
