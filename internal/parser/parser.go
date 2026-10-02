package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Parser extracts a raw document of one format into a ParsedDocument.
type Parser interface {
	Format() Format
	Parse(ctx context.Context, in Input) (*ParsedDocument, error)
}

// Service routes inputs to the right parser and owns format auto-detection.
type Service struct {
	parsers map[Format]Parser
}

func NewService() *Service {
	all := []Parser{markdownParser{}, csvParser{}, jsonParser{}}
	m := make(map[Format]Parser, len(all))
	for _, p := range all {
		m[p.Format()] = p
	}
	return &Service{parsers: m}
}

func (s *Service) Parse(ctx context.Context, in Input, f Format) (*ParsedDocument, error) {
	if len(bytes.TrimSpace(in.Content)) == 0 {
		return nil, ErrEmptyInput
	}
	if f == FormatAuto {
		f = Detect(in.Content)
	}
	p, ok := s.parsers[f]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, f)
	}
	doc, err := p.Parse(ctx, in)
	if err != nil {
		return nil, err
	}
	doc.Format = f
	if doc.Name == "" {
		doc.Name = in.Name
	}
	return doc, nil
}

// Detect sniffs the document format. Order: JSON, markdown table, else CSV.
func Detect(content []byte) Format {
	c := bytes.TrimSpace(content)
	switch {
	case json.Valid(c) && len(c) > 0 && c[0] == '[':
		return FormatJSON
	case looksLikeMarkdownTable(c):
		return FormatMarkdown
	default:
		return FormatCSV
	}
}

func looksLikeMarkdownTable(c []byte) bool {
	lines := strings.Split(string(c), "\n")
	for i := 0; i+1 < len(lines); i++ {
		if strings.Contains(lines[i], "|") && isSeparatorRow(lines[i+1]) {
			return true
		}
	}
	return false
}

// isSeparatorRow matches markdown separator rows like "|---|---|" or "| :- | -: |".
func isSeparatorRow(line string) bool {
	l := strings.TrimSpace(line)
	if l == "" || !strings.Contains(l, "-") {
		return false
	}
	for _, r := range l {
		if r != '|' && r != '-' && r != ':' && r != ' ' {
			return false
		}
	}
	return true
}
