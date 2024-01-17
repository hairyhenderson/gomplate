package integration

import (
	"testing"
)

func TestMissingKey_Default(t *testing.T) {
	inOutTest(t, `{{ .name }}`, "<no value>", "--missing-key", "default")
}

func TestMissingKey_Zero(t *testing.T) {
	inOutTest(t, `{{ .name }}`, "<no value>", "--missing-key", "zero")
}

func TestMissingKey_Fallback(t *testing.T) {
	inOutTest(t, `{{ .name | default "Alex" }}`, "Alex", "--missing-key", "default")
}

func TestMissingKey_NotSpecified(t *testing.T) {
	inOutContainsError(t, `{{ .name | default "Alex" }}`, `map has no entry for key \"name\"`)
}

func TestMissingKey_Error(t *testing.T) {
	inOutContainsError(t, `{{ .name | default "Alex" }}`, `map has no entry for key \"name\"`, "--missing-key", "error")
}
