package powerbi

import (
	"errors"

	"github.com/bushadigitallimited/bloom-parser/internal/parser"
)

var (
	ErrNotConfigured    = errors.New("power bi credentials are not configured")
	ErrUnauthenticated  = errors.New("power bi authentication failed")
	ErrPermissionDenied = errors.New("power bi access denied")
	ErrUnavailable      = errors.New("power bi api unavailable")
	ErrAPI              = errors.New("power bi api error")
)

// Config carries non-secret connection settings; secrets come from env.
type Config struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	// Default workspace (group) id; empty targets "My workspace".
	WorkspaceID string
	// BaseURL/AuthURL default to the public Power BI / Entra endpoints and are
	// overridable for tests.
	BaseURL string
	AuthURL string
}

func (c Config) Configured() bool {
	return c.TenantID != "" && c.ClientID != "" && c.ClientSecret != ""
}

type Dataset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ColumnDef struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

type TableDef struct {
	Name    string      `json:"name"`
	Columns []ColumnDef `json:"columns"`
}

type DatasetDef struct {
	Name        string     `json:"name"`
	DefaultMode string     `json:"defaultMode"`
	Tables      []TableDef `json:"tables"`
}

var dataType = map[parser.ColumnType]string{
	parser.TypeString: "String",
	parser.TypeInt:    "Int64",
	parser.TypeFloat:  "Double",
	parser.TypeBool:   "Boolean",
	parser.TypeTime:   "DateTime",
}

// tableDef maps normalized columns to a Power BI table schema.
func tableDef(name string, cols []parser.Column) TableDef {
	t := TableDef{Name: name, Columns: make([]ColumnDef, len(cols))}
	for i, c := range cols {
		dt, ok := dataType[c.Type]
		if !ok {
			dt = "String"
		}
		t.Columns[i] = ColumnDef{Name: c.Name, DataType: dt}
	}
	return t
}
