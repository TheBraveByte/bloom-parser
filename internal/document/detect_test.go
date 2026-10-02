package document

import "testing"

func TestDetect(t *testing.T) {
	cases := []struct {
		name    string
		content []byte
		want    Format
	}{
		{"png", []byte("\x89PNG\r\n\x1a\n\x00\x00"), FormatPNG},
		{"jpeg", []byte("\xff\xd8\xff\xe0"), FormatJPEG},
		{"webp", []byte("RIFF\x1a\x00\x00\x00WEBPVP8 "), FormatWEBP},
		{"tiff-le", []byte("II*\x00\x08"), FormatTIFF},
		{"tiff-be", []byte("MM\x00*\x00"), FormatTIFF},
		{"gif87", []byte("GIF87a\x01\x00"), FormatGIF},
		{"gif89", []byte("GIF89a\x01\x00"), FormatGIF},
		{"bmp", []byte("BM\x8a\x00\x00\x00"), FormatBMP},
		{"pdf", []byte("%PDF-1.7\n..."), FormatPDF},
		{"xlsx-zip", []byte("PK\x03\x04\x14\x00"), FormatXLSX},
		{"json", []byte(`[{"a":1}]`), FormatJSON},
		{"markdown", []byte("a | b\n--|--\n1 | 2"), FormatMarkdown},
		{"csv", []byte("a,b,c\n1,2,3"), FormatCSV},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Detect(c.content); got != c.want {
				t.Errorf("Detect(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

func TestFormatHelpers(t *testing.T) {
	for _, f := range []Format{FormatPNG, FormatJPEG, FormatWEBP, FormatTIFF, FormatGIF, FormatBMP} {
		if !f.IsImage() {
			t.Errorf("%v should be an image format", f)
		}
	}
	for _, f := range []Format{FormatPDF, FormatXLSX, FormatCSV, FormatJSON, FormatMarkdown} {
		if f.IsImage() {
			t.Errorf("%v should not be an image format", f)
		}
	}
}

func TestDocumentHelpers(t *testing.T) {
	d := &Document{Pages: []Page{
		{Number: 1, Err: &PageError{Code: ErrOCRUnavailable, Message: "x"}},
		{Number: 2},
	}}
	if d.TableCount() != 0 || d.FirstTable() != nil {
		t.Errorf("empty doc should have no tables")
	}
}
