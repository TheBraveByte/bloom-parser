# Testing

Run everything:

```sh
make test          # go test ./...
make vet           # go vet ./...
```

The suite is fully offline and deterministic — no network, no OCR binary, no
committed binary fixtures.

## Fixtures

`internal/testfixtures` generates realistic bytes at runtime:

- `PNG/JPEG/TIFF(w,h)` — encoded with the stdlib / `x/image` encoders.
- `WebP()` — a known-good minimal blob (x/image has no WebP encoder).
- `XLSX(order, sheets)` — built with excelize.
- `PDFText(pages…)` — a structurally valid multi-page PDF with a correct xref
  table and a real text layer.
- `ScannedPDF()` — a valid page whose content stream draws no text (simulates a
  scan with no text layer).
- `MalformedPDF()` — a PDF header with a corrupt body.

## Coverage

| Area | Tests |
|---|---|
| detection | magic-byte + text fallback for every format |
| image | PNG/JPEG/WEBP/TIFF decode + metadata; malformed; empty; OCR disabled / success / error (stub engine) |
| pdf | text, multi-page, max-pages, scanned (no OCR / needs rasterizer), malformed |
| xlsx | single/multi sheet, max-pages cap, empty sheet, malformed workbook, ragged rows, typed cells |
| pipeline | auto-detect every format end-to-end; isolated page failure; request validation (empty, too large); explicit-format override; per-format malformed input |
| service (gRPC) | `ExtractDocument` for image/xlsx/scanned-PDF over bufconn; error codes; existing parse/export/publish tests |
| export, powerbi, parser | pre-existing suites, unchanged |

## OCR tests

`ocr.Disabled()` is exercised throughout. OCR success/error/field paths are
tested with an in-package stub engine (`internal/format/image/image_test.go`),
so the recognition contract is verified without libtesseract.

To test the real engine, install libtesseract and build/run with the tag:

```sh
CGO_ENABLED=1 go test -tags tesseract ./...
```

## Manual end-to-end

```sh
make serve                     # gRPC :50051, REST :8090
curl -X POST localhost:8090/v1/documents:extract \
  -H 'Content-Type: application/json' \
  -d '{"document":{"content":"<base64>"}}'
```
