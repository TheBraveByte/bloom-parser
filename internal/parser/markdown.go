package parser

import (
	"context"
	"fmt"
	"strings"
)

// markdownParser extracts the first pipe table found in a markdown document.
type markdownParser struct{}

func (markdownParser) Format() Format { return FormatMarkdown }

func (markdownParser) Parse(_ context.Context, in Input) (*ParsedDocument, error) {
	lines := strings.Split(string(in.Content), "\n")

	start, sep := -1, -1
	for i := 0; i+1 < len(lines); i++ {
		if strings.Contains(lines[i], "|") && isSeparatorRow(lines[i+1]) {
			start, sep = i, i+1
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("%w: no markdown table found", ErrNotParseable)
	}

	var rows [][]string
	for _, l := range lines[sep+1:] {
		if !strings.Contains(l, "|") {
			break
		}
		cells := splitPipeRow(l)
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}
	return normalize(rawTable{headers: splitPipeRow(lines[start]), rows: rows})
}

// splitPipeRow splits "| a | b |" into ["a","b"], tolerating missing
// leading/trailing pipes and escaped cells kept verbatim.
func splitPipeRow(line string) []string {
	l := strings.TrimSpace(line)
	l = strings.TrimPrefix(l, "|")
	l = strings.TrimSuffix(l, "|")
	parts := strings.Split(l, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}
