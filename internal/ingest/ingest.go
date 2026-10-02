// Package ingest runs the pipeline: validate request -> detect -> adapt ->
// validate document. Format internals live in adapters; adding one is a map entry.
package ingest

import (
	"context"
	"fmt"

	"github.com/bushadigitallimited/bloom-parser/internal/document"
	"github.com/bushadigitallimited/bloom-parser/internal/format"
	"github.com/bushadigitallimited/bloom-parser/internal/format/image"
	"github.com/bushadigitallimited/bloom-parser/internal/format/pdf"
	"github.com/bushadigitallimited/bloom-parser/internal/format/text"
	"github.com/bushadigitallimited/bloom-parser/internal/format/xlsx"
	"github.com/bushadigitallimited/bloom-parser/internal/ocr"
)

// Limits bound resource use at the request boundary.
type Limits struct {
	MaxBytes int // reject inputs larger than this; 0 means unlimited
	MaxPages int // default per-document page cap; 0 means unlimited
}

// DefaultLimits are conservative defaults applied when none are configured.
var DefaultLimits = Limits{MaxBytes: 32 << 20, MaxPages: 200}

// Request is one ingestion request.
type Request struct {
	Name    string
	Content []byte
	Format  document.Format // FormatUnknown triggers detection
	Options format.Options
}

// Service routes requests to the registered adapter for the detected format.
type Service struct {
	adapters map[document.Format]format.Adapter
	limits   Limits
	engine   ocr.Engine
}

// New builds the pipeline; one image adapter serves every raster format.
func New(engine ocr.Engine, limits Limits) *Service {
	engine = ocr.OrDisabled(engine)
	img := image.New(engine)
	txt := text.New()
	adapters := map[document.Format]format.Adapter{
		document.FormatPNG:      img,
		document.FormatJPEG:     img,
		document.FormatWEBP:     img,
		document.FormatTIFF:     img,
		document.FormatGIF:      img,
		document.FormatBMP:      img,
		document.FormatPDF:      pdf.New(engine),
		document.FormatXLSX:     xlsx.New(),
		document.FormatCSV:      txt,
		document.FormatJSON:     txt,
		document.FormatMarkdown: txt,
	}
	return &Service{adapters: adapters, limits: limits, engine: engine}
}

// Engine returns the configured OCR engine.
func (s *Service) Engine() ocr.Engine { return s.engine }

// Ingest runs the pipeline. Per-page failures stay on the pages; a returned
// error means the request or whole input was unusable.
func (s *Service) Ingest(ctx context.Context, req Request) (*document.Document, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	f := req.Format
	if f == document.FormatUnknown {
		f = document.Detect(req.Content)
	}
	adapter, ok := s.adapters[f]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, f)
	}

	opts := req.Options
	if opts.MaxPages == 0 {
		opts.MaxPages = s.limits.MaxPages
	}

	pages, err := adapter.Adapt(ctx, format.Input{
		Name:    req.Name,
		Content: req.Content,
		Format:  f,
		Options: opts,
	})
	if err != nil {
		return nil, err
	}

	doc := &document.Document{Name: req.Name, Format: f, Pages: pages}
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}
