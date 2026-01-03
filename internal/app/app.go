package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	classipersistence "github.com/emiliopalmerini/quintaedizione.api/internal/classi/persistence"
	classitransports "github.com/emiliopalmerini/quintaedizione.api/internal/classi/transports"
	"github.com/emiliopalmerini/quintaedizione.api/internal/config"
	"github.com/emiliopalmerini/quintaedizione.api/internal/health"
	"github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi"
	incantesimipersistence "github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi/persistence"
	incantesimitransports "github.com/emiliopalmerini/quintaedizione.api/internal/incantesimi/transports"
	custommw "github.com/emiliopalmerini/quintaedizione.api/internal/middleware"
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

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(custommw.Logger(a.deps.Logger))
	r.Use(middleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   a.deps.Config.CORS.AllowedOrigins,
		AllowedMethods:   a.deps.Config.CORS.AllowedMethods,
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           a.deps.Config.CORS.MaxAge,
	}))

	// Rate limiting
	if a.deps.Config.RateLimit.Enabled {
		r.Use(httprate.LimitByIP(a.deps.Config.RateLimit.RequestsPerMinute, time.Minute))
	}

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Public endpoints
	healthHandler := health.NewHandler(a.deps.DB.DB, a.deps.Config.Version)
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger/quintaedizioneswagger")
	})

	// Protected API routes
	r.Route("/v1", func(r chi.Router) {
		r.Use(custommw.APIKey(a.deps.Config.APIKey))

		// Classi
		classiRepo := classipersistence.NewPostgresRepository(a.deps.DB)
		classiService := classi.NewService(classiRepo, a.deps.Logger)
		classiHandler := classitransports.NewHandler(classiService)
		r.Mount("/classi", classiHandler.Routes())

		// Incantesimi
		incantesimiRepo := incantesimipersistence.NewPostgresRepository(a.deps.DB)
		incantesimiService := incantesimi.NewService(incantesimiRepo, a.deps.Logger)
		incantesimiHandler := incantesimitransports.NewHandler(incantesimiService)
		r.Mount("/incantesimi", incantesimiHandler.Routes())
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
