package cmd

import (
	"io"
	"log/slog"
	"os"
	"runtime"

	"github.com/hairyhenderson/gomplate/v4/env"
	"github.com/lmittmann/tint"
	"golang.org/x/term"
)

func logFormat(out io.Writer) string {
	defaultFormat := "json"
	if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		defaultFormat = "console"
	}
	return env.Getenv("GOMPLATE_LOG_FORMAT", defaultFormat)
}

func createLogHandler(format string, out io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch format {
	case "console":
		// logFormat() already checks if this is a terminal, but we need to
		// check again because the format may be overridden with `GOMPLATE_LOG_FORMAT`
		useColour := false
		if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) && runtime.GOOS != "windows" {
			useColour = true
		}
		handler = tint.NewHandler(out, &tint.Options{
			Level:      level,
			TimeFormat: "15:04:05",
			NoColor:    !useColour,
		})
	case "simple":
		handler = tint.NewHandler(out, &tint.Options{
			Level:   level,
			NoColor: true,
			ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
				if attr.Key == "level" {
					attr.Value = slog.StringValue("")
				}
				if attr.Key == "time" {
					attr.Value = slog.StringValue("")
				}
				return attr
			},
		})
	case "logfmt":
		handler = slog.NewTextHandler(out, opts)
	default:
		// json is still default
		handler = slog.NewJSONHandler(out, opts)
	}

	return handler
}

func initLogger(out io.Writer, level slog.Level) {
	// default to warn level
	if level == 0 {
		level = slog.LevelWarn
	}

	handler := createLogHandler(logFormat(out), out, level)
	slog.SetDefault(slog.New(handler))
}
