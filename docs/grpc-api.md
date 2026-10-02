# gRPC & REST API

The protobuf definition in `proto/document_parser/v1/parser.proto` is the single
source of truth. `buf generate` produces the gRPC stubs, the grpc-gateway REST
proxy and the OpenAPI/Swagger document — none of these are hand-maintained.

## Service

| RPC | REST | Purpose |
|---|---|---|
| `ExtractDocument` | `POST /v1/documents:extract` | Ingest any supported format → `Document` (pages, text, tables, images, fields). |
| `ParseDocument` | `POST /v1/documents:parse` | Tabular text → single `ParsedDocument`. |
| `ExportCSV` | `POST /v1/documents:exportCsv` | `DataSource` → CSV bytes. |
| `ExportXLSX` | `POST /v1/documents:exportXlsx` | `DataSource` → XLSX bytes. |
| `PublishToPowerBI` | `POST /v1/documents:publishPowerbi` | Publish a table to a Power BI push dataset. |

`ExtractDocument` is the primary entrypoint for images, PDFs and spreadsheets.
`ParseDocument`, `ExportCSV/XLSX` and `PublishToPowerBI` are unchanged from the
previous API (backward compatible); they gained only HTTP annotations.

## Key messages

- `DocumentInput{format, content, name}` — raw bytes; `content` is base64 in
  JSON/REST.
- `ExtractOptions{ocr, ocr_languages, max_pages}`.
- `Document{name, format, pages[], warnings[], attributes}`.
- `Page{number, kind, source, text, tables[], images[], fields[], confidence,
  warnings[], error}` — `confidence` is `-1` when unknown; `error` is set only
  when the page failed.
- `ParsedDocument{name, columns[], rows[], warnings[], metadata}` — the typed
  table reused inside pages and by the export/publish endpoints.

## Error model

Domain errors map to gRPC status codes (and HTTP statuses via the gateway):

| Condition | gRPC code |
|---|---|
| missing/empty input, invalid document, unsupported/unknown format, no pages | `INVALID_ARGUMENT` |
| content not decodable as the detected format | `FAILED_PRECONDITION` |
| input exceeds the size limit | `RESOURCE_EXHAUSTED` |
| Power BI unconfigured | `FAILED_PRECONDITION` |
| Power BI auth / permission / unavailable | `UNAUTHENTICATED` / `PERMISSION_DENIED` / `UNAVAILABLE` |
| anything else | `INTERNAL` |

Errors are specific and actionable ("page 3 failed to decode", "OCR extraction
failed", "format unsupported"), never "something went wrong". Per-page failures
are not top-level errors — they are carried on `Page.error` so the rest of the
document is still returned. Credentials and document bytes never appear in
errors or logs.

## OpenAPI / Swagger

`buf generate` writes `gen/openapiv2/bloom-parser.swagger.json` from the proto
(endpoints, request/response schemas, enums). Serve or import it in any
OpenAPI UI. Regenerate with `buf generate`.

## Calling it

```sh
# REST (reflection-free)
curl -s -X POST http://localhost:8080/v1/documents:extract \
  -H 'Content-Type: application/json' \
  -d '{"document":{"name":"scan.png","content":"<base64 bytes>"},"options":{"ocr":true}}'

# gRPC (reflection enabled)
grpcurl -plaintext localhost:50051 list document_parser.v1.DocumentParserService
```
