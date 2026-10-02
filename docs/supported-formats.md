# Supported formats

| Format | `DocumentFormat` | Detected by | Produces |
|---|---|---|---|
| PNG | `PNG` | `\x89PNG\r\n\x1a\n` | one image page + metadata (+ OCR text when enabled) |
| JPEG | `JPEG` | `FF D8 FF` | one image page |
| WebP | `WEBP` | `RIFF….WEBP` | one image page |
| TIFF | `TIFF` | `II*\0` / `MM\0*` | one image page (first frame) |
| GIF | `GIF` | `GIF87a` / `GIF89a` | one image page (first frame) |
| BMP | `BMP` | `BM` + size field | one image page |
| PDF | `PDF` | `%PDF-` | one page per PDF page (native text or scanned) |
| XLSX | `XLSX` | ZIP magic `PK\x03\x04` | one page per sheet, each a typed table |
| Markdown table | `MARKDOWN_TABLE` | header row + `---` separator | one text page + one table |
| Delimited text | `CSV` | comma/semicolon/tab/pipe auto-detected | one text page + one table |
| JSON | `JSON` | leading `[` and valid JSON array | one text page + one table |

Auto-detection runs when `DOCUMENT_FORMAT_UNSPECIFIED` is sent. A declared
format is authoritative.

## Images

Decoding is pure Go: `image/png`, `image/jpeg`, `image/gif` (stdlib) and
`golang.org/x/image/{webp,tiff,bmp}`. Each page records `format`, `width`,
`height` and a logical `color_model`. Pixel data is never retained past
recognition. OCR on image pages runs when `options.ocr` is set, or by default
when the server is started with `OCR_DEFAULT=true` and a request omits options.

## PDF

Native text is extracted per page (`github.com/ledongthuc/pdf`). A page with no
extractable text is treated as **scanned** and marked `PAGE_KIND_IMAGE`:

- with no OCR engine → `OCR_UNAVAILABLE`;
- with an OCR engine → `PAGE_UNREADABLE`, because turning a scanned page into
  pixels needs a rasterizer (see Limitations).

## XLSX

Read with `github.com/xuri/excelize/v2` (already used by the XLSX exporter).
Each sheet becomes a page; its cells are normalized into a typed table with the
sheet name preserved. Structured spreadsheet data is **never** routed through
OCR.

## Limitations / deliberate trade-offs

- **No OCR backend in the default build.** OCR is abstracted behind
  `ocr.Engine`; real recognition requires a `-tags tesseract` build with
  libtesseract installed. Without it, OCR requests return a structured
  `OCR_UNAVAILABLE` per-page error.
- **Scanned-PDF rasterization is not bundled.** Pure-Go PDF rendering is not
  available; rasterizing vector/scanned PDF pages into images for OCR requires
  an external renderer (pdfium/mupdf, cgo). The seam exists (scanned pages are
  routed as image pages); only the rasterizer is pluggable-and-absent.
- **Multi-frame TIFF** yields the first frame (the x/image decoder exposes one).
