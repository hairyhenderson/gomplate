//+build integration

package integration

import (
	. "gopkg.in/check.v1"
)

type TplSuite struct{}

var _ = Suite(&TplSuite{})

func (s *TplSuite) TestTime(c *C) {
	inOutTest(c, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- tpl "{{ add .first .second }}" $nums }}`,
		"15")

	inOutTest(c, `
		{{- $nums := dict "first" 5 "second" 10 }}
		{{- $othernums := dict "first" 18 "second" -8 }}
		{{- tpl "T" "{{ add .first .second }}" $nums }}
		{{- template "T" $othernums }}`,
		"1510")
}
