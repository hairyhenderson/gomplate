package conv

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBool(t *testing.T) {
	testdata := []struct {
		in  string
		out bool
	}{
		{"", false},
		{"asdf", false},
		{"1234", false},
		{"False", false},
		{"0", false},
		{"false", false},
		{"F", false},
		{"f", false},
		{"true", true},
		{"True", true},
		{"t", true},
		{"T", true},
		{"1", true},
	}
	for _, d := range testdata {
		assert.Equal(t, d.out, Bool(d.in))
	}
}

func TestSlice(t *testing.T) {
	expected := []string{"foo", "bar"}
	actual := Slice("foo", "bar")
	assert.Equal(t, expected[0], actual[0])
	assert.Equal(t, expected[1], actual[1])
}

func TestJoin(t *testing.T) {
	testdata := []struct {
		in  interface{}
		sep string
		out string
	}{
		{[]interface{}{"foo", "bar"}, ",", "foo,bar"},
		{[]interface{}{"foo", "bar"}, ",\n", "foo,\nbar"},
		// Join handles all kinds of scalar types too...
		{[]interface{}{42, uint64(18446744073709551615)}, "-", "42-18446744073709551615"},
		{[]int{42, 100}, ",", "42,100"},
		{[]int64{42, 100}, ",", "42,100"},
		{[]uint64{42, 100}, ",", "42,100"},
		{[]bool{true, false}, ",", "true,false"},
		{[]float64{1, 2}, ",", "1,2"},
		{[]interface{}{1, "", true, 3.14, "foo", nil}, ",", "1,,true,3.14,foo,nil"},
		// and best-effort with weird types
		{[]interface{}{[]string{"foo"}, "bar"}, ",", "[foo],bar"},
	}
	for _, d := range testdata {
		out, err := Join(d.in, d.sep)
		require.NoError(t, err)
		assert.Equal(t, d.out, out)
	}
}

func TestHas(t *testing.T) {
	in := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]interface{}{
			"qux": "quux",
		},
	}

	testdata := []struct {
		in  interface{}
		key interface{}
		out bool
	}{
		{in, "foo", true},
		{in, "bar", false},
		{in["baz"], "qux", true},
		{[]string{"foo", "bar", "baz"}, "bar", true},
		{[]interface{}{"foo", "bar", "baz"}, "bar", true},
		{[]interface{}{"foo", "bar", "baz"}, 42, false},
		{[]int{1, 2, 42}, 42, true},
	}

	for _, d := range testdata {
		assert.Equal(t, d.out, Has(d.in, d.key))
	}
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
		assert.Zero(t, MustParseFloat(i, 64))
	}
	assert.InEpsilon(t, 1.0, MustParseFloat("1", 64), 1e-12)
	assert.InEpsilon(t, -1.0, MustParseFloat("-1", 64), 1e-12)
}

func TestToInt64(t *testing.T) {
	actual, err := ToInt64(1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = ToInt64(int32(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = ToInt64(int64(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = ToInt64(float32(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = ToInt64(float64(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = ToInt64(42)
	require.NoError(t, err)
	assert.Equal(t, int64(42), actual)

	actual, err = ToInt64("42.0")
	require.NoError(t, err)
	assert.Equal(t, int64(42), actual)

	actual, err = ToInt64("3.5")
	require.NoError(t, err)
	assert.Equal(t, int64(3), actual)

	actual, err = ToInt64(uint64(math.MaxUint64))
	require.NoError(t, err)
	assert.Equal(t, int64(-1), actual)

	actual, err = ToInt64(uint8(math.MaxUint8))
	require.NoError(t, err)
	assert.Equal(t, int64(0xFF), actual)

	actual, err = ToInt64(false)
	require.NoError(t, err)
	assert.Equal(t, int64(0), actual)

	actual, err = ToInt64(true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), actual)

	actual, err = ToInt64("0xFFFF")
	require.NoError(t, err)
	assert.Equal(t, int64(0xFFFF), actual)

	actual, err = ToInt64("010")
	require.NoError(t, err)
	assert.Equal(t, int64(8), actual)

	actual, err = ToInt64("4,096")
	require.NoError(t, err)
	assert.Equal(t, int64(4096), actual)

	actual, err = ToInt64("-4,096.00")
	require.NoError(t, err)
	assert.Equal(t, int64(-4096), actual)

	t.Run("error cases", func(t *testing.T) {
		_, err = ToInt64(nil)
		require.Error(t, err)

		_, err = ToInt64("")
		require.Error(t, err)

		_, err = ToInt64("foo")
		require.Error(t, err)
	})
}

func TestToInt(t *testing.T) {
	actual, err := ToInt(1)
	require.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = ToInt(int32(1))
	require.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = ToInt(int64(1))
	require.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = ToInt(float32(1))
	require.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = ToInt(float64(1))
	require.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = ToInt(42)
	require.NoError(t, err)
	assert.Equal(t, 42, actual)

	actual, err = ToInt(uint64(math.MaxUint64))
	require.NoError(t, err)
	assert.Equal(t, -1, actual)

	actual, err = ToInt(uint8(math.MaxUint8))
	require.NoError(t, err)
	assert.Equal(t, 0xFF, actual)

	actual, err = ToInt(false)
	require.NoError(t, err)
	assert.Equal(t, 0, actual)

	actual, err = ToInt(true)
	require.NoError(t, err)
	assert.Equal(t, 1, actual)

	actual, err = ToInt("0xFFFF")
	require.NoError(t, err)
	assert.Equal(t, 0xFFFF, actual)

	actual, err = ToInt("010")
	require.NoError(t, err)
	assert.Equal(t, 8, actual)

	actual, err = ToInt("4,096")
	require.NoError(t, err)
	assert.Equal(t, 4096, actual)

	actual, err = ToInt("-4,096.00")
	require.NoError(t, err)
	assert.Equal(t, -4096, actual)

	t.Run("error cases", func(t *testing.T) {
		_, err = ToInt(nil)
		require.Error(t, err)

		_, err = ToInt("")
		require.Error(t, err)

		_, err = ToInt("foo")
		require.Error(t, err)
	})
}

func TestToInt64s(t *testing.T) {
	actual, err := ToInt64s()
	require.NoError(t, err)
	assert.Equal(t, []int64{}, actual)

	actual, err = ToInt64s("0")
	require.NoError(t, err)
	assert.Equal(t, []int64{0}, actual)

	actual, err = ToInt64s("42", "15")
	require.NoError(t, err)
	assert.Equal(t, []int64{42, 15}, actual)

	actual, err = ToInt64s(false, true, 1, 2.0, uint8(3), int64(5), float32(8), "13", "-1,000")
	require.NoError(t, err)
	assert.Equal(t, []int64{0, 1, 1, 2, 3, 5, 8, 13, -1000}, actual)

	t.Run("error cases", func(t *testing.T) {
		_, err = ToInt64s("")
		require.Error(t, err)

		_, err = ToInt64s(nil, false, "", true, 1, 2.0, uint8(3), int64(5), float32(8), "13", "-1,000")
		require.Error(t, err)
	})
}

func TestToInts(t *testing.T) {
	actual, err := ToInts()
	require.NoError(t, err)
	assert.Equal(t, []int{}, actual)

	actual, err = ToInts("0")
	require.NoError(t, err)
	assert.Equal(t, []int{0}, actual)

	actual, err = ToInts("42", "15")
	require.NoError(t, err)
	assert.Equal(t, []int{42, 15}, actual)

	actual, err = ToInts(false, true, 1, 2.0, uint8(3), int64(5), float32(8), "13", "42,000")
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 1, 2, 3, 5, 8, 13, 42000}, actual)

	t.Run("error cases", func(t *testing.T) {
		_, err = ToInts("")
		require.Error(t, err)

		_, err = ToInts(nil, false, "", true, 1, 2.0, uint8(3), int64(5), float32(8), "13", "42,000")
		require.Error(t, err)
	})
}

func TestToFloat64(t *testing.T) {
	z := []interface{}{nil, "", "foo"}
	for _, n := range z {
		_, err := ToFloat64(n)
		require.Error(t, err)
	}

	z = []interface{}{0, 0.0, false, float32(0), "0", int64(0), uint(0), "0x0", "00", "0,000"}
	for _, n := range z {
		actual, err := ToFloat64(n)
		require.NoError(t, err)
		assert.Zero(t, actual)
	}

	actual, err := ToFloat64(true)
	require.NoError(t, err)
	assert.InEpsilon(t, 1.0, actual, 1e-12)

	z = []interface{}{42, 42.0, float32(42), "42", "42.0", uint8(42), "0x2A", "052"}
	for _, n := range z {
		actual, err = ToFloat64(n)
		require.NoError(t, err)
		assert.InEpsilon(t, 42.0, actual, 1e-12)
	}

	z = []interface{}{1000.34, "1000.34", "1,000.34"}
	for _, n := range z {
		actual, err = ToFloat64(n)
		require.NoError(t, err)
		assert.InEpsilon(t, 1000.34, actual, 1e-12)
	}
}

func TestToFloat64s(t *testing.T) {
	actual, err := ToFloat64s()
	require.NoError(t, err)
	assert.Equal(t, []float64{}, actual)

	actual, err = ToFloat64s(true, "2", math.Pi, uint8(4))
	require.NoError(t, err)
	assert.Equal(t, []float64{1.0, 2.0, math.Pi, 4.0}, actual)

	_, err = ToFloat64s(nil, true, "2", math.Pi, uint8(4))
	require.Error(t, err)
}

type foo struct {
	val string
}

func (f foo) String() string {
	return f.val
}

func TestToString(t *testing.T) {
	var p *string
	f := "foo"
	p = &f

	var n *string

	testdata := []struct {
		in  interface{}
		out string
	}{
		{nil, "nil"},
		{"", ""},
		{"foo", "foo"},
		{true, "true"},
		{42, "42"},
		{3.14, "3.14"},
		{-127, "-127"},
		{0xFF, "255"},
		{uint8(42), "42"},
		{math.Pi, "3.141592653589793"},
		{math.NaN(), "NaN"},
		{math.Inf(1), "+Inf"},
		{math.Inf(-1), "-Inf"},
		{foo{"bar"}, "bar"},
		{p, "foo"},
		{fmt.Errorf("hi"), "hi"},
		{n, "<nil>"},
		{[]byte("hello world"), "hello world"},
	}

	for _, d := range testdata {
		d := d
		t.Run(fmt.Sprintf("%T/%#v == %s", d.in, d.in, d.out), func(t *testing.T) {
			out := ToString(d.in)
			assert.Equal(t, d.out, out)
		})
	}
}

func TestToBool(t *testing.T) {
	trueData := []interface{}{
		true,
		1,
		int8(1),
		uint8(1),
		int32(1),
		uint32(1),
		int64(1),
		uint64(1),
		float32(1),
		float64(1),
		"1",
		"0x1",
		"1.0",
		"01",
		"true",
		"True",
		"T",
		"t",
		"TrUe",
		"yes",
		"YES",
	}
	for _, d := range trueData {
		out := ToBool(d)
		assert.True(t, out)
	}

	falseData := []interface{}{
		nil,
		false,
		42,
		uint64(math.MaxUint64),
		uint8(math.MaxUint8),
		"",
		"false",
		"foo",
		"0xFFFF",
		"010",
		"4,096",
		"-4,096.00",
	}
	for _, d := range falseData {
		out := ToBool(d)
		assert.False(t, out)
	}
}

func TestDict(t *testing.T) {
	testdata := []struct {
		expected map[string]interface{}
		args     []interface{}
	}{
		{expected: map[string]interface{}{}},
		{
			args:     []interface{}{},
			expected: map[string]interface{}{},
		},
		{
			args:     []interface{}{"foo"},
			expected: map[string]interface{}{"foo": ""},
		},
		{
			args:     []interface{}{42},
			expected: map[string]interface{}{"42": ""},
		},
		{
			args:     []interface{}{"foo", nil},
			expected: map[string]interface{}{"foo": nil},
		},
		{
			args:     []interface{}{"foo", "bar"},
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			args: []interface{}{"foo", "bar", "baz", true},
			expected: map[string]interface{}{
				"foo": "bar",
				"baz": true,
			},
		},
	}

	for _, d := range testdata {
		actual, _ := Dict(d.args...)
		assert.Equal(t, d.expected, actual)
	}
}
