# Node base + bun: vue-tsc must run under real Node — bun's node shim can't
# patch tsc internals, so .vue imports fail to resolve (TS2307).
FROM node:22-bookworm-slim AS webbuild
RUN npm i -g bun
WORKDIR /src
COPY web/package.json web/bun.lock ./web/
RUN cd web && bun install --frozen-lockfile
COPY web/ ./web/
RUN cd web && bun run build

FROM golang:1.26-bookworm AS gobuild
RUN apt-get update && apt-get install -y --no-install-recommends \
    libtesseract-dev libleptonica-dev tesseract-ocr pkg-config \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=webbuild /src/cmd/server/web/dist ./cmd/server/web/dist
RUN CGO_ENABLED=1 go build -tags tesseract -o /out/server ./cmd/server

# pinned to bookworm: trixie renamed liblept5
FROM python:3.12-slim-bookworm AS pybuild
WORKDIR /app
COPY tableextract/ ./tableextract/
RUN pip install --no-cache-dir ./tableextract

FROM python:3.12-slim-bookworm
RUN apt-get update && apt-get install -y --no-install-recommends \
    tesseract-ocr libtesseract5 liblept5 libgl1 libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=pybuild /usr/local /usr/local
COPY --from=gobuild /out/server /usr/local/bin/bloom-server
COPY tableextract/ ./tableextract/
EXPOSE 8090 50051
ENV SERVER_PORT=50051 HTTP_PORT=8090
ENTRYPOINT ["bloom-server"]
