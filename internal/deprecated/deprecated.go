package deprecated

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rs/zerolog"
)

// WarnDeprecated - use this to warn about deprecated template functions or
// datasources
func WarnDeprecated(ctx context.Context, msg string) {
	logger := zerolog.Ctx(ctx)
	if !logger.Warn().Enabled() {
		// we'll flip to slog soon, but in the meantime if we don't have a
		// logger in the context, just log it
		slog.WarnContext(ctx, fmt.Sprintf("Deprecated: %s", msg))
	}
	logger.Warn().Msgf("Deprecated: %s", msg)
}
