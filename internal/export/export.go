package export

import (
	"context"
	"fmt"

	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// Exporter renders a ParsedDocument to serialized file bytes.
type Exporter interface {
	Export(ctx context.Context, doc *parser.ParsedDocument) ([]byte, error)
	ContentType() string
	Extension() string
}

func validate(doc *parser.ParsedDocument) error {
	if doc == nil || len(doc.Columns) == 0 {
		return fmt.Errorf("%w: no columns", parser.ErrInvalidDocument)
	}
	ncol := len(doc.Columns)
	for i, r := range doc.Rows {
		if len(r) > ncol {
			return fmt.Errorf("%w: row %d has %d values for %d columns", parser.ErrInvalidDocument, i, len(r), ncol)
		}
	}
	return nil
}
