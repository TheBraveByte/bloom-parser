package pdf

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
	"github.com/TheBraveByte/bloom-parser/internal/testfixtures"
)

func adapt(t *testing.T, a *Adapter, content []byte, opts format.Options) []document.Page {
	t.Helper()
	pages, err := a.Adapt(context.Background(), format.Input{Content: content, Format: document.FormatPDF, Options: opts})
	if err != nil {
		t.Fatalf("adapt: %v", err)
	}
	return pages
}

func TestTextPDF(t *testing.T) {
	a := New(ocr.Disabled())
	pages := adapt(t, a, testfixtures.PDFText("Hello Bloom 123"), format.Options{})
	if len(pages) != 1 {
		t.Fatalf("pages = %d", len(pages))
	}
	p := pages[0]
	if p.Kind != document.PageText || p.Err != nil {
		t.Fatalf("page = %+v", p)
	}
	if !strings.Contains(p.Text, "Hello Bloom 123") {
		t.Errorf("text = %q", p.Text)
	}
}

func TestMultiPagePDF(t *testing.T) {
	a := New(ocr.Disabled())
	pages := adapt(t, a, testfixtures.PDFText("Page One", "Page Two", "Page Three"), format.Options{})
	if len(pages) != 3 {
		t.Fatalf("pages = %d, want 3", len(pages))
	}
	for i, want := range []string{"Page One", "Page Two", "Page Three"} {
		if pages[i].Number != i+1 || !strings.Contains(pages[i].Text, want) {
			t.Errorf("page %d number=%d text=%q", i, pages[i].Number, pages[i].Text)
		}
	}
}

func TestMaxPages(t *testing.T) {
	a := New(ocr.Disabled())
	pages := adapt(t, a, testfixtures.PDFText("a", "b", "c", "d"), format.Options{MaxPages: 2})
	if len(pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(pages))
	}
}

func TestScannedPDFNoOCR(t *testing.T) {
	a := New(ocr.Disabled())
	pages := adapt(t, a, testfixtures.ScannedPDF(), format.Options{})
	if len(pages) != 1 {
		t.Fatalf("pages = %d", len(pages))
	}
	p := pages[0]
	if p.Kind != document.PageImage {
		t.Errorf("kind = %v, want image", p.Kind)
	}
	if p.Err == nil || p.Err.Code != document.ErrOCRUnavailable {
		t.Errorf("err = %v, want OCR_UNAVAILABLE", p.Err)
	}
}

func TestScannedPDFWithEngineNeedsRasterizer(t *testing.T) {
	a := New(availableEngine{})
	pages := adapt(t, a, testfixtures.ScannedPDF(), format.Options{OCR: true})
	if pages[0].Err == nil || pages[0].Err.Code != document.ErrPageUnreadable {
		t.Errorf("err = %v, want PAGE_UNREADABLE", pages[0].Err)
	}
}

func TestMalformedPDF(t *testing.T) {
	a := New(ocr.Disabled())
	_, err := a.Adapt(context.Background(), format.Input{Content: testfixtures.MalformedPDF(), Format: document.FormatPDF})
	if !errors.Is(err, parser.ErrNotParseable) {
		t.Errorf("err = %v, want ErrNotParseable", err)
	}
}

type availableEngine struct{ ocr.Engine }

func (availableEngine) Available() bool { return true }
