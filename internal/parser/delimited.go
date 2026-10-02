package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
)

// csvParser handles delimited text (comma, semicolon, tab). The first row is
// treated as the header; the dialect is auto-detected.
type csvParser struct{}

func (csvParser) Format() Format { return FormatCSV }

func (csvParser) Parse(_ context.Context, in Input) (*ParsedDocument, error) {
	delim := detectDelimiter(in.Content)
	r := csv.NewReader(bytes.NewReader(in.Content))
	r.Comma = delim
	r.FieldsPerRecord = -1 // allow ragged rows; normalize aligns them
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	var raw [][]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w as csv: %v", ErrNotParseable, err)
		}
		raw = append(raw, rec)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: no rows", ErrNotParseable)
	}
	return normalize(rawTable{
		headers: raw[0],
		rows:    raw[1:],
		attrs:   map[string]string{"delimiter": string(delim)},
	})
}

// detectDelimiter picks the candidate that yields the most consistent
// multi-column split over the first lines.
func detectDelimiter(content []byte) rune {
	candidates := []rune{',', ';', '\t', '|'}
	sample := content
	if i := bytes.Index(content, []byte("\n\n\n")); i > 0 {
		sample = content[:i]
	}
	best, bestScore := ',', -1
	for _, c := range candidates {
		r := csv.NewReader(bytes.NewReader(sample))
		r.Comma = c
		r.FieldsPerRecord = -1
		r.LazyQuotes = true
		score, rows := 0, 0
		for rows < 20 {
			rec, err := r.Read()
			if err != nil {
				break
			}
			rows++
			if len(rec) > 1 {
				score += len(rec)
			}
		}
		if score > bestScore {
			best, bestScore = c, score
		}
	}
	return best
}
