package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/emiliopalmerini/quintaedizione.api/internal/classi"
	"github.com/emiliopalmerini/quintaedizione.api/internal/classi/persistence"
	"github.com/emiliopalmerini/quintaedizione.api/internal/classi/transports"
	"github.com/emiliopalmerini/quintaedizione.api/internal/config"
	"github.com/emiliopalmerini/quintaedizione.api/internal/health"
)

func main() {
	godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := runMigrations(db, logger); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	repo := persistence.NewPostgresRepository(db)
	service := classi.NewService(repo, logger)
	handler := transports.NewHandler(service)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	healthHandler := health.NewHandler(db.DB, cfg.Version)
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/health/live", healthHandler.Liveness)
	r.Mount("/classi", handler.Routes())

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger/quintaedizioneswagger")
	})

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server starting", "port", cfg.Server.Port)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}

	return nil
}

func runMigrations(db *sqlx.DB, logger *slog.Logger) error {
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

	logger.Info("migrations applied successfully")
	return nil
}
