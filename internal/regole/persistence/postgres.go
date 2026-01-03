package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/emiliopalmerini/quintaedizione.api/internal/regole"
	"github.com/emiliopalmerini/quintaedizione.api/internal/shared"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context, filter shared.ListFilter) ([]regole.Regola, int, error) {
	query := `SELECT * FROM regole WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM regole WHERE 1=1`
	args := make(map[string]any)

	if filter.Nome != nil {
		query += ` AND nome ILIKE :nome`
		countQuery += ` AND nome ILIKE :nome`
		args["nome"] = "%" + *filter.Nome + "%"
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

	var items []regole.Regola
	if err := stmt.SelectContext(ctx, &items, args); err != nil {
		return nil, 0, fmt.Errorf("execute query: %w", err)
	}

	return items, total, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*regole.Regola, error) {
	query := `SELECT * FROM regole WHERE id = $1`

	var item regole.Regola
	if err := r.db.GetContext(ctx, &item, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get regole by id: %w", err)
	}

	return &item, nil
}
