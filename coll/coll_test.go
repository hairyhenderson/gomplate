package coll

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	expected := []string{"foo", "bar"}
	actual := Slice("foo", "bar")
	assert.Equal(t, expected[0], actual[0])
	assert.Equal(t, expected[1], actual[1])
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

func TestAppend(t *testing.T) {
	out, err := Append(42, []interface{}{})
	assert.NoError(t, err)
	assert.EqualValues(t, out, []interface{}{42})

	out, err = Append(42, []interface{}{4.9, false, "foo"})
	assert.NoError(t, err)
	assert.EqualValues(t, out, []interface{}{4.9, false, "foo", 42})

	// a strange but valid use-cases, since we're converting to an []interface{}
	out, err = Append(42, []string{"foo"})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{"foo", 42}, out)

	out, err = Append("baz", []string{"foo", "bar"})
	assert.NoError(t, err)
	assert.EqualValues(t, out, []interface{}{"foo", "bar", "baz"})
}

func TestPrepend(t *testing.T) {
	out, err := Prepend(42, []interface{}{})
	assert.NoError(t, err)
	assert.EqualValues(t, out, []interface{}{42})

	out, err = Prepend(42, []interface{}{4.9, false, "foo"})
	assert.NoError(t, err)
	assert.EqualValues(t, out, []interface{}{42, 4.9, false, "foo"})

	// a strange but valid use-cases, since we're converting to an []interface{}
	out, err = Prepend(42, []string{"foo"})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{42, "foo"}, out)

	out, err = Prepend("foo", []string{"bar", "baz"})
	assert.NoError(t, err)
	assert.EqualValues(t, out, []interface{}{"foo", "bar", "baz"})
}

func TestUniq(t *testing.T) {
	out, err := Uniq([]interface{}{1, 2, 3, 1, true, false, true, "1", 2})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{1, 2, 3, true, false, "1"}, out)

	out, err = Uniq([]string{"one", "two", "one", "three"})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{"one", "two", "three"}, out)
}

func TestReverse(t *testing.T) {
	out, err := Reverse([]interface{}{})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{}, out)

	out, err = Reverse([]interface{}{8})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{8}, out)

	out, err = Reverse([]interface{}{1, 2, 3, 4})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{4, 3, 2, 1}, out)

	out, err = Reverse([]int{1, 2, 3, 4})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{4, 3, 2, 1}, out)
}

func TestMerge(t *testing.T) {
	dst := map[string]interface{}{}
	src := map[string]interface{}{}
	expected := map[string]interface{}{}

	out, err := Merge(dst, src)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": 4, "c": 5}
	src = map[string]interface{}{"a": 1, "b": 2, "c": 3}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"],
	}

	out, err = Merge(dst, src)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": 4, "c": 5}
	src = map[string]interface{}{"a": 1, "b": 2, "c": 3}
	src2 := map[string]interface{}{"a": 1, "b": 2, "c": 3, "d": 4}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": false, "c": 5}
	src = map[string]interface{}{"a": true, "b": 2, "c": 3}
	src2 = map[string]interface{}{"a": true, "b": 2, "c": 3, "d": 4}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": true, "c": 5}
	src = map[string]interface{}{"a": false,
		"b": map[string]interface{}{
			"ca": "foo",
		},
	}
	src2 = map[string]interface{}{"a": false, "b": 2, "c": 3, "d": 4}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{
		"a": true,
		"b": map[string]interface{}{
			"ca": "foo",
			"cb": "bar",
		},
		"c": 5}
	src = map[string]interface{}{
		"a": false,
		"b": map[string]interface{}{
			"ca": 8,
		},
	}
	expected = map[string]interface{}{
		"a": dst["a"], "b": map[string]interface{}{
			"ca": "foo",
			"cb": "bar",
		}, "c": dst["c"],
	}

	out, err = Merge(dst, src)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)
}

type coords struct {
	X, Y int
}

func TestSameTypes(t *testing.T) {
	data := []struct {
		in  []interface{}
		out bool
	}{
		{[]interface{}{}, true},
		{[]interface{}{"a", "b"}, true},
		{[]interface{}{1.0, 3.14}, true},
		{[]interface{}{1, 3}, true},
		{[]interface{}{true, false}, true},
		{[]interface{}{1, 3.0}, false},
		{[]interface{}{"a", nil}, false},
		{[]interface{}{"a", true}, false},
		{[]interface{}{coords{2, 3}, coords{3, 4}}, true},
		{[]interface{}{coords{2, 3}, &coords{3, 4}}, false},
	}

	for _, d := range data {
		assert.Equal(t, d.out, sameTypes(d.in))
	}
}

func TestLessThan(t *testing.T) {
	data := []struct {
		key         string
		left, right interface{}
		out         bool
	}{
		{"", nil, nil, false},
		{"", "a", "b", true},
		{"", "a", "a", false},
		{"", "b", "a", false},
		{"", 1.00, 3.14, true},
		{"", 'a', 'A', false},
		{"", 'a', 'b', true},
		{"", uint(0xff), uint(0x32), false},
		{"", 1, 3, true},
		{"", true, false, false},
		{"", map[string]interface{}{"foo": 1}, map[string]interface{}{"foo": 2}, false},
		{"foo", map[string]interface{}{"foo": 1}, map[string]interface{}{"foo": 2}, true},
		{"bar", map[string]interface{}{"foo": 1}, map[string]interface{}{"foo": 2}, false},
		{"X", coords{}, coords{-1, 2}, false},
		{"Y", &coords{1, 1}, &coords{-1, 2}, true},
		{"", &coords{1, 1}, &coords{-1, 2}, false},
		{"foo", &coords{1, 1}, &coords{-1, 2}, false},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf(`LessThan("%s")(<%T>%#v,%#v)==%v`, d.key, d.left, d.left, d.right, d.out), func(t *testing.T) {
			assert.Equal(t, d.out, lessThan(d.key)(d.left, d.right))
		})
	}
}

func TestSort(t *testing.T) {
	out, err := Sort("", 42)
	assert.Error(t, err)
	assert.Nil(t, out)

	data := []struct {
		key string
		in  interface{}
		out []interface{}
	}{
		{
			key: "",
			in:  []string{"b", "c", "a", "d"},
			out: []interface{}{"a", "b", "c", "d"},
		},
		{
			key: "",
			in:  []interface{}{"b", "c", "a", "d"},
			out: []interface{}{"a", "b", "c", "d"},
		},
		{
			key: "",
			in:  []interface{}{"c", "a", "b", 3, 1, 2},
			out: []interface{}{"c", "a", "b", 3, 1, 2},
		},
		{
			key: "",
			in:  nil,
			out: nil,
		},

		{
			key: "",
			in: []map[string]interface{}{
				{"name": "Bart", "age": 12},
				{"age": 1, "name": "Maggie"},
				{"name": "Lisa", "age": 6},
			},
			out: []interface{}{
				map[string]interface{}{"name": "Bart", "age": 12},
				map[string]interface{}{"age": 1, "name": "Maggie"},
				map[string]interface{}{"name": "Lisa", "age": 6},
			},
		},
		{
			key: "name",
			in: []map[string]interface{}{
				{"name": "Bart", "age": 12},
				{"age": 1, "name": "Maggie"},
				{"name": "Lisa", "age": 6},
			},
			out: []interface{}{
				map[string]interface{}{"name": "Bart", "age": 12},
				map[string]interface{}{"name": "Lisa", "age": 6},
				map[string]interface{}{"age": 1, "name": "Maggie"},
			},
		},
		{
			key: "age",
			in: []map[string]interface{}{
				{"name": "Bart", "age": 12},
				{"age": 1, "name": "Maggie"},
				{"name": "Lisa", "age": 6},
			},
			out: []interface{}{
				map[string]interface{}{"age": 1, "name": "Maggie"},
				map[string]interface{}{"name": "Lisa", "age": 6},
				map[string]interface{}{"name": "Bart", "age": 12},
			},
		},
		{
			key: "y",
			in: []map[string]int{
				{"x": 54, "y": 6},
				{"x": 13, "y": -8},
				{"x": 1, "y": 0},
			},
			out: []interface{}{
				map[string]int{"x": 13, "y": -8},
				map[string]int{"x": 1, "y": 0},
				map[string]int{"x": 54, "y": 6},
			},
		},
		{
			key: "X",
			in: []coords{
				{2, 4},
				{3, 3},
				{1, 5},
			},
			out: []interface{}{
				coords{1, 5},
				coords{2, 4},
				coords{3, 3},
			},
		},
		{
			key: "X",
			in: []*coords{
				{2, 4},
				{3, 3},
				{1, 5},
			},
			out: []interface{}{
				&coords{1, 5},
				&coords{2, 4},
				&coords{3, 3},
			},
		},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf(`Sort("%s",<%T>)==%#v`, d.key, d.in, d.out), func(t *testing.T) {
			out, err := Sort(d.key, d.in)
			assert.NoError(t, err)
			assert.EqualValues(t, d.out, out)
		})
	}
}
