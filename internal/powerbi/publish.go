package powerbi

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/bushadigitallimited/bloom-parser/internal/parser"
)

var tracer = otel.Tracer("github.com/bushadigitallimited/bloom-parser/internal/powerbi")

// rowsPerBatch stays under the Power BI push-API per-request row limit.
const rowsPerBatch = 5000

// Publisher owns the explicit flow: schema -> find-or-create dataset ->
// clear rows -> push rows in batches.
type Publisher struct {
	client      Client
	workspaceID string
}

func NewPublisher(client Client, defaultWorkspace string) *Publisher {
	return &Publisher{client: client, workspaceID: defaultWorkspace}
}

type PublishRequest struct {
	Doc         *parser.ParsedDocument
	WorkspaceID string // overrides the configured default when set
	DatasetName string
	TableName   string // defaults to "Table1"
}

type PublishResult struct {
	DatasetID   string
	WorkspaceID string
	TableName   string
	RowsPushed  int64
}

func (p *Publisher) Publish(ctx context.Context, req PublishRequest) (*PublishResult, error) {
	if p.client == nil {
		return nil, ErrNotConfigured
	}
	if req.DatasetName == "" {
		return nil, fmt.Errorf("%w: dataset_name is required", parser.ErrInvalidDocument)
	}
	ws := req.WorkspaceID
	if ws == "" {
		ws = p.workspaceID
	}
	table := req.TableName
	if table == "" {
		table = "Table1"
	}

	ds, err := p.client.GetDatasetByName(ctx, ws, req.DatasetName)
	if err != nil {
		return nil, fmt.Errorf("lookup dataset: %w", err)
	}
	if ds == nil {
		ds, err = p.createDataset(ctx, ws, req.DatasetName, table, req.Doc.Columns)
		if err != nil {
			return nil, err
		}
	}

	if err := p.client.DeleteRows(ctx, ws, ds.ID, table); err != nil {
		return nil, fmt.Errorf("clear table %q: %w", table, err)
	}

	pushed := int64(0)
	for batch := range batchRows(toRowMaps(req.Doc), rowsPerBatch) {
		if err := p.pushRows(ctx, ws, ds.ID, table, batch); err != nil {
			return nil, err
		}
		pushed += int64(len(batch))
	}
	return &PublishResult{
		DatasetID:   ds.ID,
		WorkspaceID: ws,
		TableName:   table,
		RowsPushed:  pushed,
	}, nil
}

// createDataset creates the push dataset with its table in one call — the
// Power BI API only accepts table definitions at dataset creation time.
func (p *Publisher) createDataset(ctx context.Context, ws, name, table string, cols []parser.Column) (*Dataset, error) {
	ctx, span := tracer.Start(ctx, "powerbi.create_dataset")
	defer span.End()
	ds, err := p.client.CreateDataset(ctx, ws, DatasetDef{
		Name:        name,
		DefaultMode: "Push",
		Tables:      []TableDef{tableDef(table, cols)},
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("create dataset: %w", err)
	}
	return ds, nil
}

func (p *Publisher) pushRows(ctx context.Context, ws, datasetID, table string, batch []map[string]any) error {
	ctx, span := tracer.Start(ctx, "powerbi.push_rows")
	defer span.End()
	if err := p.client.PostRows(ctx, ws, datasetID, table, batch); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("push rows: %w", err)
	}
	return nil
}

func toRowMaps(doc *parser.ParsedDocument) []map[string]any {
	rows := make([]map[string]any, len(doc.Rows))
	for i, r := range doc.Rows {
		m := make(map[string]any, len(doc.Columns))
		for c, col := range doc.Columns {
			if c < len(r) {
				m[col.Name] = scalar(r[c])
			}
		}
		rows[i] = m
	}
	return rows
}

func scalar(v parser.Value) any {
	switch v.Kind {
	case parser.KindString:
		return v.Str
	case parser.KindInt:
		return v.Int
	case parser.KindFloat:
		return v.Float
	case parser.KindBool:
		return v.Bool
	case parser.KindTime:
		return v.Time.Format(time.RFC3339)
	default:
		return nil
	}
}

func batchRows(rows []map[string]any, size int) func(yield func([]map[string]any) bool) {
	return func(yield func([]map[string]any) bool) {
		for start := 0; start < len(rows); start += size {
			end := min(start+size, len(rows))
			if !yield(rows[start:end]) {
				return
			}
		}
	}
}
