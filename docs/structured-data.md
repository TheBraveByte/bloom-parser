# Structured data & normalization

Raw extraction is not the final result. Every table-producing adapter funnels
its untyped cells through one normalization function so typing and warnings are
identical regardless of source:

```go
parser.NormalizeTable(headers []string, rows [][]string, attrs map[string]string)
    (*parser.ParsedDocument, error)
```

```
Raw extraction (strings)
        │  NormalizeTable
        ▼
Normalized, typed table (ParsedDocument)
        │  validateDocument
        ▼
Validated structured data
```

Extraction logic and cleaning logic are deliberately not mixed: adapters produce
strings; `NormalizeTable` owns all cleaning and typing.

## What normalization does

- **Header cleanup** — blank headers become `column_N` (`MISSING_COLUMN`);
  duplicates become `name_2` (`DUPLICATE_COLUMN`).
- **Row alignment** — short rows are right-padded with nulls (`ROW_TOO_SHORT`);
  long rows are truncated (`ROW_TOO_LONG`).
- **Column typing** — each column's dominant non-null cell type wins among
  `integer`, `decimal`, `boolean`, `datetime`, `string`; ties prefer the
  lossless `string`.
- **Cell conversion** — numbers accept thousands separators (`1,200.50`);
  datetimes accept RFC 3339, `YYYY-MM-DD`, `MM/DD/YYYY` and a few common
  layouts; `yes/no/true/false` become booleans. A cell that cannot convert to
  its column type becomes null and records a `TYPE_MISMATCH` warning — the data
  is never silently dropped.

## Values and null

`parser.Value` is a tagged union (`string`, `int`, `double`, `bool`, `datetime`)
where the null kind represents an absent cell. `Value.String()` renders a stable
canonical form (RFC 3339 for datetimes). The proto `Value` is a `oneof`; an
unset kind is null.

## Warnings vs. errors

Normalization never fails a document for messy data — it records warnings and
keeps the best-effort typed value. Hard failures (no columns, rows wider than
the column list) raise `ErrInvalidDocument`. Per-page extraction failures live
on `Page.Err`; document/request problems are returned as errors and mapped to
gRPC status codes (see [grpc-api.md](grpc-api.md)).
