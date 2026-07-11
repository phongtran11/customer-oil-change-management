package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/lam-thinh/customer-oil-change-management/internal/config"
	database "github.com/lam-thinh/customer-oil-change-management/internal/db"
	dbsqlc "github.com/lam-thinh/customer-oil-change-management/internal/db/sqlc"
	"github.com/lam-thinh/customer-oil-change-management/internal/handler"
	"github.com/lam-thinh/customer-oil-change-management/internal/router"
	"github.com/lam-thinh/customer-oil-change-management/internal/service"
	_ "github.com/lam-thinh/customer-oil-change-management/internal/swagger"
)

// @title           Customer Oil Change Management API
// @version         1.0
// @description     A production-ready REST API for managing customer oil change records.
//
// @contact.name   API Support
// @contact.url    https://github.com/lam-thinh/customer-oil-change-management
//
// @license.name  MIT
//
// @BasePath  customer-oil-change/api
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter your JWT token in the format: **Bearer &lt;token&gt;**
func main() {
	if err := run(); err != nil {
		slog.Error("startup error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// ── 0. CLI Flags ──────────────────────────────────────────────────────────
	migrateOnly := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	// ── 1. Configuration ──────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── 2. Structured Logging ─────────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// If -migrate flag is provided, run migrations and exit.
	if *migrateOnly {
		logger.Info("running migrations manually...")
		if err := runMigrations(cfg.DBURL); err != nil {
			return fmt.Errorf("manual migrations failed: %w", err)
		}
		logger.Info("manual migrations applied successfully")
		return nil
	}

	// ── 3. Database Connection ─────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DBURL)
	if err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	defer pool.Close()
	logger.Info("database connected")

	// ── 4. Run Migrations ──────────────────────────────────────────────────────
	if cfg.AppEnv == "development" {
		if err := runMigrations(cfg.DBURL); err != nil {
			return fmt.Errorf("migrations: %w", err)
		}
		logger.Info("migrations applied")
	} else {
		logger.Info("migrations skipped (not in development environment)")
	}

	// ── 5. Dependency Graph ────────────────────────────────────────────────────
	queries := dbsqlc.New(pool)

	svcs := service.NewServices(queries, service.Config{
		JWTSecret:          cfg.JWTSecret,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
	}, logger)

	handlers := handler.NewHandlers(svcs, handler.Config{
		IsProd: cfg.AppEnv != "development",
	}, logger)

	// ── 6. Router ──────────────────────────────────────────────────────────────
	r := router.New(handlers, cfg.JWTSecret)

	// ── 7. HTTP Server with Graceful Shutdown ──────────────────────────────────
	addr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr)
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		logger.Info("server stopped gracefully")
	}

	return nil
}

// runMigrations uses the goose library with the pgx stdlib adapter to apply
// all pending SQL migrations from the ./migrations directory.
func runMigrations(dbURL string) error {
	stdDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer stdDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	// Migrations directory is relative to the working directory.
	// When running from the repo root: `./migrations`
	// When running from cmd/api: `../../migrations`
	// The Dockerfile sets WORKDIR /app and copies migrations to /app/migrations.
	if err := goose.Up(stdDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
