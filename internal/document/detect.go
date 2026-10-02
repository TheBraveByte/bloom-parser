package document

import (
	"bytes"

	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// Detect matches binary formats by magic number, else falls back to the
// tabular text detector. Zip content is reported as XLSX; the adapter validates.
func Detect(content []byte) Format {
	c := bytes.TrimSpace(content)
	switch {
	case hasPrefix(content, "\x89PNG\r\n\x1a\n"):
		return FormatPNG
	case hasPrefix(content, "\xff\xd8\xff"):
		return FormatJPEG
	case isWebP(content):
		return FormatWEBP
	case isTIFF(content):
		return FormatTIFF
	case hasPrefix(content, "GIF87a"), hasPrefix(content, "GIF89a"):
		return FormatGIF
	case isBMP(content):
		return FormatBMP
	case hasPrefix(c, "%PDF-"):
		return FormatPDF
	case hasPrefix(content, "PK\x03\x04"), hasPrefix(content, "PK\x05\x06"):
		return FormatXLSX
	}
	switch parser.Detect(content) {
	case parser.FormatJSON:
		return FormatJSON
	case parser.FormatMarkdown:
		return FormatMarkdown
	default:
		return FormatCSV
	}
}

func hasPrefix(b []byte, s string) bool { return bytes.HasPrefix(b, []byte(s)) }

func isWebP(b []byte) bool {
	return len(b) >= 12 && bytes.Equal(b[0:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP"))
}

func isTIFF(b []byte) bool {
	return hasPrefix(b, "II*\x00") || hasPrefix(b, "MM\x00*")
}

func isBMP(b []byte) bool {
	return len(b) >= 6 && b[0] == 'B' && b[1] == 'M'
}
