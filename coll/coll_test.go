package coll

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		expected map[string]interface{}
		args     []interface{}
	}{
		{expected: map[string]interface{}{}},
		{args: []interface{}{}, expected: map[string]interface{}{}},
		{args: []interface{}{"foo"}, expected: map[string]interface{}{"foo": ""}},
		{args: []interface{}{42}, expected: map[string]interface{}{"42": ""}},
		{args: []interface{}{"foo", nil}, expected: map[string]interface{}{"foo": nil}},
		{args: []interface{}{"foo", "bar"}, expected: map[string]interface{}{"foo": "bar"}},
		{args: []interface{}{"foo", "bar", "baz", true}, expected: map[string]interface{}{
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
	require.Error(t, err)

	in := map[string]interface{}{
		"foo": 1,
		"bar": 2,
	}
	expected := []string{"bar", "foo"}
	keys, err := Keys(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, keys)

	in2 := map[string]interface{}{
		"baz": 3,
		"qux": 4,
	}
	expected = []string{"bar", "foo", "baz", "qux"}
	keys, err = Keys(in, in2)
	require.NoError(t, err)
	assert.EqualValues(t, expected, keys)

	in3 := map[string]interface{}{
		"Foo": 5,
		"Bar": 6,
		"foo": 7,
		"bar": 8,
	}
	expected = []string{"bar", "foo", "baz", "qux", "Bar", "Foo", "bar", "foo"}
	keys, err = Keys(in, in2, in3)
	require.NoError(t, err)
	assert.EqualValues(t, expected, keys)
}

func TestValues(t *testing.T) {
	_, err := Values()
	require.Error(t, err)

	in := map[string]interface{}{
		"foo": 1,
		"bar": 2,
	}
	expected := []interface{}{2, 1}
	values, err := Values(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, values)

	in2 := map[string]interface{}{
		"baz": 3,
		"qux": 4,
	}
	expected = []interface{}{2, 1, 3, 4}
	values, err = Values(in, in2)
	require.NoError(t, err)
	assert.EqualValues(t, expected, values)

	in3 := map[string]interface{}{
		"Foo": 5,
		"Bar": 6,
		"foo": 7,
		"bar": 8,
	}
	expected = []interface{}{2, 1, 3, 4, 6, 5, 8, 7}
	values, err = Values(in, in2, in3)
	require.NoError(t, err)
	assert.EqualValues(t, expected, values)
}

func TestAppend(t *testing.T) {
	out, err := Append(42, []interface{}{})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{42}, out)

	out, err = Append(42, []interface{}{4.9, false, "foo"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{4.9, false, "foo", 42}, out)

	// a strange but valid use-cases, since we're converting to an []interface{}
	out, err = Append(42, []string{"foo"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{"foo", 42}, out)

	out, err = Append("baz", []string{"foo", "bar"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{"foo", "bar", "baz"}, out)
}

func TestPrepend(t *testing.T) {
	out, err := Prepend(42, []interface{}{})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{42}, out)

	out, err = Prepend(42, []interface{}{4.9, false, "foo"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{42, 4.9, false, "foo"}, out)

	// a strange but valid use-cases, since we're converting to an []interface{}
	out, err = Prepend(42, []string{"foo"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{42, "foo"}, out)

	out, err = Prepend("foo", []string{"bar", "baz"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{"foo", "bar", "baz"}, out)
}

func TestUniq(t *testing.T) {
	out, err := Uniq([]interface{}{1, 2, 3, 1, true, false, true, "1", 2})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{1, 2, 3, true, false, "1"}, out)

	out, err = Uniq([]string{"one", "two", "one", "three"})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{"one", "two", "three"}, out)
}

func TestReverse(t *testing.T) {
	out, err := Reverse([]interface{}{})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{}, out)

	out, err = Reverse([]interface{}{8})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{8}, out)

	out, err = Reverse([]interface{}{1, 2, 3, 4})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{4, 3, 2, 1}, out)

	out, err = Reverse([]int{1, 2, 3, 4})
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{4, 3, 2, 1}, out)
}

func TestMerge(t *testing.T) {
	dst := map[string]interface{}{}
	src := map[string]interface{}{}
	expected := map[string]interface{}{}

	out, err := Merge(dst, src)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": 4, "c": 5}
	src = map[string]interface{}{"a": 1, "b": 2, "c": 3}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"],
	}

	out, err = Merge(dst, src)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": 4, "c": 5}
	src = map[string]interface{}{"a": 1, "b": 2, "c": 3}
	src2 := map[string]interface{}{"a": 1, "b": 2, "c": 3, "d": 4}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": false, "c": 5}
	src = map[string]interface{}{"a": true, "b": 2, "c": 3}
	src2 = map[string]interface{}{"a": true, "b": 2, "c": 3, "d": 4}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{"a": true, "c": 5}
	src = map[string]interface{}{
		"a": false,
		"b": map[string]interface{}{
			"ca": "foo",
		},
	}
	src2 = map[string]interface{}{"a": false, "b": 2, "c": 3, "d": 4}
	expected = map[string]interface{}{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	dst = map[string]interface{}{
		"a": true,
		"b": map[string]interface{}{
			"ca": "foo",
			"cb": "bar",
		},
		"c": 5,
	}
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
	require.NoError(t, err)
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
		left, right interface{}
		key         string
		out         bool
	}{
		{key: ""},
		{left: "a", right: "b", out: true},
		{left: "a", right: "a"},
		{left: "b", right: "a"},
		{left: 1.00, right: 3.14, out: true},
		{left: 'a', right: 'A'},
		{left: 'a', right: 'b', out: true},
		{left: uint(0xff), right: uint(0x32)},
		{left: 1, right: 3, out: true},
		{left: true, right: false, out: false},
		{left: map[string]interface{}{"foo": 1}, right: map[string]interface{}{"foo": 2}},
		{
			key:   "foo",
			left:  map[string]interface{}{"foo": 1},
			right: map[string]interface{}{"foo": 2},
			out:   true,
		},
		{
			key:   "bar",
			left:  map[string]interface{}{"foo": 1},
			right: map[string]interface{}{"foo": 2},
		},
		{key: "X", left: coords{}, right: coords{-1, 2}},
		{key: "Y", left: &coords{1, 1}, right: &coords{-1, 2}, out: true},
		{left: &coords{1, 1}, right: &coords{-1, 2}},
		{key: "foo", left: &coords{1, 1}, right: &coords{-1, 2}},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf(`LessThan("%s")(<%T>%#v,%#v)==%v`, d.key, d.left, d.left, d.right, d.out), func(t *testing.T) {
			assert.Equal(t, d.out, lessThan(d.key)(d.left, d.right))
		})
	}
}

func TestSort(t *testing.T) {
	out, err := Sort("", 42)
	require.Error(t, err)
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
			require.NoError(t, err)
			assert.EqualValues(t, d.out, out)
		})
	}
}

func TestFlatten(t *testing.T) {
	data := []struct {
		in       interface{}
		expected []interface{}
		depth    int
	}{
		{in: []int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{in: [3]int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{in: []interface{}{[]string{}, []int{1, 2}, 3}, expected: []interface{}{[]string{}, []int{1, 2}, 3}},
		{
			in:       []interface{}{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []interface{}{[]string{"one"}, [][]int{{1, 2}}, 3},
		},
		{depth: 1, in: []int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{depth: 1, in: [3]int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{depth: 1, in: []interface{}{[]string{}, []int{1, 2}, 3}, expected: []interface{}{1, 2, 3}},
		{
			depth:    1,
			in:       []interface{}{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []interface{}{"one", []int{1, 2}, 3},
		},
		{depth: 2, in: []int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{depth: 2, in: [3]int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{depth: 2, in: []interface{}{[]string{}, []int{1, 2}, 3}, expected: []interface{}{1, 2, 3}},
		{
			depth:    2,
			in:       []interface{}{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []interface{}{"one", 1, 2, 3},
		},
		{
			depth: 2,
			in: []interface{}{
				[]string{"one"},
				[]interface{}{
					[]interface{}{
						[]int{1},
						[]interface{}{2, []int{3}},
					},
					[]int{4, 5},
				},
				6,
			},
			expected: []interface{}{"one", []int{1}, []interface{}{2, []int{3}}, 4, 5, 6},
		},
		{depth: -1, in: []int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{depth: -1, in: [3]int{1, 2, 3}, expected: []interface{}{1, 2, 3}},
		{depth: -1, in: []interface{}{[]string{}, []int{1, 2}, 3}, expected: []interface{}{1, 2, 3}},
		{
			depth:    -1,
			in:       []interface{}{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []interface{}{"one", 1, 2, 3},
		},
		{
			depth: -1,
			in: []interface{}{
				[]string{"one"},
				[]interface{}{
					[]interface{}{
						[]int{1},
						[]interface{}{2, []int{3}},
					},
					[]int{4, 5},
				},
				6,
			},
			expected: []interface{}{"one", 1, 2, 3, 4, 5, 6},
		},
	}

	for _, d := range data {
		out, err := Flatten(d.in, d.depth)
		require.NoError(t, err)
		assert.EqualValues(t, d.expected, out)
	}

	_, err := Flatten(42, -1)
	require.Error(t, err)
}

func BenchmarkFlatten(b *testing.B) {
	data := []interface{}{
		[]int{1, 2, 3},
		[3]int{1, 2, 3},
		[]interface{}{[]string{}, []int{1, 2}, 3},
		[]interface{}{[]string{"one"}, [][]int{{1, 2}}, 3},
		[]interface{}{
			[]string{"one"},
			[]interface{}{
				[]interface{}{
					[]int{1},
					[]interface{}{2, []int{3}},
				},
				[]int{4, 5},
			},
			6,
		},
	}
	for depth := -1; depth <= 2; depth++ {
		for _, d := range data {
			b.Run(fmt.Sprintf("depth%d %T(%v)", depth, d, d), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					Flatten(d, depth)
				}
			})
		}
	}
}

func TestOmit(t *testing.T) {
	in := map[string]interface{}{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	assert.EqualValues(t, in, Omit(in, "baz"))

	expected := map[string]interface{}{
		"foo": "bar",
		"bar": true,
	}
	assert.EqualValues(t, expected, Omit(in, ""))

	expected = map[string]interface{}{
		"": "baz",
	}
	assert.EqualValues(t, expected, Omit(in, "foo", "bar"))

	assert.EqualValues(t, map[string]interface{}{}, Omit(in, "foo", "bar", ""))
}

func TestPick(t *testing.T) {
	in := map[string]interface{}{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	expected := map[string]interface{}{}
	assert.EqualValues(t, expected, Pick(in, "baz"))

	expected = map[string]interface{}{
		"": "baz",
	}
	assert.EqualValues(t, expected, Pick(in, ""))

	expected = map[string]interface{}{
		"foo": "bar",
		"bar": true,
	}
	assert.EqualValues(t, expected, Pick(in, "foo", "bar"))

	assert.EqualValues(t, in, Pick(in, "foo", "bar", ""))
}
