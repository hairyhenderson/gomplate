package funcs

import (
	"context"
	"fmt"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
)

func checkExperimental(ctx context.Context) error {
	if !config.ExperimentalEnabled(ctx) {
		return fmt.Errorf("experimental function, but experimental mode not enabled")
	}
	return nil
}
