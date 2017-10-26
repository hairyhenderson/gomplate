package conv

import (
	"math"
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

func TestToInt64(t *testing.T) {
	assert.Equal(t, int64(1), ToInt64(1))
	assert.Equal(t, int64(1), ToInt64(int32(1)))
	assert.Equal(t, int64(1), ToInt64(int64(1)))
	assert.Equal(t, int64(1), ToInt64(float32(1)))
	assert.Equal(t, int64(1), ToInt64(float64(1)))
	assert.Equal(t, int64(42), ToInt64(42))
	assert.Equal(t, int64(42), ToInt64("42.0"))
	assert.Equal(t, int64(3), ToInt64("3.5"))
	assert.Equal(t, int64(-1), ToInt64(uint64(math.MaxUint64)))
	assert.Equal(t, int64(0xFF), ToInt64(uint8(math.MaxUint8)))

	assert.Equal(t, int64(0), ToInt64(nil))
	assert.Equal(t, int64(0), ToInt64(false))
	assert.Equal(t, int64(1), ToInt64(true))
	assert.Equal(t, int64(0), ToInt64(""))
	assert.Equal(t, int64(0), ToInt64("foo"))
	assert.Equal(t, int64(0xFFFF), ToInt64("0xFFFF"))
	assert.Equal(t, int64(8), ToInt64("010"))
}

func TestToInt(t *testing.T) {
	assert.Equal(t, 1, ToInt(1))
	assert.Equal(t, 1, ToInt(int32(1)))
	assert.Equal(t, 1, ToInt(int64(1)))
	assert.Equal(t, 1, ToInt(float32(1)))
	assert.Equal(t, 1, ToInt(float64(1)))
	assert.Equal(t, 42, ToInt(42))
	assert.Equal(t, -1, ToInt(uint64(math.MaxUint64)))
	assert.Equal(t, 0xFF, ToInt(uint8(math.MaxUint8)))

	assert.Equal(t, 0, ToInt(nil))
	assert.Equal(t, 0, ToInt(false))
	assert.Equal(t, 1, ToInt(true))
	assert.Equal(t, 0, ToInt(""))
	assert.Equal(t, 0, ToInt("foo"))
	assert.Equal(t, 0xFFFF, ToInt("0xFFFF"))
	assert.Equal(t, 8, ToInt("010"))
}

func TestToInt64s(t *testing.T) {
	assert.Equal(t, []int64{}, ToInt64s())

	assert.Equal(t, []int64{0}, ToInt64s(""))
	assert.Equal(t, []int64{0}, ToInt64s("0"))
	assert.Equal(t, []int64{42, 15}, ToInt64s("42", "15"))
	assert.Equal(t, []int64{0, 0, 0, 1, 1, 2, 3, 5, 8, 13},
		ToInt64s(nil, false, "", true, 1, 2.0, uint8(3), int64(5), float32(8), "13"))
}

func TestToInts(t *testing.T) {
	assert.Equal(t, []int{}, ToInts())

	assert.Equal(t, []int{0}, ToInts(""))
	assert.Equal(t, []int{0}, ToInts("0"))
	assert.Equal(t, []int{42, 15}, ToInts("42", "15"))
	assert.Equal(t, []int{0, 0, 0, 1, 1, 2, 3, 5, 8, 13},
		ToInts(nil, false, "", true, 1, 2.0, uint8(3), int64(5), float32(8), "13"))
}

func TestToFloat64(t *testing.T) {
	z := []interface{}{0, 0.0, nil, false, float32(0), "", "0", "foo", int64(0), uint(0), "0x0", "00"}
	for _, n := range z {
		assert.Equal(t, 0.0, ToFloat64(n))
	}
	assert.Equal(t, 1.0, ToFloat64(true))
	z = []interface{}{42, 42.0, float32(42), "42", "42.0", uint8(42), "0x2A", "052"}
	for _, n := range z {
		assert.Equal(t, 42.0, ToFloat64(n))
	}
}

func TestToFloat64s(t *testing.T) {
	assert.Equal(t, []float64{}, ToFloat64s())
	assert.Equal(t, []float64{0, 1.0, 2.0, math.Pi, 4.0}, ToFloat64s(nil, true, "2", math.Pi, uint8(4)))
}
