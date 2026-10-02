// Package text adapts delimited/markdown/JSON into one text page with one typed
// table, delegating to parser.Service.
package text

import (
	"context"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// Adapter wraps the tabular parser service.
type Adapter struct {
	svc *parser.Service
}

func New() *Adapter { return &Adapter{svc: parser.NewService()} }

func (a *Adapter) Adapt(ctx context.Context, in format.Input) ([]document.Page, error) {
	doc, err := a.svc.Parse(ctx, parser.Input{Name: in.Name, Content: in.Content}, parserFormat(in.Format))
	if err != nil {
		return nil, err
	}
	page := document.Page{
		Number:     1,
		Kind:       document.PageText,
		Source:     format.Source(in.Name, "content"),
		Tables:     []*parser.ParsedDocument{doc},
		Confidence: document.ConfidenceUnknown,
	}
	return []document.Page{page}, nil
}

func parserFormat(f document.Format) parser.Format {
	switch f {
	case document.FormatMarkdown:
		return parser.FormatMarkdown
	case document.FormatJSON:
		return parser.FormatJSON
	case document.FormatCSV:
		return parser.FormatCSV
	default:
		return parser.FormatAuto
	}
}
