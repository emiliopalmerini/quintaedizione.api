package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ModuleConfig struct {
	Name           string   `json:"name"`
	NamePlural     string   `json:"name_plural"`
	NameCapital    string   `json:"name_capital"`
	NamePluralCap  string   `json:"name_plural_cap"`
	TableName      string   `json:"table_name"`
	RouteParam     string   `json:"route_param"`
	HasSubresource bool     `json:"has_subresource"`
	SubName        string   `json:"sub_name,omitempty"`
	SubNamePlural  string   `json:"sub_name_plural,omitempty"`
	SubTableName   string   `json:"sub_table_name,omitempty"`
	DBFields       []DBField `json:"db_fields"`
	SubDBFields    []DBField `json:"sub_db_fields,omitempty"`
}

type DBField struct {
	Name       string `json:"name"`
	DBName     string `json:"db_name"`
	GoType     string `json:"go_type"`
	DBType     string `json:"db_type"`
	JSONName   string `json:"json_name"`
	Nullable   bool   `json:"nullable"`
	IsJSONB    bool   `json:"is_jsonb"`
	PrimaryKey bool   `json:"primary_key"`
}

func main() {
	configFile := flag.String("config", "", "Path to module config JSON file")
	outputDir := flag.String("output", "internal", "Output directory for generated code")
	migrationsDir := flag.String("migrations", "migrations", "Output directory for migrations")
	migrationNum := flag.Int("migration-num", 3, "Migration number to use")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("Usage: codegen -config <config.json> [-output <dir>] [-migrations <dir>] [-migration-num <num>]")
		os.Exit(1)
	}

	data, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	var config ModuleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		os.Exit(1)
	}

	if err := generateModule(config, *outputDir, *migrationsDir, *migrationNum); err != nil {
		fmt.Printf("Error generating module: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated module: %s\n", config.Name)
}

func generateModule(config ModuleConfig, outputDir, migrationsDir string, migrationNum int) error {
	moduleDir := filepath.Join(outputDir, config.Name)

	dirs := []string{
		moduleDir,
		filepath.Join(moduleDir, "persistence"),
		filepath.Join(moduleDir, "transports"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(moduleDir, "models.go"):                modelsTemplate,
		filepath.Join(moduleDir, "interfaces.go"):            interfacesTemplate,
		filepath.Join(moduleDir, "service.go"):               serviceTemplate,
		filepath.Join(moduleDir, "errors.go"):                errorsTemplate,
		filepath.Join(moduleDir, "responses.go"):             responsesTemplate,
		filepath.Join(moduleDir, "persistence", "postgres.go"): postgresTemplate,
		filepath.Join(moduleDir, "transports", "http.go"):    httpTemplate,
	}

	for path, tmplStr := range files {
		if err := writeTemplate(path, tmplStr, config); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	// Generate migration
	migrationUp := filepath.Join(migrationsDir, fmt.Sprintf("%06d_create_%s.up.sql", migrationNum, config.TableName))
	migrationDown := filepath.Join(migrationsDir, fmt.Sprintf("%06d_create_%s.down.sql", migrationNum, config.TableName))

	if err := writeTemplate(migrationUp, migrationUpTemplate, config); err != nil {
		return fmt.Errorf("write migration up: %w", err)
	}
	if err := writeTemplate(migrationDown, migrationDownTemplate, config); err != nil {
		return fmt.Errorf("write migration down: %w", err)
	}

	return nil
}

func writeTemplate(path, tmplStr string, data ModuleConfig) error {
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}

	tmpl, err := template.New("").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

var modelsTemplate = `package {{.Name}}

type {{.NameCapital}} struct {
{{- range .DBFields}}
	{{.Name}} {{.GoType}} ` + "`" + `json:"{{.JSONName}}"{{if .DBName}} db:"{{.DBName}}"{{end}}` + "`" + `
{{- end}}
}
{{if .HasSubresource}}
type {{.SubName}} struct {
{{- range .SubDBFields}}
	{{.Name}} {{.GoType}} ` + "`" + `json:"{{.JSONName}}"{{if .DBName}} db:"{{.DBName}}"{{end}}` + "`" + `
{{- end}}
}
{{end}}
`

var interfacesTemplate = `package {{.Name}}

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Repository interface {
	List(ctx context.Context, filter shared.ListFilter) ([]{{.NameCapital}}, int, error)
	GetByID(ctx context.Context, id string) (*{{.NameCapital}}, error)
}
`

var serviceTemplate = `package {{.Name}}

import (
	"context"
	"log/slog"

	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) List{{.NamePluralCap}}(ctx context.Context, filter shared.ListFilter) (*List{{.NamePluralCap}}Response, error) {
	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list {{.NamePlural}}", "error", err)
		return nil, shared.NewInternalError(err)
	}

	return &List{{.NamePluralCap}}Response{
		Pagina:           filter.Page(),
		NumeroDiElementi: total,
		{{.NamePluralCap}}:           items,
	}, nil
}

func (s *Service) Get{{.NameCapital}}(ctx context.Context, id string) (*{{.NameCapital}}, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get {{.Name}}", "id", id, "error", err)
		return nil, shared.NewInternalError(err)
	}
	if item == nil {
		return nil, Err{{.NameCapital}}NotFound(id)
	}
	return item, nil
}
`

var errorsTemplate = `package {{.Name}}

import (
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

func Err{{.NameCapital}}NotFound(id string) *shared.AppError {
	return shared.NewNotFoundError("{{.NameCapital}}", id)
}
`

var responsesTemplate = `package {{.Name}}

type List{{.NamePluralCap}}Response struct {
	Pagina           int           ` + "`" + `json:"pagina"` + "`" + `
	NumeroDiElementi int           ` + "`" + `json:"numero-di-elementi"` + "`" + `
	{{.NamePluralCap}}           []{{.NameCapital}} ` + "`" + `json:"{{.NamePlural}}"` + "`" + `
}
`

var postgresTemplate = `package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/emiliopalmerini/quintaedizione.api/internal/{{.Name}}"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context, filter shared.ListFilter) ([]{{.Name}}.{{.NameCapital}}, int, error) {
	query := ` + "`" + `SELECT * FROM {{.TableName}} WHERE 1=1` + "`" + `
	countQuery := ` + "`" + `SELECT COUNT(*) FROM {{.TableName}} WHERE 1=1` + "`" + `
	args := make(map[string]any)

	if filter.Nome != nil {
		query += ` + "`" + ` AND nome ILIKE :nome` + "`" + `
		countQuery += ` + "`" + ` AND nome ILIKE :nome` + "`" + `
		args["nome"] = "%" + *filter.Nome + "%"
	}

	if len(filter.DocumentazioneDiRiferimento) > 0 {
		query += ` + "`" + ` AND documentazione_di_riferimento = ANY(:docs)` + "`" + `
		countQuery += ` + "`" + ` AND documentazione_di_riferimento = ANY(:docs)` + "`" + `
		args["docs"] = filter.DocumentazioneDiRiferimento
	}

	orderDir := "ASC"
	if filter.Sort == shared.SortDesc {
		orderDir = "DESC"
	}
	query += fmt.Sprintf(` + "`" + ` ORDER BY nome %s` + "`" + `, orderDir)
	query += ` + "`" + ` LIMIT :limit OFFSET :offset` + "`" + `
	args["limit"] = filter.Limit
	args["offset"] = filter.Offset

	var total int
	countStmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare count query: %w", err)
	}
	defer countStmt.Close()
	if err := countStmt.GetContext(ctx, &total, args); err != nil {
		return nil, 0, fmt.Errorf("execute count query: %w", err)
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare query: %w", err)
	}
	defer stmt.Close()

	var items []{{.Name}}.{{.NameCapital}}
	if err := stmt.SelectContext(ctx, &items, args); err != nil {
		return nil, 0, fmt.Errorf("execute query: %w", err)
	}

	return items, total, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*{{.Name}}.{{.NameCapital}}, error) {
	query := ` + "`" + `SELECT * FROM {{.TableName}} WHERE id = $1` + "`" + `

	var item {{.Name}}.{{.NameCapital}}
	if err := r.db.GetContext(ctx, &item, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get {{.Name}} by id: %w", err)
	}

	return &item, nil
}
`

var httpTemplate = `package transports

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/quintaedizione.api/internal/{{.Name}}"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type Handler struct {
	service *{{.Name}}.Service
}

func NewHandler(service *{{.Name}}.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List{{.NamePluralCap}})
	r.Get("/{{"{"}}{{.RouteParam}}{{"}"}}", h.Get{{.NameCapital}})

	return r
}

func (h *Handler) List{{.NamePluralCap}}(w http.ResponseWriter, r *http.Request) {
	filter, err := shared.NewListFilterFromRequest(r)
	if err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	response, err := h.service.List{{.NamePluralCap}}(r.Context(), filter)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Get{{.NameCapital}}(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "{{.RouteParam}}")
	if err := shared.ValidateID("{{.RouteParam}}", id); err != nil {
		shared.WriteError(w, shared.NewBadRequestError(err.Error(), err))
		return
	}

	item, err := h.service.Get{{.NameCapital}}(r.Context(), id)
	if err != nil {
		shared.WriteError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, item)
}
`

var migrationUpTemplate = `CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id                            VARCHAR(255) PRIMARY KEY,
    nome                          VARCHAR(255) NOT NULL,
    descrizione                   TEXT,
    documentazione_di_riferimento VARCHAR(50) DEFAULT 'DND 2024',
    created_at                    TIMESTAMP DEFAULT NOW(),
    updated_at                    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_{{.TableName}}_nome ON {{.TableName}}(nome);
CREATE INDEX IF NOT EXISTS idx_{{.TableName}}_documentazione ON {{.TableName}}(documentazione_di_riferimento);
`

var migrationDownTemplate = `DROP INDEX IF EXISTS idx_{{.TableName}}_documentazione;
DROP INDEX IF EXISTS idx_{{.TableName}}_nome;
DROP TABLE IF EXISTS {{.TableName}};
`
