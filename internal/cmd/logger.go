package cmd

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"time"

	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/term"
)

func logFormat(out io.Writer) string {
	defaultFormat := "json"
	if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		defaultFormat = "console"
	}
	return env.Getenv("GOMPLATE_LOG_FORMAT", defaultFormat)
}

func fmtField(fname string) func(i interface{}) string {
	return func(i interface{}) string {
		if i == nil || i == "" {
			return ""
		}

		if s, ok := i.(string); ok {
			for _, c := range s {
				if c <= 0x20 || c == '\\' || c == '"' {
					return fmt.Sprintf("%s=%q", fname, i)
				}
			}
		}
		return fmt.Sprintf("%s=%s", fname, i)
	}
}

func createLogger(format string, out io.Writer) zerolog.Logger {
	zerolog.MessageFieldName = "msg"

	l := zlog.Logger.Output(out)

	stdlogger := l.With().Bool("stdlog", true).Logger()
	stdlog.SetFlags(0)

	switch format {
	case "console":
		useColour := false
		if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) && runtime.GOOS != "windows" {
			useColour = true
		}
		l = l.Output(zerolog.ConsoleWriter{
			Out:        out,
			NoColor:    !useColour,
			TimeFormat: "15:04:05",
		})
		stdlogger = stdlogger.Output(zerolog.ConsoleWriter{
			Out:         out,
			NoColor:     !useColour,
			FormatLevel: func(i interface{}) string { return "" },
		})
	case "logfmt":
		w := zerolog.ConsoleWriter{
			Out:             out,
			NoColor:         true,
			FormatMessage:   fmtField(zerolog.MessageFieldName),
			FormatLevel:     fmtField(zerolog.LevelFieldName),
			FormatTimestamp: fmtField(zerolog.TimestampFieldName),
		}
		l = l.Output(w)
		stdlogger = stdlogger.Output(w)
	case "simple":
		w := zerolog.ConsoleWriter{
			Out:             out,
			NoColor:         true,
			FormatLevel:     func(i interface{}) string { return "" },
			FormatTimestamp: func(i interface{}) string { return "" },
		}
		l = l.Output(w)
		stdlogger = stdlogger.Output(w)
	}
	stdlog.SetOutput(stdlogger)

	return l
}

func initLogger(ctx context.Context, out io.Writer) context.Context {
	// default to warn level
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	zerolog.DurationFieldUnit = time.Second

	l := createLogger(logFormat(out), out)

	return l.WithContext(ctx)
}
