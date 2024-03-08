package cmd

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogFormat(t *testing.T) {
	os.Unsetenv("GOMPLATE_LOG_FORMAT")

	assert.Equal(t, "json", logFormat(nil))
	// os.Stdout isn't a terminal when this runs as a unit test...
	assert.Equal(t, "json", logFormat(os.Stdout))

	t.Setenv("GOMPLATE_LOG_FORMAT", "simple")
	assert.Equal(t, "simple", logFormat(os.Stdout))
	assert.Equal(t, "simple", logFormat(&bytes.Buffer{}))
}

// a slog handler that strips the 'time' field
type noTimestampHandler struct {
	slog.Handler
}

var _ slog.Handler = (*noTimestampHandler)(nil)

func (h *noTimestampHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Time = time.Time{}

	return h.Handler.Handle(ctx, r)
}

func TestCreateLogHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	h := createLogHandler("", buf, slog.LevelWarn)
	// strip the 'time' field for easier comparison
	h = &noTimestampHandler{h}
	l := slog.New(h)
	slog.SetDefault(l)

	l.Debug("this won't show up")
	l.Warn("hello world")

	actual := strings.TrimSpace(buf.String())
	assert.JSONEq(t, `{"level":"WARN","msg":"hello world"}`, actual)

	buf.Reset()
	h = createLogHandler("simple", buf, slog.LevelDebug)
	l = slog.New(h)
	slog.SetDefault(l)

	l.Debug("this will show up", "field", "a value")
	log.Println("hello world")

	actual = strings.TrimSpace(buf.String())
	assert.Equal(t, "this will show up field=\"a value\"\n  hello world", actual)

	buf.Reset()
	h = createLogHandler("console", buf, slog.LevelDebug)
	h = &noTimestampHandler{h}
	l = slog.New(h)
	slog.SetDefault(l)

	l.Debug("hello")
	log.Println("world")

	actual = strings.TrimSpace(buf.String())
	assert.Equal(t, "DBG hello\nINF world", actual)

	buf.Reset()
	h = createLogHandler("logfmt", buf, slog.LevelInfo)
	h = &noTimestampHandler{h}
	l = slog.New(h)
	slog.SetDefault(l)

	l.Info("hello\"", "field", "a value", "num", 84)

	actual = strings.TrimSpace(buf.String())
	assert.Equal(t, "level=INFO msg=\"hello\\\"\" field=\"a value\" num=84", actual)
}
