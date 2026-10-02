package export

import (
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/bushadigitallimited/bloom-parser/internal/parser"
)

func testDoc() *parser.ParsedDocument {
	return &parser.ParsedDocument{
		Name: "report",
		Columns: []parser.Column{
			{Name: "name", Type: parser.TypeString},
			{Name: "amount", Type: parser.TypeFloat},
			{Name: "count", Type: parser.TypeInt},
			{Name: "when", Type: parser.TypeTime},
			{Name: "ok", Type: parser.TypeBool},
		},
		Rows: [][]parser.Value{
			{parser.Str(`a,"b"`), parser.Float(1200.5), parser.Int(3), parser.Time(time.Date(2024, 3, 1, 10, 30, 0, 0, time.UTC)), parser.Bool(true)},
			{parser.Str("line1\nline2"), parser.Value{Kind: parser.KindNull}, parser.Int(-7), parser.Time(time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC)), parser.Bool(false)},
		},
	}
}

func TestCSVExport(t *testing.T) {
	out, err := CSV{}.Export(context.Background(), testDoc())
	if err != nil {
		t.Fatal(err)
	}
	recs, err := csv.NewReader(strings.NewReader(string(out))).ReadAll()
	if err != nil {
		t.Fatalf("re-read csv: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("records = %d", len(recs))
	}
	if recs[0][0] != "name" || recs[0][4] != "ok" {
		t.Errorf("header = %v", recs[0])
	}
	if recs[1][0] != `a,"b"` {
		t.Errorf("comma+quote cell = %q", recs[1][0])
	}
	if recs[2][0] != "line1\nline2" {
		t.Errorf("newline cell = %q", recs[2][0])
	}
	if recs[2][1] != "" {
		t.Errorf("null cell = %q", recs[2][1])
	}
	if recs[1][3] != "2024-03-01T10:30:00Z" {
		t.Errorf("datetime cell = %q", recs[1][3])
	}
}

func TestCSVRejectsBadDoc(t *testing.T) {
	if _, err := (CSV{}).Export(context.Background(), &parser.ParsedDocument{}); err == nil {
		t.Error("empty doc should fail")
	}
	bad := testDoc()
	bad.Rows[0] = append(bad.Rows[0], parser.Str("extra"))
	if _, err := (CSV{}).Export(context.Background(), bad); err == nil {
		t.Error("overlong row should fail")
	}
}

func TestXLSXExport(t *testing.T) {
	out, err := XLSX{}.Export(context.Background(), testDoc())
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet != "report" {
		t.Errorf("sheet = %q", sheet)
	}
	get := func(cell string) string {
		v, _ := f.GetCellValue(sheet, cell)
		return v
	}
	if get("A1") != "name" || get("E1") != "ok" {
		t.Errorf("headers: %q %q", get("A1"), get("E1"))
	}
	if get("A2") != `a,"b"` {
		t.Errorf("A2 = %q", get("A2"))
	}
	if get("B2") != "1200.5" {
		t.Errorf("B2 = %q", get("B2"))
	}
	if get("C3") != "-7" {
		t.Errorf("C3 = %q", get("C3"))
	}
	if get("E2") != "TRUE" {
		t.Errorf("E2 = %q", get("E2"))
	}
	// Datetime cell stored as a serial number with date formatting.
	if get("D2") == "" || get("D2") == "2024-03-01T10:30:00Z" {
		t.Errorf("D2 = %q (expected excel serial)", get("D2"))
	}
	if get("B3") != "" {
		t.Errorf("null cell B3 = %q", get("B3"))
	}
}

func TestSheetNameSanitization(t *testing.T) {
	doc := testDoc()
	doc.Name = "bad/name:with[chars]"
	out, err := XLSX{}.Export(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := excelize.OpenReader(strings.NewReader(string(out)))
	defer f.Close()
	if n := f.GetSheetName(0); strings.ContainsAny(n, `[]:*?/\`) {
		t.Errorf("sheet name %q contains invalid chars", n)
	}
	doc.Name = ""
	if _, err := (XLSX{}).Export(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
}
