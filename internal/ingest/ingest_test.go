package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
	"github.com/TheBraveByte/bloom-parser/internal/testfixtures"
)

func newSvc() *Service { return New(ocr.Disabled(), DefaultLimits) }

func ingest(t *testing.T, req Request) *document.Document {
	t.Helper()
	doc, err := newSvc().Ingest(context.Background(), req)
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	return doc
}

// Exercises the full pipeline for each format via auto-detection:
// detect -> adapter -> extraction -> normalization -> validation -> Document.
func TestPipelineAutoDetect(t *testing.T) {
	cases := []struct {
		name    string
		content []byte
		format  document.Format
	}{
		{"png", testfixtures.PNG(30, 30), document.FormatPNG},
		{"jpeg", testfixtures.JPEG(30, 30), document.FormatJPEG},
		{"tiff", testfixtures.TIFF(30, 30), document.FormatTIFF},
		{"pdf", testfixtures.PDFText("hello"), document.FormatPDF},
		{"xlsx", testfixtures.XLSX([]string{"S"}, map[string][][]any{"S": {{"a", "b"}, {1, 2}}}), document.FormatXLSX},
		{"csv", []byte("a,b\n1,2"), document.FormatCSV},
		{"json", []byte(`[{"a":1}]`), document.FormatJSON},
		{"markdown", []byte("a | b\n--|--\n1 | 2"), document.FormatMarkdown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := ingest(t, Request{Name: c.name, Content: c.content})
			if doc.Format != c.format {
				t.Errorf("detected %v, want %v", doc.Format, c.format)
			}
			if len(doc.Pages) == 0 {
				t.Error("no pages produced")
			}
		})
	}
}

func TestPipelineTabularProducesTable(t *testing.T) {
	doc := ingest(t, Request{Content: []byte("name,amount\nAlice,10.5\nBob,20")})
	tbl := doc.FirstTable()
	if tbl == nil || len(tbl.Columns) != 2 {
		t.Fatalf("table = %+v", tbl)
	}
	if tbl.Columns[1].Type != parser.TypeFloat {
		t.Errorf("amount type = %v", tbl.Columns[1].Type)
	}
}

func TestPipelineXLSXMultiSheet(t *testing.T) {
	wb := testfixtures.XLSX([]string{"A", "B"}, map[string][][]any{
		"A": {{"x"}, {1}}, "B": {{"y"}, {2}},
	})
	doc := ingest(t, Request{Content: wb})
	if len(doc.Pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(doc.Pages))
	}
	if doc.TableCount() != 2 {
		t.Errorf("tables = %d, want 2", doc.TableCount())
	}
}

func TestPipelinePDFPageFailureIsolated(t *testing.T) {
	doc := ingest(t, Request{Content: testfixtures.ScannedPDF()})
	if len(doc.Pages) != 1 {
		t.Fatalf("pages = %d", len(doc.Pages))
	}
	if doc.Pages[0].Err == nil || doc.Pages[0].Err.Code != document.ErrOCRUnavailable {
		t.Errorf("page error = %v", doc.Pages[0].Err)
	}
}

func TestValidateEmptyInput(t *testing.T) {
	_, err := newSvc().Ingest(context.Background(), Request{Content: nil})
	if !errors.Is(err, ErrEmptyInput) {
		t.Errorf("err = %v, want ErrEmptyInput", err)
	}
}

func TestValidateTooLarge(t *testing.T) {
	svc := New(ocr.Disabled(), Limits{MaxBytes: 10})
	_, err := svc.Ingest(context.Background(), Request{Content: []byte("more than ten bytes")})
	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("err = %v, want ErrTooLarge", err)
	}
}

func TestExplicitFormatOverridesDetection(t *testing.T) {
	_, err := newSvc().Ingest(context.Background(), Request{Content: []byte("a,b\n1,2"), Format: document.FormatJSON})
	if !errors.Is(err, parser.ErrNotParseable) {
		t.Errorf("err = %v, want ErrNotParseable", err)
	}
}

func TestMalformedContentPerFormat(t *testing.T) {
	cases := []struct {
		name    string
		content []byte
	}{
		{"bad-pdf", testfixtures.MalformedPDF()},
		{"bad-xlsx", []byte("PK\x03\x04garbage")},
		{"bad-image", []byte("\x89PNG\r\n\x1a\nnope")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := newSvc().Ingest(context.Background(), Request{Content: c.content})
			if !errors.Is(err, parser.ErrNotParseable) {
				t.Errorf("err = %v, want ErrNotParseable", err)
			}
		})
	}
}

func TestEngineExposed(t *testing.T) {
	if name := newSvc().Engine().Name(); !strings.Contains(name, "disabled") {
		t.Errorf("engine = %q", name)
	}
}
