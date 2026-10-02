//go:build !tesseract

package ocr

// Configured returns the build-time engine; the default build has none.
// Build with `-tags tesseract` for real recognition.
func Configured(_ Options) Engine { return Disabled() }
