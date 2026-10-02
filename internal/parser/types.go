package parser

import (
	"strconv"
	"time"
)

// Format identifies the shape of raw document content.
type Format int

const (
	FormatAuto Format = iota
	FormatMarkdown
	FormatCSV
	FormatJSON
)

func (f Format) String() string {
	switch f {
	case FormatMarkdown:
		return "markdown"
	case FormatCSV:
		return "csv"
	case FormatJSON:
		return "json"
	default:
		return "auto"
	}
}

// ColumnType is the normalized logical type of a column.
type ColumnType int

const (
	TypeNone ColumnType = iota // null/unset, never a final column type
	TypeString
	TypeInt
	TypeFloat
	TypeBool
	TypeTime
)

type Column struct {
	Name string
	Type ColumnType
}

type ValueKind int

const (
	KindNull ValueKind = iota
	KindString
	KindInt
	KindFloat
	KindBool
	KindTime
)

// Value is a single typed cell. KindNull represents an absent value.
type Value struct {
	Kind  ValueKind
	Str   string
	Int   int64
	Float float64
	Bool  bool
	Time  time.Time
}

func Str(s string) Value    { return Value{Kind: KindString, Str: s} }
func Int(i int64) Value     { return Value{Kind: KindInt, Int: i} }
func Float(f float64) Value { return Value{Kind: KindFloat, Float: f} }
func Bool(b bool) Value     { return Value{Kind: KindBool, Bool: b} }
func Time(t time.Time) Value {
	return Value{Kind: KindTime, Time: t}
}

// String renders the value in a stable canonical form (RFC 3339 for
// datetimes). Null renders as "".
func (v Value) String() string {
	switch v.Kind {
	case KindString:
		return v.Str
	case KindInt:
		return strconv.FormatInt(v.Int, 10)
	case KindFloat:
		return strconv.FormatFloat(v.Float, 'f', -1, 64)
	case KindBool:
		return strconv.FormatBool(v.Bool)
	case KindTime:
		return v.Time.Format(time.RFC3339)
	default:
		return ""
	}
}

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// Warning is a non-fatal parse issue. Row/Col are 0-based, -1 when N/A.
type Warning struct {
	Severity Severity
	Code     string
	Message  string
	Row      int
	Col      int
}

// ParsedDocument is the normalized source of truth consumed by every
// exporter/destination. Rows are positional: Rows[i][j] belongs to Columns[j].
type ParsedDocument struct {
	Name     string
	Columns  []Column
	Rows     [][]Value
	Warnings []Warning
	Format   Format
	Attrs    map[string]string
}

type Input struct {
	Name    string
	Content []byte
}
