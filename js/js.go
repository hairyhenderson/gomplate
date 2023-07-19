package js

import (
	_ "embed"

	"github.com/robertkrimen/otto/registry"
)

//go:embed k8s.js
var k8s string

//go:embed shared.js
var shared string

func init() {
	_ = registry.Register(func() string { return k8s })
	_ = registry.Register(func() string { return shared })
}
