// Package ocr abstracts the OCR engine behind an interface. The default build
// ships a disabled engine; the "tesseract" tag swaps in a real one.
package ocr

import (
	"context"
	"errors"
	"image"
)

// ErrUnavailable is returned by the disabled engine (mapped to OCR_UNAVAILABLE).
var ErrUnavailable = errors.New("ocr engine is not configured")

// ConfidenceUnknown marks a result with no reported confidence.
const ConfidenceUnknown = -1

// Options tunes a single recognition call.
type Options struct {
	Languages string
}

// Field is an optional key/value pair some engines expose.
type Field struct {
	Key        string
	Value      string
	Confidence float64
}

// Result is the raw output of recognizing one image.
type Result struct {
	Text       string
	Confidence float64
	Fields     []Field
}

// Engine recognizes text in a decoded image. Implementations wrap one provider.
type Engine interface {
	Name() string
	Available() bool
	// Recognize extracts text from img; it must not mutate img.
	Recognize(ctx context.Context, img image.Image, opts Options) (*Result, error)
}

type disabled struct{}

// Disabled returns an engine that always fails with ErrUnavailable.
func Disabled() Engine { return disabled{} }

// OrDisabled returns e, or the disabled engine when e is nil.
func OrDisabled(e Engine) Engine {
	if e == nil {
		return Disabled()
	}
	return e
}

func (disabled) Name() string    { return "disabled" }
func (disabled) Available() bool { return false }
func (disabled) Recognize(context.Context, image.Image, Options) (*Result, error) {
	return nil, ErrUnavailable
}
