package integration

import (
	"testing"
)

func TestRegexp_Replace(t *testing.T) {
	inOutTest(t, `{{ "1.2.3-59" | regexp.Replace "-([0-9]*)" ".$1" }}`, "1.2.3.59")
}

func TestRegexp_QuoteMeta(t *testing.T) {
	inOutTest(t, "{{ regexp.QuoteMeta `foo{(\\` }}", `foo\{\(\\`)
}
