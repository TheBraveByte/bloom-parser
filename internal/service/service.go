package service

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bushadigitallimited/bloom-parser/gen/go/document_parser/v1"
	"github.com/bushadigitallimited/bloom-parser/internal/document"
	"github.com/bushadigitallimited/bloom-parser/internal/export"
	"github.com/bushadigitallimited/bloom-parser/internal/format"
	"github.com/bushadigitallimited/bloom-parser/internal/ingest"
	"github.com/bushadigitallimited/bloom-parser/internal/parser"
	"github.com/bushadigitallimited/bloom-parser/internal/powerbi"
)

var tracer = otel.Tracer("github.com/bushadigitallimited/bloom-parser")

// Server implements DocumentParserServiceServer as thin handlers over the
// parser, exporters and Power BI publisher.
type Server struct {
	pb.UnimplementedDocumentParserServiceServer
	parser     *parser.Service
	ingest     *ingest.Service
	publisher  *powerbi.Publisher
	defaultOCR bool
}

func NewServer(ps *parser.Service, ing *ingest.Service, pub *powerbi.Publisher) *Server {
	return &Server{parser: ps, ingest: ing, publisher: pub}
}

// WithDefaultOCR sets the OCR default for requests that omit options.
func (s *Server) WithDefaultOCR(on bool) *Server {
	s.defaultOCR = on
	return s
}

// ExtractDocument runs the pipeline for any format and returns the Document.
// Observability records metadata only, never document contents.
func (s *Server) ExtractDocument(ctx context.Context, req *pb.ExtractDocumentRequest) (*pb.ExtractDocumentResponse, error) {
	ctx, span := tracer.Start(ctx, "extract.document")
	defer span.End()

	in := req.GetDocument()
	if in == nil || len(in.GetContent()) == 0 {
		return nil, fail(span, parser.ErrEmptyInput)
	}
	f, err := toDocFormat(in.GetFormat())
	if err != nil {
		return nil, fail(span, err)
	}

	// Omitted options inherit the server default; present options win.
	opts := req.GetOptions()
	useOCR := s.defaultOCR
	if opts != nil {
		useOCR = opts.GetOcr()
	}
	start := time.Now()
	doc, err := s.ingest.Ingest(ctx, ingest.Request{
		Name:    in.GetName(),
		Content: in.GetContent(),
		Format:  f,
		Options: format.Options{
			OCR:          useOCR,
			OCRLanguages: opts.GetOcrLanguages(),
			MaxPages:     int(opts.GetMaxPages()),
		},
	})
	if err != nil {
		return nil, fail(span, err)
	}

	failedPages := countFailedPages(doc)
	span.SetAttributes(
		attribute.String("document.format", doc.Format.String()),
		attribute.Int("document.pages", len(doc.Pages)),
		attribute.Int("document.failed_pages", failedPages),
	)
	slog.InfoContext(ctx, "document extracted",
		"format", doc.Format.String(),
		"pages", len(doc.Pages),
		"failed_pages", failedPages,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return &pb.ExtractDocumentResponse{Document: documentToProto(doc)}, nil
}

func countFailedPages(doc *document.Document) int {
	n := 0
	for i := range doc.Pages {
		if doc.Pages[i].Err != nil {
			n++
		}
	}
	return n
}

func (s *Server) ParseDocument(ctx context.Context, req *pb.ParseDocumentRequest) (*pb.ParseDocumentResponse, error) {
	ctx, span := tracer.Start(ctx, "parse.document")
	defer span.End()

	doc, err := s.parse(ctx, req.GetDocument())
	if err != nil {
		return nil, fail(span, err)
	}
	return &pb.ParseDocumentResponse{Document: docToProto(doc)}, nil
}

func (s *Server) ExportCSV(ctx context.Context, req *pb.ExportCSVRequest) (*pb.ExportCSVResponse, error) {
	ctx, span := tracer.Start(ctx, "export.csv")
	defer span.End()

	file, err := s.export(ctx, req.GetSource(), export.CSV{})
	if err != nil {
		return nil, fail(span, err)
	}
	return &pb.ExportCSVResponse{File: file}, nil
}

func (s *Server) ExportXLSX(ctx context.Context, req *pb.ExportXLSXRequest) (*pb.ExportXLSXResponse, error) {
	ctx, span := tracer.Start(ctx, "export.xlsx")
	defer span.End()

	file, err := s.export(ctx, req.GetSource(), export.XLSX{})
	if err != nil {
		return nil, fail(span, err)
	}
	return &pb.ExportXLSXResponse{File: file}, nil
}

func (s *Server) PublishToPowerBI(ctx context.Context, req *pb.PublishToPowerBIRequest) (*pb.PublishToPowerBIResponse, error) {
	ctx, span := tracer.Start(ctx, "powerbi.publish")
	defer span.End()

	if s.publisher == nil {
		return nil, fail(span, powerbi.ErrNotConfigured)
	}
	doc, err := s.resolve(ctx, req.GetSource())
	if err != nil {
		return nil, fail(span, err)
	}
	res, err := s.publisher.Publish(ctx, powerbi.PublishRequest{
		Doc:         doc,
		WorkspaceID: req.GetWorkspaceId(),
		DatasetName: req.GetDatasetName(),
		TableName:   req.GetTableName(),
	})
	if err != nil {
		return nil, fail(span, err)
	}
	return &pb.PublishToPowerBIResponse{
		DatasetId:   res.DatasetID,
		WorkspaceId: res.WorkspaceID,
		TableName:   res.TableName,
		RowsPushed:  res.RowsPushed,
	}, nil
}

// parse validates a DocumentInput and runs the parser.
func (s *Server) parse(ctx context.Context, in *pb.DocumentInput) (*parser.ParsedDocument, error) {
	if in == nil || len(in.GetContent()) == 0 {
		return nil, parser.ErrEmptyInput
	}
	f, err := toFormat(in.GetFormat())
	if err != nil {
		return nil, err
	}
	return s.parser.Parse(ctx, parser.Input{Name: in.GetName(), Content: in.GetContent()}, f)
}

// resolve turns a oneof DataSource into a domain document.
func (s *Server) resolve(ctx context.Context, src *pb.DataSource) (*parser.ParsedDocument, error) {
	if src == nil {
		return nil, parser.ErrEmptyInput
	}
	switch v := src.GetSource().(type) {
	case *pb.DataSource_Document:
		return s.parse(ctx, v.Document)
	case *pb.DataSource_Parsed:
		return docFromProto(v.Parsed)
	default:
		return nil, parser.ErrEmptyInput
	}
}

// export resolves the source and runs one exporter.
func (s *Server) export(ctx context.Context, src *pb.DataSource, e export.Exporter) (*pb.ExportedFile, error) {
	doc, err := s.resolve(ctx, src)
	if err != nil {
		return nil, err
	}
	data, err := e.Export(ctx, doc)
	if err != nil {
		return nil, err
	}
	return &pb.ExportedFile{
		Content:     data,
		Filename:    exportFilename(doc.Name, e.Extension()),
		ContentType: e.ContentType(),
		RowCount:    int64(len(doc.Rows)),
	}, nil
}

func exportFilename(name, ext string) string {
	base := strings.TrimSuffix(filepath.Base(strings.TrimSpace(name)), filepath.Ext(name))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "export"
	}
	return base + ext
}

// fail records the error on the span and maps it to a gRPC status.
func fail(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return toStatus(err)
}

func toStatus(err error) error {
	var code gcodes.Code
	switch {
	case errors.Is(err, parser.ErrEmptyInput),
		errors.Is(err, parser.ErrInvalidDocument),
		errors.Is(err, parser.ErrUnsupportedFormat),
		errors.Is(err, ingest.ErrNoPages):
		code = gcodes.InvalidArgument
	case errors.Is(err, ingest.ErrTooLarge):
		code = gcodes.ResourceExhausted
	case errors.Is(err, parser.ErrNotParseable):
		code = gcodes.FailedPrecondition
	case errors.Is(err, powerbi.ErrNotConfigured):
		code = gcodes.FailedPrecondition
	case errors.Is(err, powerbi.ErrUnauthenticated):
		code = gcodes.Unauthenticated
	case errors.Is(err, powerbi.ErrPermissionDenied):
		code = gcodes.PermissionDenied
	case errors.Is(err, powerbi.ErrUnavailable):
		code = gcodes.Unavailable
	default:
		code = gcodes.Internal
	}
	return status.Error(code, err.Error())
}
