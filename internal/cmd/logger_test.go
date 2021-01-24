package cmd

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLogFormat(t *testing.T) {
	os.Unsetenv("GOMPLATE_LOG_FORMAT")
	defer os.Unsetenv("GOMPLATE_LOG_FORMAT")

	assert.Equal(t, "json", logFormat(nil))
	// os.Stdout isn't a terminal when this runs as a unit test...
	assert.Equal(t, "json", logFormat(os.Stdout))

	os.Setenv("GOMPLATE_LOG_FORMAT", "simple")
	assert.Equal(t, "simple", logFormat(os.Stdout))
	assert.Equal(t, "simple", logFormat(&bytes.Buffer{}))
}

func TestCreateLogger(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	buf := &bytes.Buffer{}

	// avoid timestamps
	zlog.Logger = zerolog.New(buf)

	l := createLogger("", buf)
	l.Debug().Msg("this won't show up")
	l.Warn().Msg("hello world")

	actual := strings.TrimSpace(buf.String())
	assert.Equal(t, `{"level":"warn","msg":"hello world"}`, actual)

	buf.Reset()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	l = createLogger("simple", buf)
	l.Debug().Str("field", "a value").Msg("this will show up")
	log.Println("hello world")

	actual = strings.TrimSpace(buf.String())
	assert.Equal(t, "this will show up field=\"a value\"\nhello world stdlog=true", actual)

	buf.Reset()
	l = createLogger("console", buf)
	l.Debug().Msg("hello")
	log.Println("world")

	actual = strings.TrimSpace(buf.String())
	assert.Equal(t, "<nil> DBG hello\n<nil> world stdlog=true", actual)

	buf.Reset()
	l = createLogger("logfmt", buf)
	l.Info().Str("field", "a value").Int("num", 84).Msg("hello\"")

	actual = strings.TrimSpace(buf.String())
	assert.Equal(t, "level=info msg=\"hello\\\"\" field=\"a value\" num=84", actual)
}
