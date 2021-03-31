package integration

import (
	"testing"
)

func TestMath(t *testing.T) {
	inOutTest(t, `{{ math.Add 1 2 3 4 }} {{ add -5 5 }}`, "10 0")
	inOutTest(t, `{{ math.Sub 10 5 }} {{ sub -5 5 }}`, "5 -10")
	inOutTest(t, `{{ math.Mul 1 2 3 4 }} {{ mul -5 5 }}`, "24 -25")
	inOutTest(t, `{{ math.Div 5 2 }} {{ div -5 5 }}`, "2.5 -1")
	inOutTest(t, `{{ math.Rem 5 3 }} {{ rem 2 2 }}`, "2 0")
	inOutTest(t, `{{ math.Pow 8 4 }} {{ pow 2 2 }}`, "4096 4")
	inOutTest(t, `{{ math.Seq 0 }}, {{ seq 0 3 }}, {{ seq -5 -10 2 }}`,
		`[1 0], [0 1 2 3], [-5 -7 -9]`)
	inOutTest(t, `{{ math.Round 0.99 }}, {{ math.Round "foo" }}, {{math.Round 3.5}}`,
		`1, 0, 4`)
	inOutTest(t, `{{ math.Max -0 "+Inf" "NaN" }}, {{ math.Max 3.4 3.401 3.399 }}`,
		`+Inf, 3.401`)
}
