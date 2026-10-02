package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// CSV renders a ParsedDocument as UTF-8 CSV. encoding/csv handles quoting of
// commas, quotes and newlines; values use their canonical String form.
type CSV struct{}

func (CSV) ContentType() string { return "text/csv" }
func (CSV) Extension() string   { return ".csv" }

func (CSV) Export(_ context.Context, doc *parser.ParsedDocument) ([]byte, error) {
	if err := validate(doc); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := make([]string, len(doc.Columns))
	for i, c := range doc.Columns {
		header[i] = c.Name
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	rec := make([]string, len(doc.Columns))
	for _, row := range doc.Rows {
		for i := range rec {
			if i < len(row) {
				rec[i] = row[i].String()
			} else {
				rec[i] = ""
			}
		}
		if err := w.Write(rec); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}
	return buf.Bytes(), nil
}
