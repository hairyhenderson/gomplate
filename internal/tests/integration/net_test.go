package integration

import (
	"testing"
)

func TestNet_LookupIP(t *testing.T) {
	inOutTest(t, `{{ net.LookupIP "localhost" }}`, "127.0.0.1")
}

func TestNet_CIDRHost(t *testing.T) {
	inOutTestExperimental(t, `{{ net.ParseIPPrefix "10.12.127.0/20" | net.CIDRHost 16 }}`, "10.12.112.16")
	inOutTestExperimental(t, `{{ "10.12.127.0/20" | net.CIDRHost 16 }}`, "10.12.112.16")
	inOutTestExperimental(t, `{{ net.CIDRHost 268 "10.12.127.0/20" }}`, "10.12.113.12")
	inOutTestExperimental(t, `{{ net.CIDRHost 34 "fd00:fd12:3456:7890:00a2::/72" }}`, "fd00:fd12:3456:7890::22")
}

func TestNet_CIDRNetmask(t *testing.T) {
	inOutTestExperimental(t, `{{ "10.12.127.0/20" | net.CIDRNetmask }}`, "255.255.240.0")
	inOutTestExperimental(t, `{{ net.CIDRNetmask "10.0.0.0/12" }}`, "255.240.0.0")
	inOutTestExperimental(t, `{{ net.CIDRNetmask "fd00:fd12:3456:7890:00a2::/72" }}`, "ffff:ffff:ffff:ffff:ff00::")
}

func TestNet_CIDRSubnets(t *testing.T) {
	inOutTestExperimental(t, `{{ index ("10.0.0.0/16" | net.CIDRSubnets 2) 1 }}`, "10.0.64.0/18")
	inOutTestExperimental(t, `{{ range net.CIDRSubnets 2 "10.0.0.0/16" }}
{{ . }}{{ end }}`, `
10.0.0.0/18
10.0.64.0/18
10.0.128.0/18
10.0.192.0/18`)
}

func TestNet_CIDRSubnetSizes(t *testing.T) {
	inOutTestExperimental(t, `{{ index ("10.0.0.0/16" | net.CIDRSubnetSizes 1) 0 }}`, "10.0.0.0/17")
	inOutTestExperimental(t, `{{ index ("10.1.0.0/16" | net.CIDRSubnetSizes 4 4 8 4) 1 }}`, "10.1.16.0/20")
	inOutTestExperimental(t, `{{ range net.CIDRSubnetSizes 16 16 16 32 "fd00:fd12:3456:7890::/56" }}
{{ . }}{{ end }}`, `
fd00:fd12:3456:7800::/72
fd00:fd12:3456:7800:100::/72
fd00:fd12:3456:7800:200::/72
fd00:fd12:3456:7800:300::/88`)
}
