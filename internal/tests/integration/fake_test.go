package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake output is random, so these templates check the shape of each value and
// render only the result of the check.
func TestFake(t *testing.T) {
	inOutTest(t, `{{ fake.Name | regexp.Match "^\\S+ \\S+$" }}`, "true")
	inOutTest(t, `{{ fake.Email | regexp.Match "^[^@\\s]+@[^@\\s]+\\.[a-z]+$" }}`, "true")
	inOutTest(t, `{{ (net.ParseAddr fake.IPv4Address).Is4 }}`, "true")
	inOutTest(t, `{{ (net.ParseAddr fake.IPv6Address).Is6 }}`, "true")
	inOutTest(t, `{{ fake.Zip | regexp.Match "^[0-9]{5}$" }}`, "true")
	inOutTest(t, `{{ fake.SSN | regexp.Match "^[0-9]{3}-[0-9]{2}-[0-9]{4}$" }}`, "true")
	inOutTest(t, `{{ fake.CreditCardNumber | regexp.Match "^[0-9]{12,19}$" }}`, "true")
	inOutTest(t, `{{ fake.CreditCardExp | regexp.Match "^(0[1-9]|1[0-2])/[0-9]{2}$" }}`, "true")
	inOutTest(t, `{{ fake.URL | regexp.Match "^https?://" }}`, "true")
	inOutTest(t, `{{ gt (len fake.ProductName) 0 }}`, "true")
}

func TestFake_Seeded(t *testing.T) {
	render := func(seed string) string {
		t.Helper()

		o, e, err := cmd(t, "-i", `{{ $f := fake.Seeded `+seed+` }}{{ $f.Name }} {{ $f.Email }} {{ $f.IPv4Address }}`).run()
		require.NoError(t, err)
		assert.Empty(t, e)
		require.NotEmpty(t, o)
		return o
	}

	// The same seed must give the same output in separate runs.
	first := render("42")
	assert.Equal(t, first, render("42"))
	assert.NotEqual(t, first, render("43"))

	inOutTest(t, `{{ $a := fake.Seeded "x" }}{{ $b := "x" | fake.Seeded }}{{ eq $a.Name $b.Name }}`, "true")
}
