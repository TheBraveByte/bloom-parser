package xlsx

import (
	"context"
	"errors"
	"testing"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
	"github.com/TheBraveByte/bloom-parser/internal/testfixtures"
)

func adapt(t *testing.T, content []byte, opts format.Options) []document.Page {
	t.Helper()
	pages, err := New().Adapt(context.Background(), format.Input{Content: content, Format: document.FormatXLSX, Options: opts})
	if err != nil {
		t.Fatalf("adapt: %v", err)
	}
	return pages
}

func TestSingleSheet(t *testing.T) {
	wb := testfixtures.XLSX([]string{"Sales"}, map[string][][]any{
		"Sales": {
			{"name", "amount", "active"},
			{"Alice", 1200.5, true},
			{"Bob", 99, false},
		},
	})
	pages := adapt(t, wb, format.Options{})
	if len(pages) != 1 {
		t.Fatalf("pages = %d", len(pages))
	}
	p := pages[0]
	if p.Kind != document.PageSheet || p.Source != "Sales" {
		t.Errorf("page = %+v", p)
	}
	if len(p.Tables) != 1 {
		t.Fatalf("tables = %d", len(p.Tables))
	}
	tbl := p.Tables[0]
	if len(tbl.Columns) != 3 || tbl.Columns[1].Type != parser.TypeFloat {
		t.Errorf("columns = %+v", tbl.Columns)
	}
	if tbl.Columns[2].Type != parser.TypeBool {
		t.Errorf("active type = %v", tbl.Columns[2].Type)
	}
	if len(tbl.Rows) != 2 || tbl.Rows[0][0].Str != "Alice" {
		t.Errorf("rows = %+v", tbl.Rows)
	}
}

func TestMultipleSheets(t *testing.T) {
	wb := testfixtures.XLSX([]string{"One", "Two", "Three"}, map[string][][]any{
		"One":   {{"a"}, {1}},
		"Two":   {{"b"}, {2}},
		"Three": {{"c"}, {3}},
	})
	pages := adapt(t, wb, format.Options{})
	if len(pages) != 3 {
		t.Fatalf("pages = %d, want 3", len(pages))
	}
	for i, want := range []string{"One", "Two", "Three"} {
		if pages[i].Source != want || pages[i].Number != i+1 {
			t.Errorf("page %d = %q #%d", i, pages[i].Source, pages[i].Number)
		}
	}
}

func TestMaxPagesCapsSheets(t *testing.T) {
	wb := testfixtures.XLSX([]string{"One", "Two", "Three"}, map[string][][]any{
		"One": {{"a"}, {1}}, "Two": {{"b"}, {2}}, "Three": {{"c"}, {3}},
	})
	pages := adapt(t, wb, format.Options{MaxPages: 2})
	if len(pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(pages))
	}
}

func TestEmptySheet(t *testing.T) {
	wb := testfixtures.XLSX([]string{"Empty"}, map[string][][]any{"Empty": {}})
	pages := adapt(t, wb, format.Options{})
	if len(pages) != 1 || len(pages[0].Tables) != 0 {
		t.Fatalf("expected one empty page with no tables, got %+v", pages)
	}
	if len(pages[0].Warnings) == 0 || pages[0].Warnings[0].Code != "EMPTY_SHEET" {
		t.Errorf("warnings = %+v", pages[0].Warnings)
	}
}

func TestMalformedWorkbook(t *testing.T) {
	_, err := New().Adapt(context.Background(), format.Input{Content: []byte("PK\x03\x04garbage"), Format: document.FormatXLSX})
	if !errors.Is(err, parser.ErrNotParseable) {
		t.Errorf("err = %v, want ErrNotParseable", err)
	}
}

func TestRaggedAndTypedCells(t *testing.T) {
	wb := testfixtures.XLSX([]string{"S"}, map[string][][]any{
		"S": {
			{"id", "when"},
			{1, "2024-03-01"},
			{2},
		},
	})
	pages := adapt(t, wb, format.Options{})
	tbl := pages[0].Tables[0]
	if tbl.Columns[0].Type != parser.TypeInt || tbl.Columns[1].Type != parser.TypeTime {
		t.Errorf("types = %+v", tbl.Columns)
	}
	if tbl.Rows[1][1].Kind != parser.KindNull {
		t.Errorf("short-row cell should be null, got %v", tbl.Rows[1][1])
	}
}
