package deprecated

import (
	"context"

	"github.com/rs/zerolog"
)

// WarnDeprecated - use this to warn about deprecated template functions or
// datasources
func WarnDeprecated(ctx context.Context, msg string) {
	logger := zerolog.Ctx(ctx)
	logger.Warn().Msgf("Deprecated: %s", msg)
}
