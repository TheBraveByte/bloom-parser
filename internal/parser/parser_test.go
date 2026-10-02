package parser

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mustParse(t *testing.T, file string, f Format) *ParsedDocument {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../../tests/fixtures", file))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := NewService().Parse(context.Background(), Input{Name: file, Content: b}, f)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	return doc
}

func warningCodes(doc *ParsedDocument) map[string]int {
	m := map[string]int{}
	for _, w := range doc.Warnings {
		m[w.Code]++
	}
	return m
}

func TestParseMessyCSV(t *testing.T) {
	doc := mustParse(t, "messy.csv", FormatCSV)

	wantCols := []Column{
		{"name", TypeString}, {"name_2", TypeString}, {"amount", TypeFloat},
		{"transaction_date", TypeTime}, {"active", TypeBool}, {"notes", TypeString},
	}
	if len(doc.Columns) != len(wantCols) {
		t.Fatalf("columns = %v", doc.Columns)
	}
	for i, c := range wantCols {
		if doc.Columns[i] != c {
			t.Errorf("column %d = %+v, want %+v", i, doc.Columns[i], c)
		}
	}

	codes := warningCodes(doc)
	for _, code := range []string{"DUPLICATE_COLUMN", "ROW_TOO_SHORT", "TYPE_MISMATCH"} {
		if codes[code] == 0 {
			t.Errorf("expected warning %s, got %v", code, codes)
		}
	}

	if got := doc.Rows[0][0].Str; got != "Smith, John" {
		t.Errorf("row0 name = %q", got)
	}
	if doc.Rows[0][2].Float != 1200.50 {
		t.Errorf("row0 amount = %v", doc.Rows[0][2].Float)
	}
	if doc.Rows[1][2].Float != 2300.75 {
		t.Errorf("row1 amount = %v", doc.Rows[1][2].Float)
	}
	if doc.Rows[4][2].Kind != KindNull {
		t.Errorf("row4 amount kind = %v, want null", doc.Rows[4][2].Kind)
	}
	if doc.Rows[1][4].Bool != false || doc.Rows[5][4].Bool != true {
		t.Error("yes/no not converted to bool")
	}
	if doc.Rows[5][3].Kind != KindTime {
		t.Error("RFC3339 cell not parsed as time")
	}
	if doc.Rows[0][5].Str != "first\nline" {
		t.Errorf("row0 notes = %q", doc.Rows[0][5].Str)
	}
}

func TestParseMarkdown(t *testing.T) {
	doc := mustParse(t, "table.md", FormatAuto)
	if doc.Format != FormatMarkdown {
		t.Fatalf("format = %v", doc.Format)
	}
	if len(doc.Columns) != 4 || doc.Columns[0].Name != "customer_name" {
		t.Fatalf("columns = %v", doc.Columns)
	}
	if len(doc.Rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(doc.Rows))
	}
	if doc.Columns[1].Type != TypeFloat {
		t.Errorf("amount type = %v", doc.Columns[1].Type)
	}
	if doc.Columns[3].Type != TypeBool {
		t.Errorf("active type = %v", doc.Columns[3].Type)
	}
	if doc.Rows[2][0].Kind != KindNull || doc.Rows[2][3].Kind == KindNull {
		t.Error("empty cells not nulled correctly")
	}
}

func TestParseJSON(t *testing.T) {
	doc := mustParse(t, "records.json", FormatJSON)

	want := []string{"customer_name", "amount", "transaction_date", "active", "extra"}
	if len(doc.Columns) != len(want) {
		t.Fatalf("columns = %v", doc.Columns)
	}
	for i, n := range want {
		if doc.Columns[i].Name != n {
			t.Errorf("col %d = %q, want %q", i, doc.Columns[i].Name, n)
		}
	}
	if doc.Rows[1][2].Kind != KindNull {
		t.Error("missing key not null")
	}
	if doc.Rows[1][4].Str != `{"nested":1}` {
		t.Errorf("nested cell = %q", doc.Rows[1][4].Str)
	}
	if doc.Rows[3][1].Kind != KindNull {
		t.Error("bad amount not null")
	}
	if warningCodes(doc)["TYPE_MISMATCH"] == 0 {
		t.Error("expected TYPE_MISMATCH warnings")
	}
}

func TestDetect(t *testing.T) {
	cases := []struct {
		in   string
		want Format
	}{
		{`[{"a":1}]`, FormatJSON},
		{"a | b\n--|--\n1 | 2", FormatMarkdown},
		{"a,b,c\n1,2,3", FormatCSV},
		{"not,a table\nbut csv", FormatCSV},
	}
	for _, c := range cases {
		if got := Detect([]byte(c.in)); got != c.want {
			t.Errorf("Detect(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	if _, err := svc.Parse(ctx, Input{Content: []byte("  ")}, FormatAuto); !errors.Is(err, ErrEmptyInput) {
		t.Errorf("empty: %v", err)
	}
	if _, err := svc.Parse(ctx, Input{Content: []byte("x")}, Format(99)); !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("bad format: %v", err)
	}
	if _, err := svc.Parse(ctx, Input{Content: []byte("no table here")}, FormatMarkdown); !errors.Is(err, ErrNotParseable) {
		t.Errorf("markdown: %v", err)
	}
	if _, err := svc.Parse(ctx, Input{Content: []byte("{not json")}, FormatJSON); !errors.Is(err, ErrNotParseable) {
		t.Errorf("json: %v", err)
	}
	if _, err := svc.Parse(ctx, Input{Content: []byte(`[{"a":1},2]`)}, FormatJSON); !errors.Is(err, ErrNotParseable) {
		t.Errorf("json non-object: %v", err)
	}
	if _, err := svc.Parse(ctx, Input{Content: []byte("[]")}, FormatJSON); !errors.Is(err, ErrNotParseable) {
		t.Errorf("json empty: %v", err)
	}
	if _, err := svc.Parse(ctx, Input{Content: []byte("")}, FormatCSV); !errors.Is(err, ErrEmptyInput) {
		t.Errorf("csv empty: %v", err)
	}
}

func TestDatetimeFormats(t *testing.T) {
	in := `d
2024-03-01
03/15/2024
2024-03-01T10:30:00Z
2006/01/02`
	doc, err := NewService().Parse(context.Background(), Input{Content: []byte(in)}, FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Columns[0].Type != TypeTime {
		t.Fatalf("type = %v", doc.Columns[0].Type)
	}
	for i, r := range doc.Rows {
		if r[0].Kind != KindTime {
			t.Errorf("row %d kind = %v", i, r[0].Kind)
		}
	}
	if !doc.Rows[0][0].Time.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("date parse wrong")
	}
}

func TestLargeRowCount(t *testing.T) {
	var sb []byte
	sb = append(sb, "id,val\n"...)
	for i := 0; i < 20000; i++ {
		sb = append(sb, []byte("1,x\n")...)
	}
	doc, err := NewService().Parse(context.Background(), Input{Content: sb}, FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Rows) != 20000 {
		t.Fatalf("rows = %d", len(doc.Rows))
	}
	if doc.Columns[0].Type != TypeInt {
		t.Errorf("id type = %v", doc.Columns[0].Type)
	}
}
