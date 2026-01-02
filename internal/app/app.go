package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/classi/persistence"
	"github.com/emiliopalmerini/quintaedizione.api/internal/classi/transports"
	"github.com/emiliopalmerini/quintaedizione.api/internal/config"
	"github.com/emiliopalmerini/quintaedizione.api/internal/health"
)

type App struct {
	deps   *Dependencies
	router chi.Router
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	deps := &Dependencies{
		DB:     db,
		Logger: logger,
		Config: cfg,
	}

	app := &App{deps: deps}
	app.setupRoutes()

	logger.Info("migrations applied successfully")

	return app, nil
}

func (a *App) setupRoutes() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	healthHandler := health.NewHandler(a.deps.DB.DB, a.deps.Config.Version)
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger/quintaedizioneswagger")
	})

	r.Route("/v1", func(r chi.Router) {
		classiRepo := persistence.NewPostgresRepository(a.deps.DB)
		classiService := classi.NewService(classiRepo, a.deps.Logger)
		classiHandler := transports.NewHandler(classiService)
		r.Mount("/classi", classiHandler.Routes())
	})

	a.router = r
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) Close() error {
	return a.deps.DB.Close()
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
