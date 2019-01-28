package conv

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)
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
	assert.Equal(t, int64(4096), ToInt64("4,096"))
	assert.Equal(t, int64(-4096), ToInt64("-4,096.00"))
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
	assert.Equal(t, 4096, ToInt("4,096"))
	assert.Equal(t, -4096, ToInt("-4,096.00"))
}

func TestToInt64s(t *testing.T) {
	assert.Equal(t, []int64{}, ToInt64s())

	assert.Equal(t, []int64{0}, ToInt64s(""))
	assert.Equal(t, []int64{0}, ToInt64s("0"))
	assert.Equal(t, []int64{42, 15}, ToInt64s("42", "15"))
	assert.Equal(t, []int64{0, 0, 0, 1, 1, 2, 3, 5, 8, 13, -1000},
		ToInt64s(nil, false, "", true, 1, 2.0, uint8(3), int64(5), float32(8), "13", "-1,000"))
}

func TestToInts(t *testing.T) {
	assert.Equal(t, []int{}, ToInts())

	assert.Equal(t, []int{0}, ToInts(""))
	assert.Equal(t, []int{0}, ToInts("0"))
	assert.Equal(t, []int{42, 15}, ToInts("42", "15"))
	assert.Equal(t, []int{0, 0, 0, 1, 1, 2, 3, 5, 8, 13, 42000},
		ToInts(nil, false, "", true, 1, 2.0, uint8(3), int64(5), float32(8), "13", "42,000"))
}

func TestToFloat64(t *testing.T) {
	z := []interface{}{0, 0.0, nil, false, float32(0), "", "0", "foo", int64(0), uint(0), "0x0", "00", "0,000"}
	for _, n := range z {
		assert.Equal(t, 0.0, ToFloat64(n))
	}
	assert.Equal(t, 1.0, ToFloat64(true))
	z = []interface{}{42, 42.0, float32(42), "42", "42.0", uint8(42), "0x2A", "052"}
	for _, n := range z {
		assert.Equal(t, 42.0, ToFloat64(n))
	}
	z = []interface{}{1000.34, "1000.34", "1,000.34"}
	for _, n := range z {
		assert.Equal(t, 1000.34, ToFloat64(n))
	}
}

func TestToFloat64s(t *testing.T) {
	assert.Equal(t, []float64{}, ToFloat64s())
	assert.Equal(t, []float64{0, 1.0, 2.0, math.Pi, 4.0}, ToFloat64s(nil, true, "2", math.Pi, uint8(4)))
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
	}

	for _, d := range testdata {
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
		args     []interface{}
		expected map[string]interface{}
	}{
		{nil, map[string]interface{}{}},
		{[]interface{}{}, map[string]interface{}{}},
		{[]interface{}{"foo"}, map[string]interface{}{"foo": ""}},
		{[]interface{}{42}, map[string]interface{}{"42": ""}},
		{[]interface{}{"foo", nil}, map[string]interface{}{"foo": nil}},
		{[]interface{}{"foo", "bar"}, map[string]interface{}{"foo": "bar"}},
		{[]interface{}{"foo", "bar", "baz", true}, map[string]interface{}{
			"foo": "bar",
			"baz": true,
		}},
	}

	for _, d := range testdata {
		actual, _ := Dict(d.args...)
		assert.Equal(t, d.expected, actual)
	}
}

func TestKeys(t *testing.T) {
	_, err := Keys()
	assert.Error(t, err)

	in := map[string]interface{}{
		"foo": 1,
		"bar": 2,
	}
	expected := []string{"bar", "foo"}
	keys, err := Keys(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, keys)

	in2 := map[string]interface{}{
		"baz": 3,
		"qux": 4,
	}
	expected = []string{"bar", "foo", "baz", "qux"}
	keys, err = Keys(in, in2)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, keys)

	in3 := map[string]interface{}{
		"Foo": 5,
		"Bar": 6,
		"foo": 7,
		"bar": 8,
	}
	expected = []string{"bar", "foo", "baz", "qux", "Bar", "Foo", "bar", "foo"}
	keys, err = Keys(in, in2, in3)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, keys)
}

func TestValues(t *testing.T) {
	_, err := Values()
	assert.Error(t, err)

	in := map[string]interface{}{
		"foo": 1,
		"bar": 2,
	}
	expected := []interface{}{2, 1}
	values, err := Values(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, values)

	in2 := map[string]interface{}{
		"baz": 3,
		"qux": 4,
	}
	expected = []interface{}{2, 1, 3, 4}
	values, err = Values(in, in2)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, values)

	in3 := map[string]interface{}{
		"Foo": 5,
		"Bar": 6,
		"foo": 7,
		"bar": 8,
	}
	expected = []interface{}{2, 1, 3, 4, 6, 5, 8, 7}
	values, err = Values(in, in2, in3)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, values)
}
