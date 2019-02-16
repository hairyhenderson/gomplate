//+build integration

package integration

import (
	. "gopkg.in/check.v1"
)

type TmplSuite struct{}

var _ = Suite(&TmplSuite{})

func (s *TmplSuite) TestInline(c *C) {
	inOutTest(c, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- tpl "{{ add .first .second }}" $nums }}`,
		"15")

	inOutTest(c, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- $othernums := dict "first" 18 "second" -8 }}
		{{- tmpl.Inline "T" "{{ add .first .second }}" $nums }}
		{{- template "T" $othernums }}`,
		"1510")
}
