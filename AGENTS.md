# bloom-parser

Go gRPC/REST document-extraction service + a Python table-extraction toolkit.

## Layout

- `cmd/server` — gRPC + REST gateway server (`web/` test console embedded)
- `cmd/ocrtest` — batch image-ingestion runner (Go pipeline)
- `internal/` — document model, format adapters, ingest pipeline, ocr engine,
  parser, service layer, testfixtures
- `tableextract/` — Python package: OpenCV grid detection + Tesseract OCR +
  vision-LLM cell refinement + schema normalization

## Commands

- `make build` — Go build (cgo-free). `make build-ocr` adds the Tesseract
  engine (needs `libtesseract-dev` + `libleptonica-dev`, or on macOS
  `CGO_CFLAGS="-I/opt/homebrew/include" CGO_CXXFLAGS="-I/opt/homebrew/include"
  CGO_LDFLAGS="-L/opt/homebrew/lib"` with brew tesseract installed)
- `make check` — gofmt + go vet + ruff + pytest
- `make test` — go test + pytest
- `make docker` — build the deployment image
- `make extract ARGS="out/<dir> <imgs...>"` — table extraction
- `make refine ARGS="out/<dir>"` — LLM refinement (needs `NVIDIA_API_KEY` in
  env or `.env`; uses OpenAI-compatible NIM endpoint, override `URL`/
  `MODEL_FAST`/`MODEL_SLOW` in `tableextract/refine.py` for other providers)
- `make normalize ARGS="out/<dir>"` — emit `normalized.csv` facts
  (file, section, line_item, period, value)

## Notes

- Tesseract binary required for `extract` (`brew install tesseract` /
  `apt-get install tesseract-ocr`); set `TESSDATA_PREFIX` if tessdata is not
  found.
- `out/`, `.venv-ocr/`, `bin/` are gitignored artifacts.
