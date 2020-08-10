package funcs

import (
	"context"
	"fmt"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

func checkExperimental(ctx context.Context) error {
	if !config.FromContext(ctx).Experimental {
		return fmt.Errorf("experimental function, but experimental mode not enabled")
	}
	return nil
}
