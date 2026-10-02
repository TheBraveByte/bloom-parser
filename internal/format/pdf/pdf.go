// Package pdf adapts a PDF into one page per PDF page: native text where
// present, otherwise a scanned page routed to OCR.
package pdf

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	lpdf "github.com/ledongthuc/pdf"

	"github.com/bushadigitallimited/bloom-parser/internal/document"
	"github.com/bushadigitallimited/bloom-parser/internal/format"
	"github.com/bushadigitallimited/bloom-parser/internal/ocr"
	"github.com/bushadigitallimited/bloom-parser/internal/parser"
)

// Adapter extracts native text and detects scanned pages.
type Adapter struct {
	engine ocr.Engine
}

func New(engine ocr.Engine) *Adapter {
	return &Adapter{engine: ocr.OrDisabled(engine)}
}

func (a *Adapter) Adapt(_ context.Context, in format.Input) ([]document.Page, error) {
	r, err := lpdf.NewReader(bytes.NewReader(in.Content), int64(len(in.Content)))
	if err != nil {
		return nil, fmt.Errorf("%w: open pdf: %v", parser.ErrNotParseable, err)
	}
	total := r.NumPage()
	if total <= 0 {
		return nil, fmt.Errorf("%w: pdf has no pages", parser.ErrNotParseable)
	}
	limit := total
	if in.Options.MaxPages > 0 && in.Options.MaxPages < limit {
		limit = in.Options.MaxPages
	}

	pages := make([]document.Page, 0, limit)
	for n := 1; n <= limit; n++ {
		pages = append(pages, a.extractPage(r, n))
	}
	return pages, nil
}

func (a *Adapter) extractPage(r *lpdf.Reader, n int) document.Page {
	text, err := plainText(r, n)
	if err != nil {
		return document.FailedPage(n, pageLabel(n), document.ErrPageUnreadable, err.Error())
	}

	if strings.TrimSpace(text) != "" {
		return document.Page{
			Number:     n,
			Kind:       document.PageText,
			Source:     pageLabel(n),
			Text:       text,
			Confidence: document.ConfidenceUnknown,
		}
	}

	// No text layer: a scanned page needing rasterized pixels for OCR.
	page := document.Page{Number: n, Kind: document.PageImage, Source: pageLabel(n), Confidence: document.ConfidenceUnknown}
	if !a.engine.Available() {
		page.Err = &document.PageError{Code: document.ErrOCRUnavailable, Message: "scanned page has no text layer and no OCR engine is configured"}
		return page
	}
	page.Err = &document.PageError{Code: document.ErrPageUnreadable, Message: "scanned page requires a PDF rasterizer to produce an image for OCR"}
	return page
}

// plainText extracts one page's text, recovering from reader panics.
func plainText(r *lpdf.Reader, n int) (text string, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			text, err = "", fmt.Errorf("page %d: %v", n, rec)
		}
	}()
	p := r.Page(n)
	if p.V.IsNull() {
		return "", fmt.Errorf("page %d is missing", n)
	}
	return p.GetPlainText(nil)
}

func pageLabel(n int) string { return fmt.Sprintf("page-%d", n) }
