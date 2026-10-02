PY ?= .venv-ocr/bin/python
PIP ?= .venv-ocr/bin/pip

.PHONY: build test check fmt vet lint pytest docker clean

build: ## Build the Go server (default, cgo-free)
	go build ./...

build-ocr: ## Build with the Tesseract engine (needs libtesseract + leptonica)
	CGO_ENABLED=1 go build -tags tesseract -o bin/server ./cmd/server

fmt: ## gofmt check
	@test -z "$$(gofmt -l cmd internal)" || (gofmt -l cmd internal; exit 1)

vet: ## go vet
	go vet ./...

lint: ## ruff on the Python package
	$(PY) -m ruff check tableextract

pytest: ## Python unit tests (no tesseract needed)
	$(PY) -m pytest tableextract/tests -q

test: pytest ## All tests
	go test ./... -count=1

check: fmt vet lint pytest ## All static checks + unit tests

docker: ## Build the deployment image
	docker build -t bloom-parser .

extract: ## python -m tableextract extract <out> <imgs...>
	PYTHONPATH=tableextract $(PY) -m tableextract extract $(ARGS)

refine: ## python -m tableextract refine <out>
	PYTHONPATH=tableextract $(PY) -m tableextract refine $(ARGS)

normalize: ## python -m tableextract normalize <out>
	PYTHONPATH=tableextract $(PY) -m tableextract normalize $(ARGS)

serve: ## Run the server with /v1/table-extract wired to the venv
	go build -o bin/server ./cmd/server
	TABLEEXTRACT_PYTHON=$(abspath $(PY)) ./bin/server

clean:
	rm -rf bin out .venv-ocr tableextract/*.egg-info tableextract/**/__pycache__
