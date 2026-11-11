package funcs

import (
	"context"
	"maps"
	"net/mail"
	"net/netip"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFakeNS(t *testing.T) *FakeFuncs {
	t.Helper()

	return CreateFakeFuncs(t.Context())["fake"].(func() any)().(*FakeFuncs)
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

	// Every no-argument string method is a template function, so each one
	// must be covered above.
	var methods []string
	typ := reflect.TypeFor[*FakeFuncs]()
	stringType := reflect.TypeFor[string]()
	for i := range typ.NumMethod() {
		m := typ.Method(i)
		if m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && m.Type.Out(0) == stringType {
			methods = append(methods, m.Name)
		}
	}
	assert.Equal(t, methods, slices.Sorted(maps.Keys(tests)))

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
