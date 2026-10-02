// Package testfixtures generates realistic document bytes (images, PDFs,
// workbooks) at runtime so tests do not depend on committed binary blobs.
package testfixtures

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func solid(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 128, A: 255})
		}
	}
	return img
}

func encodeImage(w, h int, enc func(io.Writer, image.Image) error) []byte {
	var buf bytes.Buffer
	_ = enc(&buf, solid(w, h))
	return buf.Bytes()
}

// PNG returns a w×h PNG.
func PNG(w, h int) []byte { return encodeImage(w, h, png.Encode) }

// BMP returns a w×h BMP.
func BMP(w, h int) []byte { return encodeImage(w, h, bmp.Encode) }

// JPEG returns a w×h JPEG.
func JPEG(w, h int) []byte {
	return encodeImage(w, h, func(wr io.Writer, m image.Image) error {
		return jpeg.Encode(wr, m, &jpeg.Options{Quality: 80})
	})
}

// TIFF returns a w×h TIFF.
func TIFF(w, h int) []byte {
	return encodeImage(w, h, func(wr io.Writer, m image.Image) error { return tiff.Encode(wr, m, nil) })
}

// GIF returns a w×h GIF.
func GIF(w, h int) []byte {
	return encodeImage(w, h, func(wr io.Writer, m image.Image) error { return gif.Encode(wr, m, nil) })
}

// WebP returns a minimal valid 1×1 WebP. x/image has no WebP encoder, so this
// is a known-good blob used to exercise the decode path.
func WebP() []byte {
	b, _ := base64.StdEncoding.DecodeString("UklGRhoAAABXRUJQVlA4TA0AAAAvAAAAEAcQERGIiP4HAA==")
	return b
}

// XLSX builds a workbook. sheets maps in the given order to row data; the first
// row of each is treated as a header by the adapter.
func XLSX(order []string, sheets map[string][][]any) []byte {
	f := excelize.NewFile()
	defer f.Close()
	for i, name := range order {
		if i == 0 {
			f.SetSheetName(f.GetSheetName(0), name)
		} else {
			f.NewSheet(name)
		}
		for r, row := range sheets[name] {
			for c, v := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
				f.SetCellValue(name, cell, v)
			}
		}
	}
	var buf bytes.Buffer
	_, _ = f.WriteTo(&buf)
	return buf.Bytes()
}

// PDFText builds a valid PDF with one page per string, each rendered as a text
// layer. For a page with no extractable text use ScannedPDF.
func PDFText(texts ...string) []byte {
	contents := make([]string, len(texts))
	for i, t := range texts {
		contents[i] = fmt.Sprintf("BT /F1 24 Tf 72 700 Td (%s) Tj ET", escapePDF(t))
	}
	return buildPDF(contents)
}

// ScannedPDF builds a valid single-page PDF whose content stream draws no text,
// simulating a scanned page with no extractable text layer.
func ScannedPDF() []byte {
	return buildPDF([]string{"q 1 0 0 1 0 0 cm 100 100 200 200 re f Q"})
}

// MalformedPDF returns bytes with a PDF header but a corrupt body.
func MalformedPDF() []byte {
	return []byte("%PDF-1.4\nthis is not a valid pdf body\n%%EOF")
}

func escapePDF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	return strings.ReplaceAll(s, ")", "\\)")
}

func buildPDF(pageContents []string) []byte {
	objs := make([]string, 3)
	add := func(body string) int {
		objs = append(objs, body)
		return len(objs)
	}

	kids := make([]string, 0, len(pageContents))
	for _, content := range pageContents {
		stream := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content)
		contentNum := add(stream)
		pageNum := add(fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>",
			contentNum))
		kids = append(kids, fmt.Sprintf("%d 0 R", pageNum))
	}
	objs[0] = "<< /Type /Catalog /Pages 2 0 R >>"
	objs[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pageContents))
	objs[2] = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objs)+1)
	for i, o := range objs {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for i := 1; i <= len(objs); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objs)+1, xref)
	return buf.Bytes()
}
