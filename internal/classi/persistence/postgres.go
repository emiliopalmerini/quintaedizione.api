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

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// JSONB wrapper types for sqlx scanning

type proprietaLivelloSlice []classi.ProprietaLivello

func (p *proprietaLivelloSlice) Scan(src any) error {
	if src == nil {
		*p = nil
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion failed for proprietaLivelloSlice")
	}
	return json.Unmarshal(source, p)
}

func (p proprietaLivelloSlice) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

type equipaggiamentoPartenzaJSON classi.EquipaggiamentoPartenza

func (e *equipaggiamentoPartenzaJSON) Scan(src any) error {
	if src == nil {
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion failed for equipaggiamentoPartenzaJSON")
	}
	return json.Unmarshal(source, e)
}

func (e equipaggiamentoPartenzaJSON) Value() (driver.Value, error) {
	return json.Marshal(e)
}

type classeRow struct {
	ID                          string                             `db:"id"`
	Nome                        string                             `db:"nome"`
	Descrizione                 sql.NullString                     `db:"descrizione"`
	DocumentazioneDiRiferimento string                             `db:"documentazione_di_riferimento"`
	DadoVita                    string                             `db:"dado_vita"`
	EquipaggiamentoPartenza     equipaggiamentoPartenzaJSON `db:"equipaggiamento_partenza"`
	ProprietaDiClasse           proprietaLivelloSlice       `db:"proprieta_di_classe"`
}

func (r *classeRow) toClasse(sottoclassi []classi.RiferimentoSottoclasse) classi.Classe {
	c := classi.Classe{
		ID:                          r.ID,
		Nome:                        r.Nome,
		DocumentazioneDiRiferimento: r.DocumentazioneDiRiferimento,
		DadoVita:                    classi.TipoDiDado(r.DadoVita),
		ElencoSottoclassi:           sottoclassi,
		ProprietaDiClasse:           r.ProprietaDiClasse,
	}
	if r.Descrizione.Valid {
		c.Descrizione = r.Descrizione.String
	}
	if r.EquipaggiamentoPartenza.OpzioneA != nil || r.EquipaggiamentoPartenza.OpzioneB != nil {
		eq := classi.EquipaggiamentoPartenza(r.EquipaggiamentoPartenza)
		c.EquipaggiamentoPartenza = &eq
	}
	return c
}

type sottoclasseRow struct {
	ID                          string                       `db:"id"`
	Nome                        string                       `db:"nome"`
	Descrizione                 sql.NullString               `db:"descrizione"`
	DocumentazioneDiRiferimento string                       `db:"documentazione_di_riferimento"`
	IDClasseAssociata           string                       `db:"id_classe_associata"`
	ProprietaDiSottoclasse      proprietaLivelloSlice `db:"proprieta_di_sottoclasse"`
}

func (r *sottoclasseRow) toSottoClasse() classi.SottoClasse {
	s := classi.SottoClasse{
		ID:                          r.ID,
		Nome:                        r.Nome,
		DocumentazioneDiRiferimento: r.DocumentazioneDiRiferimento,
		IDClasseAssociata:           r.IDClasseAssociata,
		ProprietaDiSottoclasse:      r.ProprietaDiSottoclasse,
	}
	if r.Descrizione.Valid {
		s.Descrizione = r.Descrizione.String
	}
	return s
}

func (r *PostgresRepository) List(ctx context.Context, filter shared.ListFilter) ([]classi.Classe, int, error) {
	query := `
		SELECT id, nome, descrizione, documentazione_di_riferimento, dado_vita,
		       equipaggiamento_partenza, proprieta_di_classe
		FROM classi
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM classi WHERE 1=1`
	args := make(map[string]any)

	if filter.Nome != nil {
		query += ` AND nome ILIKE :nome`
		countQuery += ` AND nome ILIKE :nome`
		args["nome"] = "%" + shared.EscapeLike(*filter.Nome) + "%"
	}

	if len(filter.DocumentazioneDiRiferimento) > 0 {
		query += ` AND documentazione_di_riferimento = ANY(:docs)`
		countQuery += ` AND documentazione_di_riferimento = ANY(:docs)`
		args["docs"] = filter.DocumentazioneDiRiferimento
	}

	orderDir := "ASC"
	if filter.Sort == shared.SortDesc {
		orderDir = "DESC"
	}
	query += fmt.Sprintf(` ORDER BY nome %s`, orderDir)
	query += ` LIMIT :limit OFFSET :offset`
	args["limit"] = filter.Limit
	args["offset"] = filter.Offset

	// Get total count
	var total int
	countStmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare count query: %w", err)
	}
	defer countStmt.Close()
	if err := countStmt.GetContext(ctx, &total, args); err != nil {
		return nil, 0, fmt.Errorf("execute count query: %w", err)
	}

	// Get paginated results
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare query: %w", err)
	}
	defer stmt.Close()

	var rows []classeRow
	if err := stmt.SelectContext(ctx, &rows, args); err != nil {
		return nil, 0, fmt.Errorf("execute query: %w", err)
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	refMap, err := r.getSottoclassiRiferimentiByClasseIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	result := make([]classi.Classe, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.toClasse(refMap[row.ID]))
	}

	return result, total, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*classi.Classe, error) {
	query := `
		SELECT id, nome, descrizione, documentazione_di_riferimento, dado_vita,
		       equipaggiamento_partenza, proprieta_di_classe
		FROM classi
		WHERE id = $1
	`

	var row classeRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get classe by id: %w", err)
	}

	sottoclassi, err := r.getSottoclassiRiferimenti(ctx, id)
	if err != nil {
		return nil, err
	}

	classe := row.toClasse(sottoclassi)
	return &classe, nil
}

func (r *PostgresRepository) getSottoclassiRiferimenti(ctx context.Context, classeID string) ([]classi.RiferimentoSottoclasse, error) {
	query := `SELECT id FROM sottoclassi WHERE id_classe_associata = $1 ORDER BY nome`

	var ids []string
	if err := r.db.SelectContext(ctx, &ids, query, classeID); err != nil {
		return nil, fmt.Errorf("get sottoclassi riferimenti: %w", err)
	}

	result := make([]classi.RiferimentoSottoclasse, len(ids))
	for i, id := range ids {
		result[i] = classi.RiferimentoSottoclasse{IDSottoclasse: id}
	}
	return result, nil
}

type sottoclasseRef struct {
	ID                string `db:"id"`
	IDClasseAssociata string `db:"id_classe_associata"`
}

func (r *PostgresRepository) getSottoclassiRiferimentiByClasseIDs(ctx context.Context, classeIDs []string) (map[string][]classi.RiferimentoSottoclasse, error) {
	result := make(map[string][]classi.RiferimentoSottoclasse)
	if len(classeIDs) == 0 {
		return result, nil
	}

	query := `SELECT id, id_classe_associata FROM sottoclassi WHERE id_classe_associata = ANY($1) ORDER BY nome`

	var refs []sottoclasseRef
	if err := r.db.SelectContext(ctx, &refs, query, pq.Array(classeIDs)); err != nil {
		return nil, fmt.Errorf("batch get sottoclassi riferimenti: %w", err)
	}

	for _, ref := range refs {
		result[ref.IDClasseAssociata] = append(result[ref.IDClasseAssociata],
			classi.RiferimentoSottoclasse{IDSottoclasse: ref.ID})
	}
	return result, nil
}

func (r *PostgresRepository) ListSottoclassi(ctx context.Context, classeID string, filter shared.ListFilter) ([]classi.SottoClasse, int, error) {
	query := `
		SELECT id, nome, descrizione, documentazione_di_riferimento,
		       id_classe_associata, proprieta_di_sottoclasse
		FROM sottoclassi
		WHERE id_classe_associata = :classe_id
	`
	countQuery := `SELECT COUNT(*) FROM sottoclassi WHERE id_classe_associata = :classe_id`
	args := map[string]any{"classe_id": classeID}

	if filter.Nome != nil {
		query += ` AND nome ILIKE :nome`
		countQuery += ` AND nome ILIKE :nome`
		args["nome"] = "%" + shared.EscapeLike(*filter.Nome) + "%"
	}

	if len(filter.DocumentazioneDiRiferimento) > 0 {
		query += ` AND documentazione_di_riferimento = ANY(:docs)`
		countQuery += ` AND documentazione_di_riferimento = ANY(:docs)`
		args["docs"] = filter.DocumentazioneDiRiferimento
	}

	orderDir := "ASC"
	if filter.Sort == shared.SortDesc {
		orderDir = "DESC"
	}
	query += fmt.Sprintf(` ORDER BY nome %s`, orderDir)
	query += ` LIMIT :limit OFFSET :offset`
	args["limit"] = filter.Limit
	args["offset"] = filter.Offset

	// Get total count
	var total int
	countStmt, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare count query: %w", err)
	}
	defer countStmt.Close()
	if err := countStmt.GetContext(ctx, &total, args); err != nil {
		return nil, 0, fmt.Errorf("execute count query: %w", err)
	}

	// Get paginated results
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare query: %w", err)
	}
	defer stmt.Close()

	var rows []sottoclasseRow
	if err := stmt.SelectContext(ctx, &rows, args); err != nil {
		return nil, 0, fmt.Errorf("execute query: %w", err)
	}

	result := make([]classi.SottoClasse, len(rows))
	for i, row := range rows {
		result[i] = row.toSottoClasse()
	}

	return result, total, nil
}

func (r *PostgresRepository) GetSottoclasseByID(ctx context.Context, classeID, sottoclasseID string) (*classi.SottoClasse, error) {
	query := `
		SELECT id, nome, descrizione, documentazione_di_riferimento,
		       id_classe_associata, proprieta_di_sottoclasse
		FROM sottoclassi
		WHERE id = $1 AND id_classe_associata = $2
	`

	var row sottoclasseRow
	if err := r.db.GetContext(ctx, &row, query, sottoclasseID, classeID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get sottoclasse by id: %w", err)
	}

	sottoclasse := row.toSottoClasse()
	return &sottoclasse, nil
}
