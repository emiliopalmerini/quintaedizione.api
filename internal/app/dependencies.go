package app

import (
	"log/slog"

	"github.com/jmoiron/sqlx"

	"github.com/emiliopalmerini/quintaedizione.api/internal/config"
)

type Dependencies struct {
	DB     *sqlx.DB
	Logger *slog.Logger
	Config *config.Config
}
