package integration

import (
	"testing"
)

func TestNet_LookupIP(t *testing.T) {
	inOutTest(t, `{{ net.LookupIP "localhost" }}`, "127.0.0.1")
}
