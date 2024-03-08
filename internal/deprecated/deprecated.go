package deprecated

import (
	"context"
	"fmt"
	"log/slog"
)

// WarnDeprecated - use this to warn about deprecated template functions or
// datasources
func WarnDeprecated(ctx context.Context, msg string) {
	slog.WarnContext(ctx, fmt.Sprintf("Deprecated: %s", msg))
}
