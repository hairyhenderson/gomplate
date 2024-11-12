package funcs

import (
	"context"
	"log/slog"
)

const TraceFuncsLevel = slog.LevelDebug - 1

func trace(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, TraceFuncsLevel, msg, attrs...)
}
