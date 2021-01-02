package cmd

import (
	"context"
	"io"
	stdlog "log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

func initLogger(ctx context.Context, out io.Writer) context.Context {
	// default to warn level
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	zerolog.DurationFieldUnit = time.Second

	stdlogger := log.With().Bool("stdlog", true).Logger()
	stdlog.SetFlags(0)
	stdlog.SetOutput(stdlogger)

	if f, ok := out.(*os.File); ok {
		if term.IsTerminal(int(f.Fd())) {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
			noLevelWriter := zerolog.ConsoleWriter{
				Out:         f,
				FormatLevel: func(i interface{}) string { return "" },
			}
			stdlogger = stdlogger.Output(noLevelWriter)
			stdlog.SetOutput(stdlogger)
		}
	}

	return log.Logger.WithContext(ctx)
}
