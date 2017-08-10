package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	assert.False(t, Bool(""))
	assert.False(t, Bool("asdf"))
	assert.False(t, Bool("1234"))
	assert.False(t, Bool("False"))
	assert.False(t, Bool("0"))
	assert.False(t, Bool("false"))
	assert.False(t, Bool("F"))
	assert.False(t, Bool("f"))
	assert.True(t, Bool("true"))
	assert.True(t, Bool("True"))
	assert.True(t, Bool("t"))
	assert.True(t, Bool("T"))
	assert.True(t, Bool("1"))
}

func TestSlice(t *testing.T) {
	expected := []string{"foo", "bar"}
	actual := Slice("foo", "bar")
	assert.Equal(t, expected[0], actual[0])
	assert.Equal(t, expected[1], actual[1])
}

func TestJoin(t *testing.T) {

	assert.Equal(t, "foo,bar", Join([]interface{}{"foo", "bar"}, ","))
	assert.Equal(t, "foo,\nbar", Join([]interface{}{"foo", "bar"}, ",\n"))
	// Join handles all kinds of scalar types too...
	assert.Equal(t, "42-18446744073709551615", Join([]interface{}{42, uint64(18446744073709551615)}, "-"))
	assert.Equal(t, "1,,true,3.14,foo,nil", Join([]interface{}{1, "", true, 3.14, "foo", nil}, ","))
	// and best-effort with weird types
	assert.Equal(t, "[foo],bar", Join([]interface{}{[]string{"foo"}, "bar"}, ","))
}

func TestHas(t *testing.T) {

	in := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]interface{}{
			"qux": "quux",
		},
	}

	assert.True(t, Has(in, "foo"))
	assert.False(t, Has(in, "bar"))
	assert.True(t, Has(in["baz"], "qux"))
}

func TestMustParseInt(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0, int(MustParseInt(i, 10, 64)))
	}
	assert.Equal(t, 1, int(MustParseInt("1", 10, 64)))
	assert.Equal(t, -1, int(MustParseInt("-1", 10, 64)))
}

func TestMustAtoi(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0, MustAtoi(i))
	}
	assert.Equal(t, 1, MustAtoi("1"))
	assert.Equal(t, -1, MustAtoi("-1"))
}

func TestMustParseUint(t *testing.T) {
	for _, i := range []string{"0", "-0", "-1", "foo", "", "*&^%"} {
		assert.Equal(t, uint64(0), MustParseUint(i, 10, 64))
	}
	assert.Equal(t, uint64(1), MustParseUint("1", 10, 64))
}

func TestMustParseFloat(t *testing.T) {
	for _, i := range []string{"0", "-0", "foo", "", "*&^%"} {
		assert.Equal(t, 0.0, MustParseFloat(i, 64))
	}
	assert.Equal(t, 1.0, MustParseFloat("1", 64))
	assert.Equal(t, -1.0, MustParseFloat("-1", 64))
}
