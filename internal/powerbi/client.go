package powerbi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://api.powerbi.com/v1.0/myorg"

// Client is the Power BI REST surface the publisher needs.
type Client interface {
	// GetDatasetByName returns nil, nil when no dataset matches.
	GetDatasetByName(ctx context.Context, workspaceID, name string) (*Dataset, error)
	CreateDataset(ctx context.Context, workspaceID string, def DatasetDef) (*Dataset, error)
	DeleteRows(ctx context.Context, workspaceID, datasetID, table string) error
	PostRows(ctx context.Context, workspaceID, datasetID, table string, rows []map[string]any) error
}

type httpClient struct {
	cfg    Config
	http   *http.Client
	tokens *tokenSource
}

func NewClient(cfg Config, hc *http.Client) Client {
	if hc == nil {
		hc = http.DefaultClient
	}
	return &httpClient{cfg: cfg, http: hc, tokens: &tokenSource{cfg: cfg, client: hc}}
}

func (c *httpClient) base() string {
	if c.cfg.BaseURL != "" {
		return strings.TrimRight(c.cfg.BaseURL, "/")
	}
	return defaultBaseURL
}

func scope(workspaceID string) string {
	if workspaceID == "" {
		return ""
	}
	return "/groups/" + workspaceID
}

func (c *httpClient) GetDatasetByName(ctx context.Context, workspaceID, name string) (*Dataset, error) {
	var out struct {
		Value []Dataset `json:"value"`
	}
	if err := c.do(ctx, http.MethodGet, scope(workspaceID)+"/datasets", nil, &out); err != nil {
		return nil, err
	}
	for _, d := range out.Value {
		if d.Name == name {
			return &d, nil
		}
	}
	return nil, nil
}

func (c *httpClient) CreateDataset(ctx context.Context, workspaceID string, def DatasetDef) (*Dataset, error) {
	var out Dataset
	if err := c.do(ctx, http.MethodPost, scope(workspaceID)+"/datasets", def, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) DeleteRows(ctx context.Context, workspaceID, datasetID, table string) error {
	path := fmt.Sprintf("%s/datasets/%s/tables/%s/rows", scope(workspaceID), datasetID, url.PathEscape(table))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *httpClient) PostRows(ctx context.Context, workspaceID, datasetID, table string, rows []map[string]any) error {
	path := fmt.Sprintf("%s/datasets/%s/tables/%s/rows", scope(workspaceID), datasetID, url.PathEscape(table))
	return c.do(ctx, http.MethodPost, path, map[string]any{"rows": rows}, nil)
}

func (c *httpClient) do(ctx context.Context, method, path string, body, out any) error {
	token, err := c.tokens.Token(ctx)
	if err != nil {
		return err
	}

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base()+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return statusError(resp)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func statusError(resp *http.Response) error {
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	var kind error
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		kind = ErrUnauthenticated
	case resp.StatusCode == http.StatusForbidden:
		kind = ErrPermissionDenied
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
		kind = ErrUnavailable
	default:
		kind = ErrAPI
	}
	detail := strings.TrimSpace(string(msg))
	if detail == "" {
		return fmt.Errorf("%w: %s", kind, resp.Status)
	}
	return fmt.Errorf("%w: %s: %s", kind, resp.Status, detail)
}
