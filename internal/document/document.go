// Package document is the common representation every input format is adapted
// into. Tabular data reuses parser.ParsedDocument.
package document

import "github.com/TheBraveByte/bloom-parser/internal/parser"

// Format identifies a supported source format.
type Format int

const (
	FormatUnknown Format = iota
	FormatPNG
	FormatJPEG
	FormatWEBP
	FormatTIFF
	FormatPDF
	FormatXLSX
	FormatCSV
	FormatJSON
	FormatMarkdown
	FormatGIF
	FormatBMP
)

func (f Format) String() string {
	switch f {
	case FormatPNG:
		return "png"
	case FormatJPEG:
		return "jpeg"
	case FormatWEBP:
		return "webp"
	case FormatTIFF:
		return "tiff"
	case FormatGIF:
		return "gif"
	case FormatBMP:
		return "bmp"
	case FormatPDF:
		return "pdf"
	case FormatXLSX:
		return "xlsx"
	case FormatCSV:
		return "csv"
	case FormatJSON:
		return "json"
	case FormatMarkdown:
		return "markdown"
	default:
		return "unknown"
	}
}

// IsImage reports whether the format is a raster image.
func (f Format) IsImage() bool {
	switch f {
	case FormatPNG, FormatJPEG, FormatWEBP, FormatTIFF, FormatGIF, FormatBMP:
		return true
	default:
		return false
	}
}

// PageKind describes what a page primarily carries.
type PageKind int

const (
	PageText PageKind = iota
	PageImage
	PageSheet
)

func (k PageKind) String() string {
	switch k {
	case PageImage:
		return "image"
	case PageSheet:
		return "sheet"
	default:
		return "text"
	}
}

// Document is the result of ingesting one input: one page per image, sheet or
// PDF page.
type Document struct {
	Name     string
	Format   Format
	Pages    []Page
	Warnings []parser.Warning
	Attrs    map[string]string
}

// Page is one unit of a document. A failed page carries Err; siblings survive.
type Page struct {
	Number     int
	Kind       PageKind
	Source     string
	Text       string
	Tables     []*parser.ParsedDocument
	Images     []ImageInfo
	Fields     []Field
	Confidence float64
	Warnings   []parser.Warning
	Err        *PageError
}

// ConfidenceUnknown marks a page/field whose engine reported no confidence.
const ConfidenceUnknown = -1

// ImageInfo records decoded image metadata without retaining pixel data.
type ImageInfo struct {
	Format     string
	Width      int
	Height     int
	ColorModel string
}

// Field is a single extracted key/value pair (e.g. an OCR form field).
type Field struct {
	Key        string
	Value      parser.Value
	Confidence float64
}

// PageError is a per-page failure with a stable Code and a byte-free Message.
type PageError struct {
	Code    string
	Message string
}

func (e *PageError) Error() string { return e.Code + ": " + e.Message }

// Stable page-error codes.
const (
	ErrOCRUnavailable  = "OCR_UNAVAILABLE"
	ErrOCRFailed       = "OCR_FAILED"
	ErrPageUnreadable  = "PAGE_UNREADABLE"
	ErrSheetUnreadable = "SHEET_UNREADABLE"
)

// FailedPage builds a page that carries only an error.
func FailedPage(number int, source, code, msg string) Page {
	return Page{Number: number, Source: source, Err: &PageError{Code: code, Message: msg}}
}

// TableCount returns the number of tables across all pages.
func (d *Document) TableCount() int {
	n := 0
	for i := range d.Pages {
		n += len(d.Pages[i].Tables)
	}
	return n
}

// FirstTable returns the first table in page order, or nil. Bridges a Document
// to the single-table exporters and Power BI publisher.
func (d *Document) FirstTable() *parser.ParsedDocument {
	for i := range d.Pages {
		if len(d.Pages[i].Tables) > 0 {
			return d.Pages[i].Tables[0]
		}
	}
	return nil
}
