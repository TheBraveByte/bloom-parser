# User guide

How to run bloom-parser and get data out of it, for each interface: the web
console, the REST API, gRPC, and the `tableextract` CLI.

## Setup

```sh
git clone https://github.com/TheBraveByte/bloom-parser
cd bloom-parser

python3 -m venv .venv-ocr
.venv-ocr/bin/pip install -e "tableextract[dev]"

cp .env.example .env     # edit only what you need
```

Requirements:

| Tool | Needed for |
|---|---|
| Go 1.26+ | the server |
| Python 3.10+ | `tableextract` pipeline |
| Tesseract (`brew install tesseract` / `apt-get install tesseract-ocr`) | `/v1/table-extract` and the `tesseract` build tag |
| Bun | building the web console (`make web`) |
| `NVIDIA_API_KEY` in `.env` | optional vision-LLM refinement |

## Run

```sh
make serve     # builds the console, builds bin/server, starts both listeners
```

You should see:

```
grpc server listening        addr=[::]:50051
http/rest gateway listening  addr=:8090
```

`GET /healthz` returns `200` for probes. For a cgo-free build without OCR, use
`make build` and run `bin/server` directly — everything except OCR still works.

## Web console

Open `http://localhost:8090`. The console has two modes:

- **Extract document** — drop an image, PDF, XLSX or text file; get pages of
  text, tables, images and fields back. Toggle OCR per request.
- **Table extract** — drop one or more scanned table images; get a merged
  `normalized.csv` plus per-file stats. The *refine* toggle enables the
  vision-LLM pass (needs `NVIDIA_API_KEY`, noticeably slower).

## REST API

Document bytes are base64 inside JSON. All routes are generated from
`proto/document_parser/v1/parser.proto`.

**Extract any document:**

```sh
curl -s http://localhost:8090/v1/documents:extract -X POST \
  -H 'Content-Type: application/json' \
  -d '{"document":{"name":"report.pdf","content":"<base64>"},
       "options":{"ocr":true,"max_pages":10}}'
```

**Parse tabular text** (CSV, markdown table, JSON rows):

```sh
curl -s http://localhost:8090/v1/documents:parse -X POST \
  -H 'Content-Type: application/json' \
  -d '{"document":{"name":"data.csv","content":"<base64>"}}'
```

**Export** — accepts either a fresh `document` or a `parsed_document` returned
by a previous call:

```sh
curl -s http://localhost:8090/v1/documents:exportCsv -X POST \
  -H 'Content-Type: application/json' \
  -d '{"parsed_document":{"columns":[...],"rows":[...]}}' --output out.csv
```

**Scanned-table extraction** — multipart form, up to 64 files:

```sh
curl -s http://localhost:8090/v1/table-extract -X POST \
  -F 'file=@sheet1.jpg' -F 'file=@sheet2.png' -F 'refine=false'
```

returns `{csv, rows, flagged, files[]}` — the `csv` field is the merged
normalized output for all files.

## gRPC

Reflection is enabled, so `grpcurl` works with no proto files:

```sh
grpcurl -plaintext localhost:50051 list document_parser.v1.DocumentParserService

grpcurl -plaintext -d '{
  "document": {"name": "data.csv", "content": "<base64>"}
}' localhost:50051 document_parser.v1.DocumentParserService/ParseDocument
```

## tableextract CLI

The Python pipeline can also run standalone on a folder of scans:

```sh
make extract   ARGS="out/run1 scans/*.png"   # grid + OCR → CSVs + crops/
make refine    ARGS="out/run1"               # vision-LLM pass (needs NVIDIA_API_KEY)
make normalize ARGS="out/run1"               # → normalized.csv
```

Outputs in `out/run1/`:

| File | Contents |
|---|---|
| `pages.csv` | per-file summary: mode (ruled/borderless), row/cell counts, errors |
| `tables.csv` | raw extracted rows, ragged, with per-row flag column |
| `crops.csv` + `crops/` | low-confidence cells and their image crops |
| `tables_refined.csv`, `refine_log.csv` | post-LLM grid + old→new audit log |
| `normalized.csv` | final facts: `file, row, section, line_item, period, value, raw, flag` |

Rows with a non-empty `flag` are the ones worth eyeballing — the pipeline kept
the value but wasn't confident in it.

## Configuration

Everything is env-driven (see `.env.example`). The common knobs:

| Variable | Default | Purpose |
|---|---|---|
| `SERVER_PORT` / `HTTP_PORT` | `50051` / `8090` | listen ports |
| `MAX_DOCUMENT_BYTES` | 32 MiB | request size cap |
| `MAX_DOCUMENT_PAGES` | 200 | per-document page cap |
| `OCR_LANGUAGES` | `eng` | Tesseract language hint |
| `OCR_DEFAULT` | `false` | run OCR when a request omits options |
| `TABLEEXTRACT_PYTHON` / `_DIR` / `_TIMEOUT` | `python3` / `tableextract` / `3m` | Python subprocess wiring |
| `POWERBI_*` | unset | publish destination; unset = RPC returns `FAILED_PRECONDITION` |
| `NVIDIA_API_KEY` / `LLM_API_KEY` | unset | vision-LLM refinement |

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| Pages report `OCR_UNAVAILABLE` | binary built without `-tags tesseract`; use `make build-ocr` |
| `/v1/table-extract` errors mention `tesseract` | Tesseract binary not installed, or set `TESSDATA_PREFIX` |
| `refine` makes no difference / log shows missing key | `NVIDIA_API_KEY`/`LLM_API_KEY` not set in `.env` |
| Console shows a plain HTML page | frontend not built — run `make web` |
| `413`-ish errors on upload | file exceeds `MAX_DOCUMENT_BYTES` or 64-file batch cap |
| Table timeouts | large batches are 4-at-a-time with a 3m/file budget; raise `TABLEEXTRACT_TIMEOUT` |
