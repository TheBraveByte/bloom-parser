package powerbi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bushadigitallimited/bloom-parser/internal/parser"
)

// fakePowerBI is an in-memory Power BI + Entra stub.
type fakePowerBI struct {
	t *testing.T
	*httptest.Server

	tokenCalls int
	requests   []string
	existing   []Dataset
	datasets   []DatasetDef
	pushedRows []map[string]any
}

func newFakePowerBI(t *testing.T) *fakePowerBI {
	f := &fakePowerBI{t: t}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	return f
}

func (f *fakePowerBI) serve(w http.ResponseWriter, r *http.Request) {
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)

	if strings.HasSuffix(r.URL.Path, "/oauth2/v2.0/token") {
		f.tokenCalls++
		if err := r.ParseForm(); err != nil {
			f.t.Errorf("token form: %v", err)
		}
		if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_secret") != "s3cret" {
			http.Error(w, "bad creds", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": "tok-1", "expires_in": 3600})
		return
	}

	if r.Header.Get("Authorization") != "Bearer tok-1" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/groups/ws-1/datasets"):
		json.NewEncoder(w).Encode(map[string]any{"value": f.existing})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/groups/ws-1/datasets"):
		var def DatasetDef
		json.NewDecoder(r.Body).Decode(&def)
		f.datasets = append(f.datasets, def)
		json.NewEncoder(w).Encode(Dataset{ID: "ds-1", Name: def.Name})
	case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/rows"):
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/rows"):
		var body struct {
			Rows []map[string]any `json:"rows"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		f.pushedRows = append(f.pushedRows, body.Rows...)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "no route: "+r.URL.Path, http.StatusNotFound)
	}
}

func (f *fakePowerBI) config() Config {
	return Config{
		TenantID:     "t-1",
		ClientID:     "c-1",
		ClientSecret: "s3cret",
		WorkspaceID:  "ws-1",
		BaseURL:      f.URL + "/v1.0/myorg",
		AuthURL:      f.URL,
	}
}

func TestPublishEndToEnd(t *testing.T) {
	fake := newFakePowerBI(t)
	defer fake.Close()

	doc := &parser.ParsedDocument{
		Name: "sales",
		Columns: []parser.Column{
			{Name: "name", Type: parser.TypeString},
			{Name: "amount", Type: parser.TypeFloat},
			{Name: "count", Type: parser.TypeInt},
			{Name: "when", Type: parser.TypeTime},
			{Name: "ok", Type: parser.TypeBool},
		},
		Rows: [][]parser.Value{
			{parser.Str("a"), parser.Float(1.5), parser.Int(2), parser.Time(tm), parser.Bool(true)},
			{parser.Str("b"), parser.Value{Kind: parser.KindNull}, parser.Int(3), parser.Value{Kind: parser.KindNull}, parser.Bool(false)},
		},
	}

	pub := NewPublisher(NewClient(fake.config(), nil), "ws-1")
	res, err := pub.Publish(context.Background(), PublishRequest{
		Doc:         doc,
		DatasetName: "sales",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.DatasetID != "ds-1" || res.TableName != "Table1" || res.RowsPushed != 2 {
		t.Errorf("result = %+v", res)
	}
	if fake.tokenCalls != 1 {
		t.Errorf("token fetched %d times, want 1 (cached)", fake.tokenCalls)
	}
	if len(fake.datasets) != 1 {
		t.Fatalf("datasets created = %d", len(fake.datasets))
	}
	def := fake.datasets[0]
	if def.DefaultMode != "Push" || len(def.Tables) != 1 || def.Tables[0].Name != "Table1" {
		t.Fatalf("dataset def = %+v", def)
	}
	wantTypes := []string{"String", "Double", "Int64", "DateTime", "Boolean"}
	for i, w := range wantTypes {
		if def.Tables[0].Columns[i].DataType != w {
			t.Errorf("col %d type = %q, want %q", i, def.Tables[0].Columns[i].DataType, w)
		}
	}
	if len(fake.pushedRows) != 2 {
		t.Fatalf("pushed rows = %d", len(fake.pushedRows))
	}
	if fake.pushedRows[0]["name"] != "a" || fake.pushedRows[0]["ok"] != true {
		t.Errorf("row0 = %v", fake.pushedRows[0])
	}
	if fake.pushedRows[1]["amount"] != nil {
		t.Errorf("null cell pushed as %v", fake.pushedRows[1]["amount"])
	}
}

func TestPublishSkipsCreateWhenDatasetExists(t *testing.T) {
	fake := newFakePowerBI(t)
	defer fake.Close()

	doc := &parser.ParsedDocument{
		Columns: []parser.Column{{Name: "a", Type: parser.TypeString}},
		Rows:    [][]parser.Value{{parser.Str("x")}},
	}
	fake.existing = []Dataset{{ID: "ds-existing", Name: "sales"}}

	pub := NewPublisher(NewClient(fake.config(), nil), "ws-1")
	res, err := pub.Publish(context.Background(), PublishRequest{Doc: doc, DatasetName: "sales"})
	if err != nil {
		t.Fatal(err)
	}
	if res.DatasetID != "ds-existing" {
		t.Errorf("dataset = %q", res.DatasetID)
	}
	if len(fake.datasets) != 0 {
		t.Error("dataset should not be recreated")
	}
}

func TestClientErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{401, ErrUnauthenticated},
		{403, ErrPermissionDenied},
		{429, ErrUnavailable},
		{500, ErrUnavailable},
		{400, ErrAPI},
	}
	for _, c := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "oauth2") {
				json.NewEncoder(w).Encode(map[string]any{"access_token": "t", "expires_in": 3600})
				return
			}
			http.Error(w, "boom", c.status)
		}))
		cfg := Config{TenantID: "t", ClientID: "c", ClientSecret: "s", BaseURL: srv.URL, AuthURL: srv.URL}
		_, err := NewClient(cfg, nil).GetDatasetByName(context.Background(), "", "x")
		if !errors.Is(err, c.want) {
			t.Errorf("status %d -> %v, want %v", c.status, err, c.want)
		}
		srv.Close()
	}
}

func TestBadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
	}))
	defer srv.Close()
	cfg := Config{TenantID: "t", ClientID: "bad", ClientSecret: "bad", BaseURL: srv.URL, AuthURL: srv.URL}
	_, err := NewClient(cfg, nil).GetDatasetByName(context.Background(), "", "x")
	if !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("err = %v", err)
	}
}

var tm = mustTime("2024-03-01T10:30:00Z")

func mustTime(s string) (t time.Time) { t, _ = time.Parse(time.RFC3339, s); return }
