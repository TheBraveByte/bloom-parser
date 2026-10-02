package image

import (
	"context"
	"errors"
	"image"
	"testing"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
	"github.com/TheBraveByte/bloom-parser/internal/testfixtures"
)

type stubEngine struct {
	avail bool
	res   *ocr.Result
	err   error
}

func (s stubEngine) Name() string    { return "stub" }
func (s stubEngine) Available() bool { return s.avail }
func (s stubEngine) Recognize(context.Context, image.Image, ocr.Options) (*ocr.Result, error) {
	return s.res, s.err
}

func adapt(t *testing.T, a *Adapter, content []byte, f document.Format, opts format.Options) document.Page {
	t.Helper()
	pages, err := a.Adapt(context.Background(), format.Input{Content: content, Format: f, Options: opts})
	if err != nil {
		t.Fatalf("adapt: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(pages))
	}
	return pages[0]
}

func TestDecodeFormats(t *testing.T) {
	a := New(ocr.Disabled())
	cases := []struct {
		name    string
		content []byte
		format  document.Format
		encFmt  string
	}{
		{"png", testfixtures.PNG(40, 20), document.FormatPNG, "png"},
		{"jpeg", testfixtures.JPEG(40, 20), document.FormatJPEG, "jpeg"},
		{"tiff", testfixtures.TIFF(40, 20), document.FormatTIFF, "tiff"},
		{"webp", testfixtures.WebP(), document.FormatWEBP, "webp"},
		{"gif", testfixtures.GIF(40, 20), document.FormatGIF, "gif"},
		{"bmp", testfixtures.BMP(40, 20), document.FormatBMP, "bmp"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			page := adapt(t, a, c.content, c.format, format.Options{})
			if page.Kind != document.PageImage {
				t.Errorf("kind = %v", page.Kind)
			}
			if len(page.Images) != 1 {
				t.Fatalf("images = %d", len(page.Images))
			}
			if page.Images[0].Format != c.encFmt {
				t.Errorf("enc format = %q, want %q", page.Images[0].Format, c.encFmt)
			}
			if page.Images[0].Width == 0 || page.Images[0].Height == 0 {
				t.Errorf("dimensions = %dx%d", page.Images[0].Width, page.Images[0].Height)
			}
		})
	}
}

func TestMalformedImage(t *testing.T) {
	a := New(ocr.Disabled())
	_, err := a.Adapt(context.Background(), format.Input{Content: []byte("\x89PNGnot-an-image"), Format: document.FormatPNG})
	if !errors.Is(err, parser.ErrNotParseable) {
		t.Errorf("err = %v, want ErrNotParseable", err)
	}
}

func TestEmptyImage(t *testing.T) {
	a := New(ocr.Disabled())
	_, err := a.Adapt(context.Background(), format.Input{Content: nil, Format: document.FormatPNG})
	if !errors.Is(err, parser.ErrNotParseable) {
		t.Errorf("err = %v, want ErrNotParseable", err)
	}
}

func TestOCRDisabled(t *testing.T) {
	a := New(ocr.Disabled())
	page := adapt(t, a, testfixtures.PNG(20, 20), document.FormatPNG, format.Options{OCR: true})
	if page.Err == nil || page.Err.Code != document.ErrOCRUnavailable {
		t.Fatalf("err = %v, want OCR_UNAVAILABLE", page.Err)
	}
	if len(page.Images) != 1 {
		t.Error("image metadata lost on OCR failure")
	}
}

func TestOCRSuccess(t *testing.T) {
	eng := stubEngine{avail: true, res: &ocr.Result{Text: "hello", Confidence: 0.9, Fields: []ocr.Field{{Key: "k", Value: "v", Confidence: 0.8}}}}
	a := New(eng)
	page := adapt(t, a, testfixtures.PNG(20, 20), document.FormatPNG, format.Options{OCR: true})
	if page.Err != nil {
		t.Fatalf("unexpected err: %v", page.Err)
	}
	if page.Text != "hello" || page.Confidence != 0.9 {
		t.Errorf("text=%q conf=%v", page.Text, page.Confidence)
	}
	if len(page.Fields) != 1 || page.Fields[0].Key != "k" || page.Fields[0].Value.Str != "v" {
		t.Errorf("fields = %v", page.Fields)
	}
}

func TestOCRError(t *testing.T) {
	eng := stubEngine{avail: true, err: errors.New("boom")}
	a := New(eng)
	page := adapt(t, a, testfixtures.PNG(20, 20), document.FormatPNG, format.Options{OCR: true})
	if page.Err == nil || page.Err.Code != document.ErrOCRFailed {
		t.Fatalf("err = %v, want OCR_FAILED", page.Err)
	}
}
