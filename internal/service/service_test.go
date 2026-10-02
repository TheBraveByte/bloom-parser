package service

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/TheBraveByte/bloom-parser/gen/go/document_parser/v1"
	"github.com/TheBraveByte/bloom-parser/internal/ingest"
	"github.com/TheBraveByte/bloom-parser/internal/ocr"
	"github.com/TheBraveByte/bloom-parser/internal/parser"
	"github.com/TheBraveByte/bloom-parser/internal/powerbi"
	"github.com/TheBraveByte/bloom-parser/internal/testfixtures"
)

func dial(t *testing.T, pub *powerbi.Publisher) pb.DocumentParserServiceClient {
	return dialServer(t, NewServer(parser.NewService(), ingest.New(ocr.Disabled(), ingest.DefaultLimits), pub))
}

func dialServer(t *testing.T, handler *Server) pb.DocumentParserServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	pb.RegisterDocumentParserServiceServer(srv, handler)
	go srv.Serve(lis)
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return pb.NewDocumentParserServiceClient(conn)
}

const sampleMD = `
| name  | amount | when       |
| ----- | ------ | ---------- |
| John  | 1200.5 | 2024-03-01 |
| Alice | 99     | 2024-04-02 |
`

func grpcCode(err error) codes.Code {
	s, _ := status.FromError(err)
	return s.Code()
}

func TestParseDocumentRPC(t *testing.T) {
	c := dial(t, nil)
	resp, err := c.ParseDocument(context.Background(), &pb.ParseDocumentRequest{
		Document: &pb.DocumentInput{Content: []byte(sampleMD), Name: "sales"},
	})
	if err != nil {
		t.Fatal(err)
	}
	doc := resp.GetDocument()
	if doc.GetMetadata().GetDetectedFormat() != pb.DocumentFormat_DOCUMENT_FORMAT_MARKDOWN_TABLE {
		t.Errorf("format = %v", doc.GetMetadata().GetDetectedFormat())
	}
	if doc.GetMetadata().GetRowCount() != 2 || doc.GetMetadata().GetColumnCount() != 3 {
		t.Errorf("metadata = %v", doc.GetMetadata())
	}
	if doc.GetColumns()[1].GetType() != pb.ColumnType_COLUMN_TYPE_DECIMAL {
		t.Errorf("amount type = %v", doc.GetColumns()[1].GetType())
	}
	if got := doc.GetRows()[0].GetValues()[0].GetStringValue(); got != "John" {
		t.Errorf("cell = %q", got)
	}
	if got := doc.GetRows()[0].GetValues()[2].GetDatetimeValue(); got != "2024-03-01T00:00:00Z" {
		t.Errorf("datetime = %q", got)
	}
}

func TestExtractDocumentImageRPC(t *testing.T) {
	c := dial(t, nil)
	resp, err := c.ExtractDocument(context.Background(), &pb.ExtractDocumentRequest{
		Document: &pb.DocumentInput{Content: testfixtures.PNG(40, 20), Name: "scan.png"},
	})
	if err != nil {
		t.Fatal(err)
	}
	doc := resp.GetDocument()
	if doc.GetFormat() != pb.DocumentFormat_DOCUMENT_FORMAT_PNG {
		t.Errorf("format = %v", doc.GetFormat())
	}
	if len(doc.GetPages()) != 1 || doc.GetPages()[0].GetKind() != pb.PageKind_PAGE_KIND_IMAGE {
		t.Fatalf("pages = %v", doc.GetPages())
	}
	img := doc.GetPages()[0].GetImages()
	if len(img) != 1 || img[0].GetFormat() != "png" || img[0].GetWidth() != 40 {
		t.Errorf("image info = %v", img)
	}
}

func TestExtractDocumentXLSXRPC(t *testing.T) {
	c := dial(t, nil)
	wb := testfixtures.XLSX([]string{"Sales", "Costs"}, map[string][][]any{
		"Sales": {{"name", "amount"}, {"Alice", 10.5}},
		"Costs": {{"item", "qty"}, {"pen", 3}},
	})
	resp, err := c.ExtractDocument(context.Background(), &pb.ExtractDocumentRequest{
		Document: &pb.DocumentInput{Content: wb, Name: "book.xlsx"},
	})
	if err != nil {
		t.Fatal(err)
	}
	pages := resp.GetDocument().GetPages()
	if len(pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(pages))
	}
	if pages[0].GetSource() != "Sales" || len(pages[0].GetTables()) != 1 {
		t.Errorf("sheet 0 = %+v", pages[0])
	}
	if got := pages[0].GetTables()[0].GetColumns()[1].GetType(); got != pb.ColumnType_COLUMN_TYPE_DECIMAL {
		t.Errorf("amount type = %v", got)
	}
}

func TestExtractDocumentScannedPDFPageError(t *testing.T) {
	c := dial(t, nil)
	resp, err := c.ExtractDocument(context.Background(), &pb.ExtractDocumentRequest{
		Document: &pb.DocumentInput{Content: testfixtures.ScannedPDF(), Name: "scan.pdf"},
	})
	if err != nil {
		t.Fatal(err)
	}
	pages := resp.GetDocument().GetPages()
	if len(pages) != 1 || pages[0].GetError().GetCode() != "OCR_UNAVAILABLE" {
		t.Fatalf("expected OCR_UNAVAILABLE page error, got %+v", pages)
	}
}

func TestExtractDocumentDefaultOCR(t *testing.T) {
	c := dialServer(t, NewServer(parser.NewService(), ingest.New(ocr.Disabled(), ingest.DefaultLimits), nil).WithDefaultOCR(true))
	resp, err := c.ExtractDocument(context.Background(), &pb.ExtractDocumentRequest{
		Document: &pb.DocumentInput{Content: testfixtures.PNG(20, 20), Name: "x.png"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := resp.GetDocument().GetPages()[0].GetError().GetCode(); got != "OCR_UNAVAILABLE" {
		t.Errorf("default OCR not applied: page error = %q", got)
	}
}

func TestExtractDocumentGIF(t *testing.T) {
	c := dial(t, nil)
	resp, err := c.ExtractDocument(context.Background(), &pb.ExtractDocumentRequest{
		Document: &pb.DocumentInput{Content: testfixtures.GIF(24, 24), Name: "a.gif"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetDocument().GetFormat() != pb.DocumentFormat_DOCUMENT_FORMAT_GIF {
		t.Errorf("format = %v", resp.GetDocument().GetFormat())
	}
}

func TestExtractDocumentErrors(t *testing.T) {
	c := dial(t, nil)
	ctx := context.Background()

	if _, err := c.ExtractDocument(ctx, &pb.ExtractDocumentRequest{}); grpcCode(err) != codes.InvalidArgument {
		t.Errorf("empty request code = %v", grpcCode(err))
	}
	_, err := c.ExtractDocument(ctx, &pb.ExtractDocumentRequest{
		Document: &pb.DocumentInput{Content: testfixtures.MalformedPDF()},
	})
	if grpcCode(err) != codes.FailedPrecondition {
		t.Errorf("malformed pdf code = %v", grpcCode(err))
	}
}

func TestExportCSVFromParsedDocument(t *testing.T) {
	c := dial(t, nil)
	parsed, err := c.ParseDocument(context.Background(), &pb.ParseDocumentRequest{
		Document: &pb.DocumentInput{Content: []byte(sampleMD), Name: "sales"},
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.ExportCSV(context.Background(), &pb.ExportCSVRequest{
		Source: &pb.DataSource{Source: &pb.DataSource_Parsed{Parsed: parsed.GetDocument()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f := resp.GetFile()
	if f.GetFilename() != "sales.csv" || f.GetContentType() != "text/csv" || f.GetRowCount() != 2 {
		t.Errorf("file meta = %v %v %v", f.GetFilename(), f.GetContentType(), f.GetRowCount())
	}
	want := "name,amount,when\nJohn,1200.5,2024-03-01T00:00:00Z\nAlice,99,2024-04-02T00:00:00Z\n"
	if string(f.GetContent()) != want {
		t.Errorf("csv = %q", f.GetContent())
	}
}

func TestExportXLSXInlineSource(t *testing.T) {
	c := dial(t, nil)
	resp, err := c.ExportXLSX(context.Background(), &pb.ExportXLSXRequest{
		Source: &pb.DataSource{Source: &pb.DataSource_Document{
			Document: &pb.DocumentInput{Content: []byte("a,b\n1,2"), Name: "t"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f := resp.GetFile()
	if f.GetFilename() != "t.xlsx" || len(f.GetContent()) < 100 || string(f.GetContent()[:2]) != "PK" {
		t.Errorf("bad xlsx file: name=%q len=%d", f.GetFilename(), len(f.GetContent()))
	}
}

func TestErrorCodes(t *testing.T) {
	c := dial(t, nil)
	ctx := context.Background()

	cases := []struct {
		name string
		call func() error
		want codes.Code
	}{
		{"parse missing doc", func() error {
			_, err := c.ParseDocument(ctx, &pb.ParseDocumentRequest{})
			return err
		}, codes.InvalidArgument},
		{"parse empty content", func() error {
			_, err := c.ParseDocument(ctx, &pb.ParseDocumentRequest{Document: &pb.DocumentInput{}})
			return err
		}, codes.InvalidArgument},
		{"unparseable content", func() error {
			_, err := c.ParseDocument(ctx, &pb.ParseDocumentRequest{
				Document: &pb.DocumentInput{Format: pb.DocumentFormat_DOCUMENT_FORMAT_JSON, Content: []byte("nope")},
			})
			return err
		}, codes.FailedPrecondition},
		{"export no source", func() error {
			_, err := c.ExportCSV(ctx, &pb.ExportCSVRequest{})
			return err
		}, codes.InvalidArgument},
		{"powerbi unconfigured", func() error {
			_, err := c.PublishToPowerBI(ctx, &pb.PublishToPowerBIRequest{
				Source:      &pb.DataSource{Source: &pb.DataSource_Document{Document: &pb.DocumentInput{Content: []byte("a\n1")}}},
				DatasetName: "d",
			})
			return err
		}, codes.FailedPrecondition},
		{"powerbi missing dataset name", func() error {
			_, err := c.PublishToPowerBI(ctx, &pb.PublishToPowerBIRequest{
				Source: &pb.DataSource{Source: &pb.DataSource_Document{Document: &pb.DocumentInput{Content: []byte("a\n1")}}},
			})
			return err
		}, codes.FailedPrecondition},
	}
	for _, tc := range cases {
		if got := grpcCode(tc.call()); got != tc.want {
			t.Errorf("%s: code = %v, want %v", tc.name, got, tc.want)
		}
	}
}
