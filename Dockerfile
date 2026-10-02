FROM oven/bun:1 AS webbuild
WORKDIR /src
COPY web/package.json web/bun.lock ./web/
RUN cd web && bun install --frozen-lockfile
COPY web/ ./web/
COPY cmd/server/web/dist/.keep ./cmd/server/web/dist/.keep
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
RUN CGO_ENABLED=1 go build -tags tesseract -o /out/server ./cmd/server \
    && CGO_ENABLED=0 go build -o /out/ocrtest ./cmd/ocrtest

FROM python:3.12-slim AS pybuild
WORKDIR /app
COPY tableextract/ ./tableextract/
RUN pip install --no-cache-dir ./tableextract

FROM python:3.12-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    tesseract-ocr libtesseract5 liblept5 libgl1 libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=pybuild /usr/local /usr/local
COPY --from=gobuild /out/server /usr/local/bin/bloom-server
COPY --from=gobuild /out/ocrtest /usr/local/bin/ocrtest
COPY tableextract/ ./tableextract/
EXPOSE 8080 50051
ENV SERVER_PORT=50051 HTTP_PORT=8080
ENTRYPOINT ["bloom-server"]
