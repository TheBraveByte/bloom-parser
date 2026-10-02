// Package image adapts raster images (PNG, JPEG, WEBP, TIFF, GIF, BMP) into one
// document page, running the OCR engine when extraction is requested.
package image

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"

	// Register the supported decoders.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// Adapter decodes images and optionally recognizes text via an OCR engine.
type Adapter struct {
	engine ocr.Engine
}

func New(engine ocr.Engine) *Adapter {
	return &Adapter{engine: ocr.OrDisabled(engine)}
}

func (a *Adapter) Adapt(ctx context.Context, in format.Input) ([]document.Page, error) {
	img, encFmt, err := image.Decode(bytes.NewReader(in.Content))
	if err != nil {
		return nil, fmt.Errorf("%w: decode %s image: %v", parser.ErrNotParseable, in.Format, err)
	}

	b := img.Bounds()
	page := document.Page{
		Number:     1,
		Kind:       document.PageImage,
		Source:     format.Source(in.Name, "image"),
		Confidence: document.ConfidenceUnknown,
		Images: []document.ImageInfo{{
			Format:     encFmt,
			Width:      b.Dx(),
			Height:     b.Dy(),
			ColorModel: colorModelName(img),
		}},
	}

	if in.Options.OCR {
		a.recognize(ctx, img, in.Options, &page)
	}
	return []document.Page{page}, nil
}

func (a *Adapter) recognize(ctx context.Context, img image.Image, opts format.Options, page *document.Page) {
	res, err := a.engine.Recognize(ctx, img, ocr.Options{Languages: opts.OCRLanguages})
	if err != nil {
		code := document.ErrOCRFailed
		if errors.Is(err, ocr.ErrUnavailable) {
			code = document.ErrOCRUnavailable
		}
		page.Err = &document.PageError{Code: code, Message: err.Error()}
		return
	}
	page.Text = res.Text
	page.Confidence = res.Confidence
	for _, f := range res.Fields {
		page.Fields = append(page.Fields, document.Field{
			Key:        f.Key,
			Value:      parser.Str(f.Value),
			Confidence: f.Confidence,
		})
	}
}

func colorModelName(img image.Image) string {
	if _, ok := img.ColorModel().(color.Palette); ok {
		return "palette"
	}
	switch img.ColorModel() {
	case color.GrayModel, color.Gray16Model:
		return "gray"
	case color.RGBAModel, color.RGBA64Model, color.NRGBAModel, color.NRGBA64Model:
		return "rgba"
	case color.CMYKModel:
		return "cmyk"
	case color.YCbCrModel, color.NYCbCrAModel:
		return "ycbcr"
	default:
		return "other"
	}
}
