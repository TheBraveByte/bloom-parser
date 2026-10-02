package export

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/TheBraveByte/bloom-parser/internal/parser"
)

const dateTimeNumFmt = "yyyy-mm-dd hh:mm:ss"

// XLSX renders a ParsedDocument as a real .xlsx workbook: bold header row,
// typed cells (numbers numeric, booleans boolean, datetimes date-formatted).
type XLSX struct{}

func (XLSX) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}
func (XLSX) Extension() string { return ".xlsx" }

var invalidSheetChars = regexp.MustCompile(`[\\/*?\[\]:]`)

func (XLSX) Export(_ context.Context, doc *parser.ParsedDocument) ([]byte, error) {
	if err := validate(doc); err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	defer f.Close()

	sheet := sheetName(doc.Name)
	if err := f.SetSheetName(f.GetSheetName(0), sheet); err != nil {
		return nil, fmt.Errorf("set sheet name: %w", err)
	}

	dateStyle, err := f.NewStyle(&excelize.Style{CustomNumFmt: ptr(dateTimeNumFmt)})
	if err != nil {
		return nil, fmt.Errorf("create date style: %w", err)
	}
	boldStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, fmt.Errorf("create header style: %w", err)
	}

	for c, col := range doc.Columns {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		if err := f.SetCellStr(sheet, cell, col.Name); err != nil {
			return nil, fmt.Errorf("write header %q: %w", col.Name, err)
		}
	}
	if err := f.SetCellStyle(sheet, "A1", lastCell(len(doc.Columns), 1), boldStyle); err != nil {
		return nil, fmt.Errorf("style header: %w", err)
	}

	for r, row := range doc.Rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := setCell(f, sheet, cell, v, dateStyle); err != nil {
				return nil, fmt.Errorf("write cell %s: %w", cell, err)
			}
		}
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("serialize xlsx: %w", err)
	}
	return buf.Bytes(), nil
}

func setCell(f *excelize.File, sheet, cell string, v parser.Value, dateStyle int) error {
	switch v.Kind {
	case parser.KindNull:
		return nil
	case parser.KindString:
		return f.SetCellStr(sheet, cell, v.Str)
	case parser.KindInt:
		return f.SetCellInt(sheet, cell, v.Int)
	case parser.KindFloat:
		return f.SetCellFloat(sheet, cell, v.Float, -1, 64)
	case parser.KindBool:
		return f.SetCellBool(sheet, cell, v.Bool)
	case parser.KindTime:
		if err := f.SetCellValue(sheet, cell, v.Time); err != nil {
			return err
		}
		return f.SetCellStyle(sheet, cell, cell, dateStyle)
	default:
		return f.SetCellStr(sheet, cell, v.String())
	}
}

func sheetName(name string) string {
	n := invalidSheetChars.ReplaceAllString(strings.TrimSpace(name), " ")
	n = strings.TrimSpace(n)
	if n == "" {
		return "Data"
	}
	if len(n) > 31 {
		n = n[:31]
	}
	return n
}

func lastCell(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func ptr[T any](v T) *T { return &v }
