package integration

import (
	"testing"
)

func TestBase64_Encode(t *testing.T) {
	inOutTest(t, `{{ "foo" | base64.Encode }}`, "Zm9v")
}

func TestBase64_Decode(t *testing.T) {
	inOutTest(t, `{{ "Zm9v" | base64.Decode }}`, "foo")
}
