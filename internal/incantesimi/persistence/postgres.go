package persistence

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// JSONB scanning helpers

func scanJSON(src any, dest any) error {
	if src == nil {
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return errors.New("scanJSON: source is not []byte")
	}
	return json.Unmarshal(source, dest)
}

type effettoIncantesimoJSON incantesimi.EffettoIncantesimo

func (e *effettoIncantesimoJSON) Scan(src any) error { return scanJSON(src, e) }
func (e effettoIncantesimoJSON) Value() (driver.Value, error) {
	return json.Marshal(e)
}

type componentiSlice []incantesimi.Componente

func (c *componentiSlice) Scan(src any) error {
	if src == nil {
		*c = nil
		return nil
	}
	var arr []string
	if err := pq.Array(&arr).Scan(src); err != nil {
		return err
	}
	result := make([]incantesimi.Componente, len(arr))
	for i, v := range arr {
		result[i] = incantesimi.Componente(v)
	}
	*c = result
	return nil
}

func (c componentiSlice) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	arr := make([]string, len(c))
	for i, v := range c {
		arr[i] = string(v)
	}
	return pq.Array(arr).Value()
}

type incantesimoRow struct {
	ID                          string                  `db:"id"`
	Nome                        string                  `db:"nome"`
	Livello                     int32                   `db:"livello"`
	ScuolaDiMagia               string                  `db:"scuola_di_magia"`
	TempoDiLancio               string                  `db:"tempo_di_lancio"`
	Gittata                     string                  `db:"gittata"`
	Area                        sql.NullString          `db:"area"`
	Concentrazione              bool                    `db:"concentrazione"`
	SemprePreparato             bool                    `db:"sempre_preparato"`
	Rituale                     bool                    `db:"rituale"`
	Componenti                  componentiSlice         `db:"componenti"`
	ComponentiMateriali         sql.NullString          `db:"componenti_materiali"`
	Durata                      string                  `db:"durata"`
	Descrizione                 string                  `db:"descrizione"`
	EffettoIncantesimo          *effettoIncantesimoJSON `db:"effetto_incantesimo"`
	EffettoLivelloMaggiore      *effettoIncantesimoJSON `db:"effetto_livello_maggiore"`
	Classi                      string                  `db:"classi"`
	DocumentazioneDiRiferimento string                  `db:"documentazione_di_riferimento"`
}

func (r *incantesimoRow) toIncantesimo() incantesimi.Incantesimo {
	inc := incantesimi.Incantesimo{
		ID:                          r.ID,
		Nome:                        r.Nome,
		Livello:                     r.Livello,
		ScuolaDiMagia:               incantesimi.ScuolaDiMagia(r.ScuolaDiMagia),
		TempoDiLancio:               r.TempoDiLancio,
		Gittata:                     r.Gittata,
		Concentrazione:              r.Concentrazione,
		SemprePreparato:             r.SemprePreparato,
		Rituale:                     r.Rituale,
		Componenti:                  r.Componenti,
		Durata:                      r.Durata,
		Descrizione:                 r.Descrizione,
		Classi:                      r.Classi,
		DocumentazioneDiRiferimento: r.DocumentazioneDiRiferimento,
	}
	if r.Area.Valid {
		inc.Area = r.Area.String
	}
	if r.ComponentiMateriali.Valid {
		inc.ComponentiMateriali = r.ComponentiMateriali.String
	}
	if r.EffettoIncantesimo != nil {
		e := incantesimi.EffettoIncantesimo(*r.EffettoIncantesimo)
		inc.EffettoIncantesimo = &e
	}
	if r.EffettoLivelloMaggiore != nil {
		e := incantesimi.EffettoIncantesimo(*r.EffettoLivelloMaggiore)
		inc.EffettoLivelloMaggiore = &e
	}
	return inc
}

const incantesimiColumns = `id, nome, livello, scuola_di_magia, tempo_di_lancio, gittata, area,
	concentrazione, sempre_preparato, rituale, componenti, componenti_materiali,
	durata, descrizione, effetto_incantesimo, effetto_livello_maggiore,
	classi, documentazione_di_riferimento`

type paginatedQuery struct {
	query      string
	countQuery string
	args       map[string]any
}

func newIncantesimiPaginatedQuery(filter incantesimi.IncantesimiFilter) *paginatedQuery {
	baseQuery := `SELECT ` + incantesimiColumns + ` FROM incantesimi WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM incantesimi WHERE 1=1`
	args := make(map[string]any)

	// Base ListFilter fields
	if filter.Nome != nil {
		baseQuery += ` AND nome ILIKE :nome`
		countQuery += ` AND nome ILIKE :nome`
		args["nome"] = "%" + shared.EscapeLike(*filter.Nome) + "%"
	}
	if len(filter.DocumentazioneDiRiferimento) > 0 {
		baseQuery += ` AND documentazione_di_riferimento = ANY(:docs)`
		countQuery += ` AND documentazione_di_riferimento = ANY(:docs)`
		args["docs"] = pq.Array(filter.DocumentazioneDiRiferimento)
	}

	// Domain-specific filters
	if filter.Livello != nil {
		baseQuery += ` AND livello = :livello`
		countQuery += ` AND livello = :livello`
		args["livello"] = *filter.Livello
	}
	if filter.ScuolaDiMagia != nil {
		baseQuery += ` AND scuola_di_magia = :scuola_di_magia`
		countQuery += ` AND scuola_di_magia = :scuola_di_magia`
		args["scuola_di_magia"] = *filter.ScuolaDiMagia
	}
	if filter.Concentrazione != nil {
		baseQuery += ` AND concentrazione = :concentrazione`
		countQuery += ` AND concentrazione = :concentrazione`
		args["concentrazione"] = *filter.Concentrazione
	}
	if filter.Rituale != nil {
		baseQuery += ` AND rituale = :rituale`
		countQuery += ` AND rituale = :rituale`
		args["rituale"] = *filter.Rituale
	}
	if len(filter.Componenti) > 0 {
		baseQuery += ` AND componenti @> :componenti_filter`
		countQuery += ` AND componenti @> :componenti_filter`
		args["componenti_filter"] = pq.Array(filter.Componenti)
	}
	if filter.ComponentiMateriali != nil {
		baseQuery += ` AND componenti_materiali ILIKE :componenti_materiali`
		countQuery += ` AND componenti_materiali ILIKE :componenti_materiali`
		args["componenti_materiali"] = "%" + shared.EscapeLike(*filter.ComponentiMateriali) + "%"
	}
	if filter.TempoDiLancio != nil {
		baseQuery += ` AND tempo_di_lancio ILIKE :tempo_di_lancio`
		countQuery += ` AND tempo_di_lancio ILIKE :tempo_di_lancio`
		args["tempo_di_lancio"] = "%" + shared.EscapeLike(*filter.TempoDiLancio) + "%"
	}
	if filter.Gittata != nil {
		baseQuery += ` AND gittata ILIKE :gittata`
		countQuery += ` AND gittata ILIKE :gittata`
		args["gittata"] = "%" + shared.EscapeLike(*filter.Gittata) + "%"
	}
	if filter.Durata != nil {
		baseQuery += ` AND durata ILIKE :durata`
		countQuery += ` AND durata ILIKE :durata`
		args["durata"] = "%" + shared.EscapeLike(*filter.Durata) + "%"
	}
	for i, classe := range filter.Classi {
		key := fmt.Sprintf("classe_%d", i)
		baseQuery += fmt.Sprintf(` AND classi ILIKE :%s`, key)
		countQuery += fmt.Sprintf(` AND classi ILIKE :%s`, key)
		args[key] = "%" + shared.EscapeLike(classe) + "%"
	}

	// Sort and pagination
	orderDir := "ASC"
	if filter.Sort == shared.SortDesc {
		orderDir = "DESC"
	}
	baseQuery += fmt.Sprintf(` ORDER BY nome %s`, orderDir)
	baseQuery += ` LIMIT :limit OFFSET :offset`
	args["limit"] = filter.Limit
	args["offset"] = filter.Offset

	return &paginatedQuery{query: baseQuery, countQuery: countQuery, args: args}
}

func (q *paginatedQuery) count(ctx context.Context, db *sqlx.DB) (int, error) {
	var total int
	stmt, err := db.PrepareNamedContext(ctx, q.countQuery)
	if err != nil {
		return 0, fmt.Errorf("prepare count query: %w", err)
	}
	defer stmt.Close()
	if err := stmt.GetContext(ctx, &total, q.args); err != nil {
		return 0, fmt.Errorf("execute count query: %w", err)
	}
	return total, nil
}

func (q *paginatedQuery) selectRows(ctx context.Context, db *sqlx.DB, dest any) error {
	stmt, err := db.PrepareNamedContext(ctx, q.query)
	if err != nil {
		return fmt.Errorf("prepare query: %w", err)
	}
	defer stmt.Close()
	if err := stmt.SelectContext(ctx, dest, q.args); err != nil {
		return fmt.Errorf("execute query: %w", err)
	}
	return nil
}

func (r *PostgresRepository) List(ctx context.Context, filter incantesimi.IncantesimiFilter) ([]incantesimi.Incantesimo, int, error) {
	q := newIncantesimiPaginatedQuery(filter)

	total, err := q.count(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	var rows []incantesimoRow
	if err := q.selectRows(ctx, r.db, &rows); err != nil {
		return nil, 0, err
	}

	result := make([]incantesimi.Incantesimo, len(rows))
	for i, row := range rows {
		result[i] = row.toIncantesimo()
	}

	return result, total, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*incantesimi.Incantesimo, error) {
	query := `SELECT ` + incantesimiColumns + ` FROM incantesimi WHERE id = $1`

	var row incantesimoRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get incantesimo by id: %w", err)
	}

	inc := row.toIncantesimo()
	return &inc, nil
}
