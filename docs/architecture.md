# Architecture

Bloom Parser ingests documents of many formats through **one pipeline** and
returns a single common representation. The core pipeline knows nothing about
individual file formats; each format lives behind an adapter.

```
Request ─▶ validate ─▶ detect format ─▶ adapter ─▶ pages ─▶ validate ─▶ Document
                                           │
                 ┌─────────────────────────┼───────────────────────────┐
                 ▼             ▼            ▼            ▼               ▼
            image adapter  pdf adapter  xlsx adapter  text adapter   (future: docx…)
                 │
                 ▼
            OCR engine (pluggable: disabled | tesseract | …)
```

## Packages and dependency direction

Dependencies point inward; no cycles.

| Package | Responsibility |
|---|---|
| `internal/document` | The common representation: `Document`, `Page`, `Table`, `ImageInfo`, `Field`, `PageError`, `Format` + `Detect`. Leaf (imports only `parser`). |
| `internal/parser` | Typed tabular primitives (`ParsedDocument`, `Column`, `Value`, `Warning`), tabular text extractors (csv/json/markdown) and the shared `NormalizeTable`. |
| `internal/ocr` | `Engine` interface, `Disabled` default, build-tagged Tesseract engine. Leaf. |
| `internal/format` | The `Adapter` contract (`Input`, `Options`, `Adapter`). |
| `internal/format/{image,pdf,xlsx,text}` | One adapter per format. Import `document`, `format`, `parser`, and (image/pdf) `ocr`. |
| `internal/ingest` | Pipeline orchestrator + request/document validation + resource limits. Registers adapters. |
| `internal/export` | `Exporter` (CSV, XLSX) over a `ParsedDocument`. Unchanged. |
| `internal/powerbi` | Power BI push-dataset publisher. Unchanged. |
| `internal/service` | Thin gRPC handlers + proto↔domain conversion. |
| `cmd/server` | Wiring: config, logging, tracing, gRPC server + REST gateway. |

## Why `ParsedDocument` is reused for tables

The existing `parser.ParsedDocument` is exactly a typed table (columns + typed
rows + warnings). A `document.Page` holds zero or more of them, so exporters and
the Power BI publisher keep operating on a single `ParsedDocument` unchanged,
while the richer `Document` wraps pages, images, text and fields around it.
`Document.FirstTable()` bridges the two.

## Extending with a new format

1. Implement `format.Adapter` in `internal/format/<name>`.
2. Register it in `ingest.New` for its `document.Format` value(s).
3. Add the format to `document.Detect` and the `DocumentFormat` proto enum.

Normalization, validation, the gRPC/REST surface and exporters are unchanged.

## OCR is replaceable

The pipeline depends only on `ocr.Engine`. The default build ships
`ocr.Disabled()`; building with `-tags tesseract` swaps in a real engine. No
other package references an OCR provider. See [extraction.md](extraction.md).
