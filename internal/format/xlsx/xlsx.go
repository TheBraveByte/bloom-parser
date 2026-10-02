// Package xlsx adapts a workbook into one page per sheet, reading cells
// directly (never OCR) and typing them via parser.NormalizeTable.
package xlsx

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/TheBraveByte/bloom-parser/internal/document"
	"github.com/TheBraveByte/bloom-parser/internal/format"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

// Adapter reads workbooks with excelize.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (Adapter) Adapt(_ context.Context, in format.Input) ([]document.Page, error) {
	f, err := excelize.OpenReader(bytes.NewReader(in.Content))
	if err != nil {
		return nil, fmt.Errorf("%w: open workbook: %v", parser.ErrNotParseable, err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("%w: workbook has no sheets", parser.ErrNotParseable)
	}

	pages := make([]document.Page, 0, len(sheets))
	for i, name := range sheets {
		if in.Options.MaxPages > 0 && i >= in.Options.MaxPages {
			break
		}
		pages = append(pages, sheetPage(f, name, i+1))
	}
	return pages, nil
}

func sheetPage(f *excelize.File, name string, number int) document.Page {
	page := document.Page{Number: number, Kind: document.PageSheet, Source: name, Confidence: document.ConfidenceUnknown}

	rows, err := f.GetRows(name)
	if err != nil {
		page.Err = &document.PageError{Code: document.ErrSheetUnreadable, Message: err.Error()}
		return page
	}
	if len(rows) == 0 {
		page.Warnings = append(page.Warnings, parser.Warning{
			Severity: parser.SeverityInfo, Code: "EMPTY_SHEET",
			Message: fmt.Sprintf("sheet %q has no rows", name), Row: -1, Col: -1,
		})
		return page
	}

	ncols := 0
	for _, r := range rows {
		if len(r) > ncols {
			ncols = len(r)
		}
	}
	headers := make([]string, ncols)
	copy(headers, rows[0])

	table, err := parser.NormalizeTable(headers, rows[1:], map[string]string{"sheet": name})
	if err != nil {
		page.Err = &document.PageError{Code: document.ErrSheetUnreadable, Message: err.Error()}
		return page
	}
	table.Name = name
	page.Tables = append(page.Tables, table)
	return page
}
