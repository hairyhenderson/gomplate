package gomplate

import (
	"context"
	"testing"
)

func TestSprigFuncs(t *testing.T) {
	funcs := CreateFuncs(context.Background(), nil)
	if _, ok := funcs["semver"]; !ok {
		t.Errorf("semver function not found")
	}
}
