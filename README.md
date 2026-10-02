# bloom-parser

Modular document ingestion & structured-data extraction, exposed over gRPC and
REST. One pipeline accepts images, PDFs, spreadsheets and tabular text, and
returns a common `Document`; tabular results can be exported to CSV / XLSX or
published to Power BI.

```
          Upload / Request
                 │
                 ▼
        validate → detect format → adapter → pages → validate
                 │                   │
                 │        ┌──────────┼───────────┬───────────┐
                 │        ▼          ▼           ▼           ▼
                 │     image       pdf         xlsx        text
                 │       │ (OCR engine: disabled | tesseract | …)
                 ▼       ▼
              ┌─────────────────┐
              │    Document     │  pages → text · tables · images · fields
              └───────┬─────────┘
            (tables)  │
              ┌───────┴────────┐
              ▼                ▼
        CSV / XLSX        Power BI
```

The core pipeline never knows a format's internals — each format lives behind an
adapter, and the OCR engine lives behind an interface. Adding a format is one
adapter plus one registration. See [docs/architecture.md](docs/architecture.md).

Docs: [how it works](docs/how-it-works.md) (with the end-to-end flow chart) ·
[user guide](docs/user-guide.md) · [architecture](docs/architecture.md) ·
[extraction](docs/extraction.md) · [gRPC/REST API](docs/grpc-api.md) ·
[parsing pipeline](docs/parsing-pipeline.md) · [supported formats](docs/supported-formats.md)

## Layout

```
cmd/server             entrypoint (config, logging, tracing, gRPC + REST gateway)
internal/document      common representation (Document/Page/…) + format detection
internal/parser        typed table primitives, tabular extractors, NormalizeTable
internal/ocr           OCR Engine interface, disabled default, tesseract (build tag)
internal/format        Adapter contract
internal/format/{image,pdf,xlsx,text}   one adapter per format
internal/ingest        pipeline orchestrator + validation + limits
internal/export        CSV + XLSX exporters
internal/powerbi       Power BI push-dataset publisher
internal/service       thin gRPC handlers + proto<->domain conversion
internal/testfixtures  runtime-generated image/pdf/xlsx bytes for tests
proto/                 API contract (source of truth)
gen/                   buf-generated gRPC, REST gateway, OpenAPI
docs/                  architecture & implementation docs
```

## Prerequisites

- Go 1.26+
- `buf` — `go install github.com/bufbuild/buf/cmd/buf@latest`
- Codegen plugins on `PATH`: `protoc-gen-go`, `protoc-gen-go-grpc`,
  `protoc-gen-grpc-gateway`, `protoc-gen-openapiv2`
- Optional: `grpcurl` for manual gRPC calls; libtesseract for OCR (see below)

## Common tasks

```sh
make help          # list all targets
make build         # compile all Go packages (cgo-free)
make build-ocr     # build bin/server with the Tesseract engine
make check         # gofmt + go vet + ruff + pytest
make test          # go test + pytest
make serve         # build + run: gRPC :SERVER_PORT, REST :HTTP_PORT
make docker        # build the deployment image
make extract ARGS="out/run1 img1.png img2.png"   # table extraction
make refine ARGS="out/run1"      # vision-LLM cell refinement (needs NVIDIA_API_KEY)
make normalize ARGS="out/run1"   # emit normalized.csv facts
```

Regenerate gRPC/REST/OpenAPI code after editing `proto/`:

```sh
buf generate
```

## Configuration

Copy `.env.example` to `.env`. Everything is env-driven.

| Variable | Purpose |
|---|---|
| `SERVER_PORT` | gRPC port (default `50051`) |
| `HTTP_PORT` | REST gateway port (default `8080`) |
| `LOG_FORMAT` / `LOG_LEVEL` | `json`/`text`, `debug`…`error` |
| `OTEL_ENABLED` | `true` → stdout span exporter |
| `MAX_DOCUMENT_BYTES` | request size cap (default 32 MiB) |
| `MAX_DOCUMENT_PAGES` | default page/sheet cap (default 200) |
| `OCR_LANGUAGES` | OCR language hint (tesseract build only) |
| `POWERBI_*` | Entra service principal + default workspace |

## Supported formats

Images (PNG, JPEG, WebP, TIFF), PDF (native text + scanned pages), XLSX, and
tabular text (markdown tables, delimited, JSON). Format is auto-detected from
content, or declared explicitly. Full matrix and limitations:
[docs/supported-formats.md](docs/supported-formats.md).

## API

Primary RPC, `ExtractDocument`, runs the full pipeline for any format:

```sh
# REST
curl -s -X POST http://localhost:8080/v1/documents:extract \
  -H 'Content-Type: application/json' \
  -d '{"document":{"name":"scan.png","content":"<base64 bytes>"},"options":{"ocr":true}}'

# gRPC (reflection enabled)
grpcurl -plaintext localhost:50051 list document_parser.v1.DocumentParserService
```

`ParseDocument`, `ExportCSV`, `ExportXLSX` and `PublishToPowerBI` remain for
tabular parse/export/publish flows. Error model, messages and the generated
OpenAPI doc (`gen/openapiv2/bloom-parser.swagger.json`) are described in
[docs/grpc-api.md](docs/grpc-api.md).

## OCR

The default build has no OCR backend: OCR requests return a structured
`OCR_UNAVAILABLE` per-page error, and every non-OCR path works normally. Enable
real recognition with the `tesseract` build tag:

```sh
# macOS: brew install tesseract | Debian: apt-get install libtesseract-dev libleptonica-dev
CGO_ENABLED=1 go build -tags tesseract -o bin/server ./cmd/server
```

See [docs/extraction.md](docs/extraction.md).

## Extending

- **New format**: implement `format.Adapter`, register it in `ingest.New`, add
  it to `document.Detect` and the `DocumentFormat` enum. Normalization,
  validation, exporters and the API are untouched.
- **New OCR provider**: implement `ocr.Engine`; nothing else changes.
- **New destination**: implement `export.Exporter` or call the `powerbi.Client`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
