package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

type SlogAdapter struct{}

func (s *SlogAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	var slogLvl slog.Level

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		slogLvl = slog.LevelDebug
	case tracelog.LogLevelInfo:
		slogLvl = slog.LevelInfo
	case tracelog.LogLevelWarn:
		slogLvl = slog.LevelWarn
	case tracelog.LogLevelError:
		slogLvl = slog.LevelError
	default:
		slogLvl = slog.LevelInfo
	}

	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}
	slog.LogAttrs(ctx, slogLvl, msg, attrs...)
}

// NewPool creates and validates a new pgxpool connection pool.
func NewPool(ctx context.Context, dbURL string, logLevel string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("database: parse config: %w", err)
	}

	var level tracelog.LogLevel
	switch logLevel {
	case "debug":
		level = tracelog.LogLevelDebug
	case "info":
		level = tracelog.LogLevelInfo
	case "warn":
		level = tracelog.LogLevelWarn
	case "error":
		level = tracelog.LogLevelError
	default:
		level = tracelog.LogLevelInfo
	}

	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &SlogAdapter{},
		LogLevel: level,
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return pool, nil
}
