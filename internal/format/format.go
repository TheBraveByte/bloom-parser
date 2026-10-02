// Package format is the contract between the pipeline and per-format adapters.
package format

import (
	"context"

	"github.com/bushadigitallimited/bloom-parser/internal/document"
)

// Options controls optional extraction behaviour shared by adapters.
type Options struct {
	OCR          bool   // attempt OCR on image pages
	OCRLanguages string // engine language hint, e.g. "eng"
	MaxPages     int    // cap on pages processed; 0 means all
}

// Input is a resolved, validated document ready for adaptation.
type Input struct {
	Name    string
	Content []byte
	Format  document.Format
	Options Options
}

// Adapter converts one input into pages. A returned error means the whole input
// failed; a single bad page is reported via Page.Err. Returns ≥1 page on nil err.
type Adapter interface {
	Adapt(ctx context.Context, in Input) ([]document.Page, error)
}

// Source returns name when non-empty, else fallback.
func Source(name, fallback string) string {
	if name == "" {
		return fallback
	}
	return name
}
