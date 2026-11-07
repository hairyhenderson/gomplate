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

func TestMissingKey_JSONPath_Default(t *testing.T) {
	inOutTest(t, `{{ dict "foo" "bar" | jsonpath ".bogus" }}`, "<no value>", "--missing-key", "default")
}

func TestMissingKey_JSONPath_Zero(t *testing.T) {
	inOutTest(t, `{{ dict "foo" "bar" | jsonpath ".bogus" }}`, "<no value>", "--missing-key", "zero")
}

func TestMissingKey_JSONPath_Fallback(t *testing.T) {
	inOutTest(t, `{{ dict "foo" "bar" | jsonpath ".bogus" | default "Alex" }}`, "Alex", "--missing-key", "default")
}

func TestMissingKey_JSONPath_NotSpecified(t *testing.T) {
	inOutContainsError(t, `{{ dict "foo" "bar" | jsonpath ".bogus" }}`, `bogus is not found`)
}

func TestMissingKey_JSONPath_Error(t *testing.T) {
	inOutContainsError(t, `{{ dict "foo" "bar" | jsonpath ".bogus" }}`, `bogus is not found`, "--missing-key", "error")
}

func TestMissingKey_JSONPath_Namespace(t *testing.T) {
	inOutTest(t, `{{ dict "foo" "bar" | coll.JSONPath ".bogus" }}`, "<no value>", "--missing-key", "default")
}

func TestMissingKey_JSONPath_NestedMissing(t *testing.T) {
	inOutTest(t, `{{ dict "foo" (dict "bar" "val") | jsonpath "$.foo.baz.qux" }}`, "<no value>", "--missing-key", "default")
}

func TestMissingKey_JSONPath_NestedMissingError(t *testing.T) {
	inOutContainsError(t, `{{ dict "foo" (dict "bar" "val") | jsonpath "$.foo.baz.qux" }}`, `baz is not found`, "--missing-key", "error")
}

func TestMissingKey_JSONPath_ArrayBounds(t *testing.T) {
	inOutTest(t, `{{ dict "foo" (coll.Slice 1 2) | jsonpath "$.foo[99]" }}`, "<no value>", "--missing-key", "default")
}

func TestMissingKey_JSONPath_ArrayBoundsError(t *testing.T) {
	inOutContainsError(t, `{{ dict "foo" (coll.Slice 1 2) | jsonpath "$.foo[99]" }}`, `array index out of bounds`, "--missing-key", "error")
}

func TestMissingKey_JSONPath_Valid(t *testing.T) {
	inOutTest(t, `{{ dict "foo" "bar" | jsonpath ".foo" }}`, "bar", "--missing-key", "default")
	inOutTest(t, `{{ dict "foo" "bar" | jsonpath ".foo" }}`, "bar", "--missing-key", "error")
}
