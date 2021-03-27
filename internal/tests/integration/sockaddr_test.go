package integration

import (
	"testing"
)

func TestSockaddr(t *testing.T) {
	inOutContains(t, `{{ range (sockaddr.GetAllInterfaces | sockaddr.Include "type" "ipv4") -}}
{{ . | sockaddr.Attr "address" }}
{{end}}`, "127.0.0.1")
}
