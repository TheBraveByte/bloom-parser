package ingest

import (
	"errors"
	"fmt"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// Sentinel errors; some alias parser's so the gRPC layer keeps one mapping.
var (
	ErrEmptyInput        = parser.ErrEmptyInput
	ErrUnsupportedFormat = parser.ErrUnsupportedFormat
	ErrTooLarge          = errors.New("document exceeds the maximum allowed size")
	ErrNoPages           = errors.New("document produced no pages")
)

func (s *Service) validateRequest(req Request) error {
	if len(req.Content) == 0 {
		return ErrEmptyInput
	}
	if s.limits.MaxBytes > 0 && len(req.Content) > s.limits.MaxBytes {
		return fmt.Errorf("%w: %d bytes exceeds limit of %d", ErrTooLarge, len(req.Content), s.limits.MaxBytes)
	}
	return nil
}

func validateDocument(doc *document.Document) error {
	if len(doc.Pages) == 0 {
		return ErrNoPages
	}
	for pi := range doc.Pages {
		for ti, t := range doc.Pages[pi].Tables {
			if err := validateTable(t); err != nil {
				return fmt.Errorf("page %d table %d: %w", doc.Pages[pi].Number, ti, err)
			}
		}
	}
	return nil
}

func validateTable(t *parser.ParsedDocument) error {
	if t == nil || len(t.Columns) == 0 {
		return fmt.Errorf("%w: table has no columns", parser.ErrInvalidDocument)
	}
	ncol := len(t.Columns)
	for ri, row := range t.Rows {
		if len(row) > ncol {
			return fmt.Errorf("%w: row %d has %d values for %d columns", parser.ErrInvalidDocument, ri, len(row), ncol)
		}
	}
	return nil
}
