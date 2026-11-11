package integration

import (
	"testing"
)

// fake output is random, so these templates check the shape of each value and
// render only the result of the check.
func TestFake(t *testing.T) {
	inOutTest(t, `{{ fake.Name | regexp.Match "^\\S+ \\S+$" }}`, "true")
	inOutTest(t, `{{ fake.Email | regexp.Match "^[^@\\s]+@[^@\\s]+\\.[a-z]+$" }}`, "true")
	inOutTest(t, `{{ (net.ParseAddr fake.IPv4Address).Is4 }}`, "true")
	inOutTest(t, `{{ (net.ParseAddr fake.IPv6Address).Is6 }}`, "true")
	inOutTest(t, `{{ fake.Zip | regexp.Match "^[0-9]{5}$" }}`, "true")
	inOutTest(t, `{{ fake.URL | regexp.Match "^https?://" }}`, "true")
	inOutTest(t, `{{ gt (len fake.ProductName) 0 }}`, "true")
}
