package service

import (
	"fmt"
	"time"

	pb "github.com/bushadigitallimited/bloom-parser/gen/go/document_parser/v1"
	"github.com/bushadigitallimited/bloom-parser/internal/document"
	"github.com/bushadigitallimited/bloom-parser/internal/parser"
)

func toFormat(f pb.DocumentFormat) (parser.Format, error) {
	switch f {
	case pb.DocumentFormat_DOCUMENT_FORMAT_UNSPECIFIED:
		return parser.FormatAuto, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_MARKDOWN_TABLE:
		return parser.FormatMarkdown, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_CSV:
		return parser.FormatCSV, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_JSON:
		return parser.FormatJSON, nil
	default:
		return parser.FormatAuto, fmt.Errorf("%w: %v", parser.ErrUnsupportedFormat, f)
	}
}

func fromFormat(f parser.Format) pb.DocumentFormat {
	switch f {
	case parser.FormatMarkdown:
		return pb.DocumentFormat_DOCUMENT_FORMAT_MARKDOWN_TABLE
	case parser.FormatCSV:
		return pb.DocumentFormat_DOCUMENT_FORMAT_CSV
	case parser.FormatJSON:
		return pb.DocumentFormat_DOCUMENT_FORMAT_JSON
	default:
		return pb.DocumentFormat_DOCUMENT_FORMAT_UNSPECIFIED
	}
}

// toDocFormat maps a protobuf DocumentFormat to the ingestion format.
// DOCUMENT_FORMAT_UNSPECIFIED becomes FormatUnknown (triggers auto-detection).
func toDocFormat(f pb.DocumentFormat) (document.Format, error) {
	switch f {
	case pb.DocumentFormat_DOCUMENT_FORMAT_UNSPECIFIED:
		return document.FormatUnknown, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_MARKDOWN_TABLE:
		return document.FormatMarkdown, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_CSV:
		return document.FormatCSV, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_JSON:
		return document.FormatJSON, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_PNG:
		return document.FormatPNG, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_JPEG:
		return document.FormatJPEG, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_WEBP:
		return document.FormatWEBP, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_TIFF:
		return document.FormatTIFF, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_GIF:
		return document.FormatGIF, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_BMP:
		return document.FormatBMP, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_PDF:
		return document.FormatPDF, nil
	case pb.DocumentFormat_DOCUMENT_FORMAT_XLSX:
		return document.FormatXLSX, nil
	default:
		return document.FormatUnknown, fmt.Errorf("%w: %v", parser.ErrUnsupportedFormat, f)
	}
}

func fromDocFormat(f document.Format) pb.DocumentFormat {
	switch f {
	case document.FormatMarkdown:
		return pb.DocumentFormat_DOCUMENT_FORMAT_MARKDOWN_TABLE
	case document.FormatCSV:
		return pb.DocumentFormat_DOCUMENT_FORMAT_CSV
	case document.FormatJSON:
		return pb.DocumentFormat_DOCUMENT_FORMAT_JSON
	case document.FormatPNG:
		return pb.DocumentFormat_DOCUMENT_FORMAT_PNG
	case document.FormatJPEG:
		return pb.DocumentFormat_DOCUMENT_FORMAT_JPEG
	case document.FormatWEBP:
		return pb.DocumentFormat_DOCUMENT_FORMAT_WEBP
	case document.FormatTIFF:
		return pb.DocumentFormat_DOCUMENT_FORMAT_TIFF
	case document.FormatGIF:
		return pb.DocumentFormat_DOCUMENT_FORMAT_GIF
	case document.FormatBMP:
		return pb.DocumentFormat_DOCUMENT_FORMAT_BMP
	case document.FormatPDF:
		return pb.DocumentFormat_DOCUMENT_FORMAT_PDF
	case document.FormatXLSX:
		return pb.DocumentFormat_DOCUMENT_FORMAT_XLSX
	default:
		return pb.DocumentFormat_DOCUMENT_FORMAT_UNSPECIFIED
	}
}

func fromPageKind(k document.PageKind) pb.PageKind {
	switch k {
	case document.PageImage:
		return pb.PageKind_PAGE_KIND_IMAGE
	case document.PageSheet:
		return pb.PageKind_PAGE_KIND_SHEET
	default:
		return pb.PageKind_PAGE_KIND_TEXT
	}
}

func toColumnType(t pb.ColumnType) parser.ColumnType {
	switch t {
	case pb.ColumnType_COLUMN_TYPE_INTEGER:
		return parser.TypeInt
	case pb.ColumnType_COLUMN_TYPE_DECIMAL:
		return parser.TypeFloat
	case pb.ColumnType_COLUMN_TYPE_BOOLEAN:
		return parser.TypeBool
	case pb.ColumnType_COLUMN_TYPE_DATETIME:
		return parser.TypeTime
	default:
		return parser.TypeString
	}
}

func fromColumnType(t parser.ColumnType) pb.ColumnType {
	switch t {
	case parser.TypeInt:
		return pb.ColumnType_COLUMN_TYPE_INTEGER
	case parser.TypeFloat:
		return pb.ColumnType_COLUMN_TYPE_DECIMAL
	case parser.TypeBool:
		return pb.ColumnType_COLUMN_TYPE_BOOLEAN
	case parser.TypeTime:
		return pb.ColumnType_COLUMN_TYPE_DATETIME
	default:
		return pb.ColumnType_COLUMN_TYPE_STRING
	}
}

func valueToProto(v parser.Value) *pb.Value {
	switch v.Kind {
	case parser.KindString:
		return &pb.Value{Kind: &pb.Value_StringValue{StringValue: v.Str}}
	case parser.KindInt:
		return &pb.Value{Kind: &pb.Value_IntValue{IntValue: v.Int}}
	case parser.KindFloat:
		return &pb.Value{Kind: &pb.Value_DoubleValue{DoubleValue: v.Float}}
	case parser.KindBool:
		return &pb.Value{Kind: &pb.Value_BoolValue{BoolValue: v.Bool}}
	case parser.KindTime:
		return &pb.Value{Kind: &pb.Value_DatetimeValue{DatetimeValue: v.Time.Format(time.RFC3339)}}
	default:
		return &pb.Value{}
	}
}

func valueFromProto(v *pb.Value) parser.Value {
	switch k := v.GetKind().(type) {
	case *pb.Value_StringValue:
		return parser.Str(k.StringValue)
	case *pb.Value_IntValue:
		return parser.Int(k.IntValue)
	case *pb.Value_DoubleValue:
		return parser.Float(k.DoubleValue)
	case *pb.Value_BoolValue:
		return parser.Bool(k.BoolValue)
	case *pb.Value_DatetimeValue:
		if t, err := time.Parse(time.RFC3339, k.DatetimeValue); err == nil {
			return parser.Time(t)
		}
	}
	return parser.Value{Kind: parser.KindNull}
}

func docToProto(d *parser.ParsedDocument) *pb.ParsedDocument {
	out := &pb.ParsedDocument{
		Name:    d.Name,
		Columns: make([]*pb.Column, len(d.Columns)),
		Rows:    make([]*pb.Row, len(d.Rows)),
		Metadata: &pb.DocumentMetadata{
			DetectedFormat: fromFormat(d.Format),
			RowCount:       int32(len(d.Rows)),
			ColumnCount:    int32(len(d.Columns)),
			Attributes:     d.Attrs,
		},
	}
	for i, c := range d.Columns {
		out.Columns[i] = &pb.Column{Name: c.Name, Type: fromColumnType(c.Type)}
	}
	for i, r := range d.Rows {
		row := &pb.Row{Values: make([]*pb.Value, len(r))}
		for j, v := range r {
			row.Values[j] = valueToProto(v)
		}
		out.Rows[i] = row
	}
	out.Warnings = warningsToProto(d.Warnings)
	return out
}

func warningsToProto(ws []parser.Warning) []*pb.ParseWarning {
	if len(ws) == 0 {
		return nil
	}
	out := make([]*pb.ParseWarning, len(ws))
	for i, w := range ws {
		out[i] = &pb.ParseWarning{
			Severity: pb.ParseWarning_Severity(w.Severity + 1),
			Code:     w.Code,
			Message:  w.Message,
			Row:      int32(w.Row),
			Column:   int32(w.Col),
		}
	}
	return out
}

func documentToProto(d *document.Document) *pb.Document {
	out := &pb.Document{
		Name:       d.Name,
		Format:     fromDocFormat(d.Format),
		Attributes: d.Attrs,
		Warnings:   warningsToProto(d.Warnings),
		Pages:      make([]*pb.Page, len(d.Pages)),
	}
	for i := range d.Pages {
		out.Pages[i] = pageToProto(&d.Pages[i])
	}
	return out
}

func pageToProto(p *document.Page) *pb.Page {
	page := &pb.Page{
		Number:     int32(p.Number),
		Kind:       fromPageKind(p.Kind),
		Source:     p.Source,
		Text:       p.Text,
		Confidence: p.Confidence,
		Warnings:   warningsToProto(p.Warnings),
	}
	for _, t := range p.Tables {
		page.Tables = append(page.Tables, docToProto(t))
	}
	for _, im := range p.Images {
		page.Images = append(page.Images, &pb.ImageInfo{
			Format: im.Format, Width: int32(im.Width), Height: int32(im.Height), ColorModel: im.ColorModel,
		})
	}
	for _, f := range p.Fields {
		page.Fields = append(page.Fields, &pb.Field{
			Key: f.Key, Value: valueToProto(f.Value), Confidence: f.Confidence,
		})
	}
	if p.Err != nil {
		page.Error = &pb.PageError{Code: p.Err.Code, Message: p.Err.Message}
	}
	return page
}

// docFromProto rebuilds a domain document from a client-supplied
// ParsedDocument. Rows longer than the column list are rejected; shorter rows
// are right-padded with nulls.
func docFromProto(p *pb.ParsedDocument) (*parser.ParsedDocument, error) {
	if len(p.GetColumns()) == 0 {
		return nil, fmt.Errorf("%w: no columns", parser.ErrInvalidDocument)
	}
	doc := &parser.ParsedDocument{
		Name:    p.GetName(),
		Columns: make([]parser.Column, len(p.GetColumns())),
		Attrs:   p.GetMetadata().GetAttributes(),
	}
	for i, c := range p.GetColumns() {
		doc.Columns[i] = parser.Column{Name: c.GetName(), Type: toColumnType(c.GetType())}
	}
	for i, pr := range p.GetRows() {
		if len(pr.GetValues()) > len(doc.Columns) {
			return nil, fmt.Errorf("%w: row %d has %d values for %d columns", parser.ErrInvalidDocument, i, len(pr.GetValues()), len(doc.Columns))
		}
		row := make([]parser.Value, len(doc.Columns))
		for j := range row {
			if j < len(pr.GetValues()) {
				row[j] = valueFromProto(pr.GetValues()[j])
			}
		}
		doc.Rows = append(doc.Rows, row)
	}
	return doc, nil
}
