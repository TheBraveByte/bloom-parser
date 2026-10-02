# Extraction & the OCR engine

Extraction is the stage inside each adapter that turns decoded input into text,
tables and fields. It is kept separate from normalization (cleaning/typing).

| Adapter | Extraction strategy |
|---|---|
| image | decode → metadata; OCR via `ocr.Engine` when `Options.OCR` is set |
| pdf | native text layer per page; scanned pages routed to OCR |
| xlsx | read cells directly (no OCR) |
| text | delegate to `parser.Service` (csv/json/markdown) |

## The OCR abstraction

```go
type Engine interface {
    Name() string
    Available() bool
    Recognize(ctx context.Context, img image.Image, opts Options) (*Result, error)
}
```

The pipeline depends only on this interface. Nothing else in the codebase
imports an OCR provider.

### Engines

- **`ocr.Disabled()`** (default) — `Available()` is false and `Recognize`
  returns `ErrUnavailable`. It lets the full pipeline (decode, validation,
  per-page error handling) run and be tested with no OCR backend. This mirrors
  how Power BI is optional: a missing capability yields a structured,
  actionable error rather than a crash.
- **Tesseract** (`-tags tesseract`) — wraps `github.com/otiai10/gosseract/v2`
  (cgo + libtesseract). Lives in `engine_tesseract.go`, excluded from the
  default build so the module compiles and tests cleanly without the native
  dependency.

`ocr.Configured(opts)` returns the engine selected at build time, so
`cmd/server` wires the engine without referencing a concrete provider.

### How failures surface

A decoded image whose OCR step fails keeps its metadata and gains a per-page
error:

| Condition | `Page.Err.Code` |
|---|---|
| OCR requested, engine disabled | `OCR_UNAVAILABLE` |
| OCR requested, engine errored | `OCR_FAILED` |
| scanned PDF page, no rasterizer | `PAGE_UNREADABLE` |
| image bytes undecodable | whole input fails with `ErrNotParseable` |

## Building with Tesseract

```sh
# macOS:  brew install tesseract
# Debian: apt-get install -y libtesseract-dev libleptonica-dev
CGO_ENABLED=1 go build -tags tesseract -o bin/server ./cmd/server
```

Set `OCR_LANGUAGES` (e.g. `eng+deu`) to pass a language hint.
