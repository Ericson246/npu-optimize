package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level  int
	Format string
	Writer io.Writer
}

func Init(cfg Config) {
	level := slog.LevelWarn
	switch cfg.Level {
	case 1:
		level = slog.LevelInfo
	case 2, 3:
		level = slog.LevelDebug
	}

	w := cfg.Writer
	if w == nil {
		w = os.Stderr
	}

	opts := &slog.HandlerOptions{Level: level}
	if cfg.Level >= 2 {
		opts.AddSource = true
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	default:
		handler = slog.NewTextHandler(w, opts)
	}

	slog.SetDefault(slog.New(handler))
}
