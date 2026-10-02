# Parsing pipeline

`ingest.Service.Ingest` runs every request through the same stages. Each stage
has one responsibility.

```
Request{name, content, format?, options}
   │
   ▼  validateRequest        content present; within MaxBytes
   ▼  detect format          document.Detect(content) when format is Unknown
   ▼  select adapter         adapters[format]; unknown → ErrUnsupportedFormat
   ▼  adapt                  adapter.Adapt → []document.Page
   ▼  validateDocument       ≥1 page; every table well-formed
   ▼
Document{name, format, pages[], warnings, attrs}
```

## Per-stage detail

- **validateRequest** — empty content → `ErrEmptyInput`; oversize → `ErrTooLarge`
  (gRPC `RESOURCE_EXHAUSTED`).
- **detect** — magic-byte sniffing for binary formats (PNG/JPEG/WEBP/TIFF/PDF/
  ZIP→XLSX), falling back to the tabular text detector. An explicit
  `format` skips detection and is authoritative (forcing the wrong format fails
  loudly rather than silently re-detecting).
- **adapt** — the only format-aware stage. Returns one page per image, per PDF
  page, or per sheet. A whole-input failure (corrupt container) returns an
  error; a failure confined to one page is recorded in `Page.Err` so siblings
  survive.
- **extraction** happens inside adapters: native PDF text, spreadsheet cells, or
  OCR on images. OCR runs only when `Options.OCR` is set.
- **normalization** happens when an adapter builds a table: `parser.NormalizeTable`
  cleans headers, aligns rows, infers column types and converts cells. See
  [structured-data.md](structured-data.md).
- **validateDocument** — a defensive, cross-cutting check that the document has
  at least one page and every table's rows fit its columns.

## Multi-page aggregation

Pages keep their 1-based `Number`, a `Source` label (`page-3`, a sheet name, an
image name), their own `Warnings` and an optional `Err`. Nothing is flattened:
callers can see precisely which page failed and why without losing the rest of
the document.

## Resource safety

- `Limits.MaxBytes` bounds request size before any decoding.
- `Limits.MaxPages` (and per-request `Options.MaxPages`) caps how many pages /
  sheets are processed, bounding work for large PDFs and workbooks.
- Pages are processed sequentially — no unbounded fan-out of goroutines per
  page.
- Adapters stream from in-memory readers and never write temporary files.
