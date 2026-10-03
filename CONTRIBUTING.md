# Contributing to bloom-parser

Thanks for your interest! This document covers how to set up a dev environment,
the checks every change must pass, and the conventions this repo follows. By
participating you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Dev setup

Requirements:

- Go 1.26+ (the Go module lives at the repo root)
- Python 3.10+ and a virtualenv for the `tableextract` package
- Tesseract (`brew install tesseract` / `apt-get install tesseract-ocr`)
  — needed for the OCR build tag and the `tableextract` pipeline
- Optional: `buf` + protoc plugins if you are changing `proto/`

```sh
git clone https://github.com/TheBraveByte/bloom-parser
cd bloom-parser
python3 -m venv .venv-ocr
.venv-ocr/bin/pip install -e "tableextract[dev]"
cp .env.example .env   # fill in only what you need
```

## Workflow

1. Fork the repo and create a branch off `main`.
2. Make your change. Keep it focused — one concern per pull request.
3. Run the full check suite before pushing:

   ```sh
   make check   # gofmt + go vet + ruff + pytest
   make test    # go test ./... + pytest
   ```

   CI runs the same steps on every pull request; please keep it green.

4. Open a pull request with a clear description of what and why.

## Conventions

**Go**

- `gofmt`-clean code; `go vet` must pass.
- Exported API surface carries doc comments; inline comments are kept minimal —
  prefer self-explanatory code and good names.
- Each supported format lives behind `format.Adapter`; the OCR engine behind
  `ocr.Engine`. Extend via interfaces rather than branching on formats.
- Generated code under `gen/` is produced by `buf generate` — never edit it by
  hand; change `proto/` and regenerate instead.

**Python (`tableextract/`)**

- Linted with `ruff` (line length 88, rules E/F/W/I/UP/B/BLE).
- Tests with `pytest` in `tableextract/tests/` — add coverage for new behavior.

**Commits & PRs**

- Write clear, imperative commit messages focused on the "why".
- Do not commit secrets: `.env`, API keys and credentials are gitignored.
- Do not commit large data files or generated artifacts (`out/`, `bin/`).

## Reporting issues

Open a GitHub issue with a minimal reproduction, expected vs actual behavior,
and environment details (OS, Go/Python versions, build tags used).
