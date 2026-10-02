//go:build tesseract

// Real OCR engine; compiles only with `-tags tesseract` (needs libtesseract + cgo).
package ocr

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"

	"github.com/otiai10/gosseract/v2"
)

type tesseract struct {
	languages string
}

// Configured returns a Tesseract-backed engine.
func Configured(opts Options) Engine {
	lang := opts.Languages
	if lang == "" {
		lang = "eng"
	}
	return &tesseract{languages: lang}
}

func (t *tesseract) Name() string    { return "tesseract" }
func (t *tesseract) Available() bool { return true }

func (t *tesseract) Recognize(ctx context.Context, img image.Image, opts Options) (*Result, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode image for ocr: %w", err)
	}

	client := gosseract.NewClient()
	defer client.Close()

	lang := opts.Languages
	if lang == "" {
		lang = t.languages
	}
	if err := client.SetLanguage(splitLangs(lang)...); err != nil {
		return nil, fmt.Errorf("set ocr language: %w", err)
	}
	if err := client.SetImageFromBytes(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("load image into ocr: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("ocr recognition: %w", err)
	}
	return &Result{Text: text, Confidence: meanConfidence(client)}, nil
}

func meanConfidence(c *gosseract.Client) float64 {
	boxes, err := c.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil || len(boxes) == 0 {
		return ConfidenceUnknown
	}
	var sum float64
	for _, b := range boxes {
		sum += b.Confidence
	}
	return sum / float64(len(boxes)) / 100.0
}

func splitLangs(s string) []string {
	var out []string
	for _, p := range bytes.Split([]byte(s), []byte("+")) {
		if len(p) > 0 {
			out = append(out, string(p))
		}
	}
	if len(out) == 0 {
		out = []string{"eng"}
	}
	return out
}
