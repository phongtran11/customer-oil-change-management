package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// ContextHandler wraps a slog.Handler to automatically inject context values
// (like go-chi's RequestID) into the log record attributes.
type ContextHandler struct {
	slog.Handler
}

// Handle implements slog.Handler. It extracts the request ID from the context
// and adds it to the log record if present.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if reqID := middleware.GetReqID(ctx); reqID != "" {
			r.AddAttrs(slog.String("request_id", reqID))
		}
	}
	return h.Handler.Handle(ctx, r)
}

// parseLevel maps a string representation of a log level to slog.Level.
// Defaults to slog.LevelInfo if the level is unknown or empty.
func parseLevel(lvl string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Init configures and sets the global default slog logger based on the environment and level.
func Init(env string, level string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Wrap the handler with ContextHandler to support request ID injection
	logger := slog.New(&ContextHandler{Handler: handler})
	slog.SetDefault(logger)
}

// Middleware returns a custom request logging middleware using slog.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			slog.InfoContext(r.Context(), "HTTP request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_ip", r.RemoteAddr),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
