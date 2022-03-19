package integration

import (
	"testing"
)

func TestNet_LookupIP(t *testing.T) {
	inOutTest(t, `{{ net.LookupIP "localhost" }}`, "127.0.0.1")
}

func TestNet_ParseCIDR(t *testing.T) {
	inOutTest(t, `{{ net.StdParseCIDR "10.12.127.0/20" }}`, "10.12.112.0/20")
}

func TestNet_CidrHost(t *testing.T) {
	inOutTest(t, `{{ net.StdParseCIDR "10.12.127.0/20" | net.CidrHost 16 }}`, "10.12.112.16")
	inOutTest(t, `{{ "10.12.127.0/20" | net.CidrHost 16 }}`, "10.12.112.16")
	inOutTest(t, `{{ net.CidrHost 268 "10.12.127.0/20" }}`, "10.12.113.12")
	inOutTest(t, `{{ net.CidrHost 34 "fd00:fd12:3456:7890:00a2::/72" }}`, "fd00:fd12:3456:7890::22")
}

func TestNet_CidrNetmask(t *testing.T) {
	inOutTest(t, `{{ "10.12.127.0/20" | net.CidrNetmask }}`, "255.255.240.0")
	inOutTest(t, `{{ net.CidrNetmask "10.0.0.0/12" }}`, "255.240.0.0")
}

func TestNet_CidrSubnets(t *testing.T) {
	inOutTest(t, `{{ index ("10.0.0.0/16" | net.CidrSubnets 2) 1 }}`, "10.0.64.0/18")
	inOutTest(t, `{{ range net.CidrSubnets 2 "10.0.0.0/16" }}
{{ . }}{{ end }}`, `
10.0.0.0/18
10.0.64.0/18
10.0.128.0/18
10.0.192.0/18`)
}

func TestNet_CidrSubnetSizes(t *testing.T) {
	inOutTest(t, `{{ index ("10.0.0.0/16" | net.CidrSubnetSizes 1) 0 }}`, "10.0.0.0/17")
	inOutTest(t, `{{ index ("10.1.0.0/16" | net.CidrSubnetSizes 4 4 8 4) 1 }}`, "10.1.16.0/20")
	inOutTest(t, `{{ range net.CidrSubnetSizes 16 16 16 32 "fd00:fd12:3456:7890::/56" }}
{{ . }}{{ end }}`, `
fd00:fd12:3456:7800::/72
fd00:fd12:3456:7800:100::/72
fd00:fd12:3456:7800:200::/72
fd00:fd12:3456:7800:300::/88`)
}
