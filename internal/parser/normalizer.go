package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type rawTable struct {
	headers  []string
	rows     [][]string
	attrs    map[string]string
	warnings []Warning
}

var timeLayouts = []string{
	time.RFC3339,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"01/02/2006",
	"1/2/2006",
	"2006/01/02",
	"02-Jan-2006",
	"Jan 2, 2006",
}

// NormalizeTable types an untyped string table; the shared entry point other
// format adapters use so normalization is never duplicated.
func NormalizeTable(headers []string, rows [][]string, attrs map[string]string) (*ParsedDocument, error) {
	return normalize(rawTable{headers: headers, rows: rows, attrs: attrs})
}

func normalize(raw rawTable) (*ParsedDocument, error) {
	if len(raw.headers) == 0 {
		return nil, ErrNotParseable
	}
	doc := &ParsedDocument{Attrs: raw.attrs, Warnings: raw.warnings}
	doc.Columns = normalizeHeaders(raw.headers, &doc.Warnings)

	ncol := len(doc.Columns)
	rawRows := make([][]string, len(raw.rows))
	for i, r := range raw.rows {
		rawRows[i] = alignRow(r, ncol, i, &doc.Warnings)
	}

	for c := 0; c < ncol; c++ {
		doc.Columns[c].Type = inferColumnType(rawRows, c)
	}

	doc.Rows = make([][]Value, len(rawRows))
	for i, r := range rawRows {
		row := make([]Value, ncol)
		for c := 0; c < ncol; c++ {
			v, ok := convert(r[c], doc.Columns[c].Type)
			if !ok {
				doc.Warnings = append(doc.Warnings, Warning{
					Severity: SeverityWarning,
					Code:     "TYPE_MISMATCH",
					Message:  fmt.Sprintf("value %q does not match column type %d", r[c], doc.Columns[c].Type),
					Row:      i,
					Col:      c,
				})
			}
			row[c] = v
		}
		doc.Rows[i] = row
	}
	return doc, nil
}

func normalizeHeaders(headers []string, warnings *[]Warning) []Column {
	cols := make([]Column, len(headers))
	seen := map[string]int{}
	for i, h := range headers {
		name := strings.TrimSpace(h)
		if name == "" {
			name = fmt.Sprintf("column_%d", i+1)
			*warnings = append(*warnings, Warning{Severity: SeverityInfo, Code: "MISSING_COLUMN", Message: fmt.Sprintf("empty header at position %d renamed to %q", i, name), Row: -1, Col: i})
		}
		if n, dup := seen[name]; dup {
			name = fmt.Sprintf("%s_%d", name, n+1)
			*warnings = append(*warnings, Warning{Severity: SeverityInfo, Code: "DUPLICATE_COLUMN", Message: fmt.Sprintf("duplicate header renamed to %q", name), Row: -1, Col: i})
		}
		seen[name]++
		cols[i] = Column{Name: name}
	}
	return cols
}

func alignRow(r []string, ncol, rowIdx int, warnings *[]Warning) []string {
	switch {
	case len(r) < ncol:
		*warnings = append(*warnings, Warning{Severity: SeverityWarning, Code: "ROW_TOO_SHORT", Message: fmt.Sprintf("row has %d cells, expected %d; padded with nulls", len(r), ncol), Row: rowIdx, Col: -1})
		out := make([]string, ncol)
		copy(out, r)
		return out
	case len(r) > ncol:
		*warnings = append(*warnings, Warning{Severity: SeverityWarning, Code: "ROW_TOO_LONG", Message: fmt.Sprintf("row has %d cells, expected %d; extras dropped", len(r), ncol), Row: rowIdx, Col: -1})
		return r[:ncol]
	default:
		return r
	}
}

func inferColumnType(rows [][]string, col int) ColumnType {
	var nInt, nFloat, nBool, nTime, nStr int
	for _, r := range rows {
		switch classify(r[col]) {
		case TypeInt:
			nInt++
		case TypeFloat:
			nFloat++
		case TypeBool:
			nBool++
		case TypeTime:
			nTime++
		case TypeString:
			nStr++
		}
	}
	best, bestN := TypeString, nStr
	if nTime > bestN {
		best, bestN = TypeTime, nTime
	}
	if nBool > bestN {
		best, bestN = TypeBool, nBool
	}
	if n := nInt + nFloat; n > bestN {
		best = TypeInt
		if nFloat > 0 {
			best = TypeFloat
		}
	}
	return best
}

func classify(s string) ColumnType {
	s = strings.TrimSpace(s)
	if s == "" {
		return TypeNone
	}
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return TypeInt
	}
	if _, err := parseNumber(s); err == nil {
		return TypeFloat
	}
	if _, ok := parseBoolText(s); ok {
		return TypeBool
	}
	if _, ok := parseTime(s); ok {
		return TypeTime
	}
	return TypeString
}

func convert(s string, t ColumnType) (Value, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Value{Kind: KindNull}, true
	}
	switch t {
	case TypeInt:
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return Int(i), true
		}
	case TypeFloat:
		if f, err := parseNumber(s); err == nil {
			return Float(f), true
		}
	case TypeBool:
		if b, ok := parseBoolText(s); ok {
			return Bool(b), true
		}
	case TypeTime:
		if tm, ok := parseTime(s); ok {
			return Time(tm), true
		}
	case TypeString:
		return Str(s), true
	}
	return Value{Kind: KindNull}, false
}

func parseNumber(s string) (float64, error) {
	if strings.Contains(s, ",") {
		if !validThousands(s) {
			return 0, strconv.ErrSyntax
		}
		s = strings.ReplaceAll(s, ",", "")
	}
	return strconv.ParseFloat(s, 64)
}

func validThousands(s string) bool {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ",")
	if len(parts) < 2 || !isNumericText(parts[0]) || len(strings.TrimLeft(parts[0], "+-")) > 3 {
		return false
	}
	for _, p := range parts[1:] {
		if len(p) != 3 || !allDigits(p) {
			return false
		}
	}
	return true
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func parseBoolText(s string) (bool, bool) {
	switch strings.ToLower(s) {
	case "true", "yes":
		return true, true
	case "false", "no":
		return false, true
	}
	return false, false
}

func parseTime(s string) (time.Time, bool) {
	if len(s) < 6 {
		return time.Time{}, false
	}
	for _, l := range timeLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func isNumericText(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) && r != '.' && r != '-' && r != '+' {
			return false
		}
	}
	return len(s) > 0
}
