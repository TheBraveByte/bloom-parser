package parser

import "errors"

var (
	ErrEmptyInput        = errors.New("document content is empty")
	ErrUnsupportedFormat = errors.New("unsupported document format")
	ErrNotParseable      = errors.New("content is not parseable")
	ErrInvalidDocument   = errors.New("invalid parsed document")
)
