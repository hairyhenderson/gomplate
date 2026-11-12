package funcs

import (
	"context"
	"fmt"
	"maps"
	"net/mail"
	"net/netip"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFakeNS(t *testing.T) *FakeFuncs {
	t.Helper()

	f := CreateFakeFuncs(t.Context())["fake"].(func() any)().(*FakeFuncs)
	f.now = fixedClock(2026)
	return f
}

func fixedClock(year int) func() time.Time {
	return func() time.Time { return time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC) }
}

func TestCreateFakeFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateFakeFuncs(ctx)
			actual := fmap["fake"].(func() any)().(*FakeFuncs)

			assert.Equal(t, ctx, actual.ctx)
			assert.NotSame(t, gofakeit.GlobalFaker, actual.faker)
			assert.NotNil(t, actual.now)
		})
	}
}

func TestFakeFuncs(t *testing.T) {
	t.Parallel()

	f := testFakeNS(t)

	// check is nil when only a non-empty result can be asserted.
	tests := map[string]struct {
		fn    func() string
		check func(t *testing.T, s string)
	}{
		"Name":       {f.Name, matches(`^\S+ \S+$`)},
		"FirstName":  {f.FirstName, nil},
		"MiddleName": {f.MiddleName, nil},
		"LastName":   {f.LastName, nil},
		"NamePrefix": {f.NamePrefix, nil},
		"NameSuffix": {f.NameSuffix, nil},
		"Hobby":      {f.Hobby, nil},

		"Email":          {f.Email, isEmail},
		"Phone":          {f.Phone, matches(`^[0-9]{10}$`)},
		"PhoneFormatted": {f.PhoneFormatted, matches(`^[0-9().-]{12,}$`)},

		"SSN":              {f.SSN, isSSN},
		"CreditCardNumber": {f.CreditCardNumber, isCardNumber},
		"CreditCardExp":    {f.CreditCardExp, matches(`^(0[1-9]|1[0-2])/[0-9]{2}$`)},

		"Street":         {f.Street, matches(`^[0-9]+ \S`)},
		"City":           {f.City, nil},
		"State":          {f.State, nil},
		"StateAbr":       {f.StateAbr, matches(`^[A-Z]{2}$`)},
		"Zip":            {f.Zip, matches(`^[0-9]{5}$`)},
		"Country":        {f.Country, nil},
		"CountryAbr":     {f.CountryAbr, matches(`^[A-Z]{2}$`)},
		"TimeZoneRegion": {f.TimeZoneRegion, matches(`^[A-Z][A-Za-z]+(/[A-Za-z0-9_+-]+)+$`)},

		"CompanySuffix": {f.CompanySuffix, nil},
		"JobTitle":      {f.JobTitle, nil},
		"BuzzWord":      {f.BuzzWord, nil},
		"BS":            {f.BS, nil},
		"Slogan":        {f.Slogan, nil},

		"DomainName":   {f.DomainName, matches(`^[a-z0-9-]+\.[a-z]+$`)},
		"DomainSuffix": {f.DomainSuffix, matches(`^[a-z]+$`)},
		"URL":          {f.URL, isHTTPURL},
		"IPv4Address":  {f.IPv4Address, isIPv4},
		"IPv6Address":  {f.IPv6Address, isIPv6},
		"HTTPMethod":   {f.HTTPMethod, matches(`^[A-Z]+$`)},
		"UserAgent":    {f.UserAgent, matches(`^(Mozilla|Opera)/`)},

		"ProductName":        {f.ProductName, nil},
		"ProductCategory":    {f.ProductCategory, nil},
		"ProductDescription": {f.ProductDescription, nil},
		"Word":               {f.Word, nil},
		"Phrase":             {f.Phrase, nil},
	}

	// Every generator must be covered above.
	assert.Equal(t, generatorNames(), slices.Sorted(maps.Keys(tests)))

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for range 100 {
				s := tt.fn()
				require.NotEmpty(t, s)
				if tt.check != nil {
					tt.check(t, s)
				}
			}
		})
	}
}

func TestFakeFuncs_CreditCardExp(t *testing.T) {
	t.Parallel()

	// 2095 checks that the year wraps to two digits.
	for _, year := range []int{2026, 2095} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			t.Parallel()

			f := testFakeNS(t)
			f.now = fixedClock(year)

			want := make([]string, 0, 10)
			for offset := 1; offset <= 10; offset++ {
				want = append(want, fmt.Sprintf("%02d", (year+offset)%100))
			}

			for range 200 {
				exp := f.CreditCardExp()
				require.Regexp(t, `^(0[1-9]|1[0-2])/[0-9]{2}$`, exp)
				assert.Contains(t, want, exp[3:])
			}
		})
	}
}

func TestFakeFuncs_RandomlySeeded(t *testing.T) {
	t.Parallel()

	a, b := testFakeNS(t), testFakeNS(t)

	// With a fixed or shared seed both namespaces would produce the same
	// sequence; with independent random seeds a match is vanishingly unlikely.
	as, bs := make([]string, 10), make([]string, 10)
	for i := range as {
		as[i], bs[i] = a.Name(), b.Name()
	}
	assert.NotEqual(t, as, bs)
}

func TestFakeFuncs_Seeded(t *testing.T) {
	t.Parallel()

	f := testFakeNS(t)

	// Seeds are compared as strings, so 42 and "42" are the same seed.
	a, b := f.Seeded(42), f.Seeded("42")
	assert.Equal(t, f.ctx, a.ctx)
	assert.NotSame(t, f.faker, a.faker)

	// Run every generator, so that one whose output isn't fully determined
	// by the source fails here.
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	for _, name := range generatorNames() {
		for range 20 {
			want := va.MethodByName(name).Call(nil)[0].String()
			got := vb.MethodByName(name).Call(nil)[0].String()
			require.Equal(t, want, got, name)
		}
	}

	c, d := f.Seeded(1), f.Seeded(2)
	cs, ds := make([]string, 10), make([]string, 10)
	for i := range cs {
		cs[i], ds[i] = c.Name(), d.Name()
	}
	assert.NotEqual(t, cs, ds)
}

func TestFakeFuncs_Concurrent(t *testing.T) {
	t.Parallel()

	f := testFakeNS(t)

	// Run with -race: gomplate's Renderer can be used from several
	// goroutines, and they share one namespace.
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 500 {
				assert.NotEmpty(t, f.Name())
				assert.NotEmpty(t, f.Email())
				assert.NotEmpty(t, f.ProductDescription())
			}
		})
	}
	wg.Wait()
}

// generatorNames returns the sorted names of the namespace's generators: its
// methods that take no arguments and return a string.
func generatorNames() []string {
	var names []string
	typ := reflect.TypeFor[*FakeFuncs]()
	for i := range typ.NumMethod() {
		m := typ.Method(i)
		if m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && m.Type.Out(0) == reflect.TypeFor[string]() {
			names = append(names, m.Name)
		}
	}
	return names
}

func matches(pattern string) func(*testing.T, string) {
	return func(t *testing.T, s string) {
		t.Helper()
		assert.Regexp(t, pattern, s)
	}
}

func isEmail(t *testing.T, s string) {
	t.Helper()

	addr, err := mail.ParseAddress(s)
	require.NoError(t, err)
	assert.Equal(t, s, addr.Address)
}

func isHTTPURL(t *testing.T, s string) {
	t.Helper()

	u, err := url.Parse(s)
	require.NoError(t, err)
	assert.Contains(t, []string{"http", "https"}, u.Scheme)
	assert.NotEmpty(t, u.Host)
}

// isSSN checks that s could have been issued by the SSA: the area is never
// 000, 666 or 900-999, the group never 00 and the serial never 0000.
func isSSN(t *testing.T, s string) {
	t.Helper()

	require.Regexp(t, `^[0-9]{3}-[0-9]{2}-[0-9]{4}$`, s)
	area, group, serial := s[:3], s[4:6], s[7:]
	assert.NotContains(t, []string{"000", "666"}, area, s)
	assert.Less(t, area, "900", s)
	assert.NotEqual(t, "00", group, s)
	assert.NotEqual(t, "0000", serial, s)
}

func isCardNumber(t *testing.T, s string) {
	t.Helper()

	require.Regexp(t, `^[0-9]{12,19}$`, s)

	// Luhn check: double every second digit from the right.
	sum := 0
	for i := range len(s) {
		d := int(s[len(s)-1-i] - '0')
		if i%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	assert.Zero(t, sum%10, "%s fails the Luhn check", s)
}

func isIPv4(t *testing.T, s string) {
	t.Helper()

	addr, err := netip.ParseAddr(s)
	require.NoError(t, err)
	assert.True(t, addr.Is4(), s)
}

func isIPv6(t *testing.T, s string) {
	t.Helper()

	addr, err := netip.ParseAddr(s)
	require.NoError(t, err)
	assert.True(t, addr.Is6() && !addr.Is4In6(), s)
}
