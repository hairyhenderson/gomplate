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
	in := map[string]any{
		"foo": "bar",
		"baz": map[string]any{
			"qux": "quux",
		},
	}

	testdata := []struct {
		in  any
		key any
		out bool
	}{
		{in, "foo", true},
		{in, "bar", false},
		{in["baz"], "qux", true},
		{[]string{"foo", "bar", "baz"}, "bar", true},
		{[]any{"foo", "bar", "baz"}, "bar", true},
		{[]any{"foo", "bar", "baz"}, 42, false},
		{[]int{1, 2, 42}, 42, true},
	}

	for _, d := range testdata {
		assert.Equal(t, d.out, Has(d.in, d.key))
	}
}

func TestDict(t *testing.T) {
	testdata := []struct {
		expected map[string]any
		args     []any
	}{
		{expected: map[string]any{}},
		{args: []any{}, expected: map[string]any{}},
		{args: []any{"foo"}, expected: map[string]any{"foo": ""}},
		{args: []any{42}, expected: map[string]any{"42": ""}},
		{args: []any{"foo", nil}, expected: map[string]any{"foo": nil}},
		{args: []any{"foo", "bar"}, expected: map[string]any{"foo": "bar"}},
		{args: []any{"foo", "bar", "baz", true}, expected: map[string]any{
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

	in := map[string]any{
		"foo": 1,
		"bar": 2,
	}
	expected := []string{"bar", "foo"}
	keys, err := Keys(in)
	require.NoError(t, err)
	assert.Equal(t, expected, keys)

	in2 := map[string]any{
		"baz": 3,
		"qux": 4,
	}
	expected = []string{"bar", "foo", "baz", "qux"}
	keys, err = Keys(in, in2)
	require.NoError(t, err)
	assert.Equal(t, expected, keys)

	in3 := map[string]any{
		"Foo": 5,
		"Bar": 6,
		"foo": 7,
		"bar": 8,
	}
	expected = []string{"bar", "foo", "baz", "qux", "Bar", "Foo", "bar", "foo"}
	keys, err = Keys(in, in2, in3)
	require.NoError(t, err)
	assert.Equal(t, expected, keys)
}

func TestValues(t *testing.T) {
	_, err := Values()
	require.Error(t, err)

	in := map[string]any{
		"foo": 1,
		"bar": 2,
	}
	expected := []any{2, 1}
	values, err := Values(in)
	require.NoError(t, err)
	assert.Equal(t, expected, values)

	in2 := map[string]any{
		"baz": 3,
		"qux": 4,
	}
	expected = []any{2, 1, 3, 4}
	values, err = Values(in, in2)
	require.NoError(t, err)
	assert.Equal(t, expected, values)

	in3 := map[string]any{
		"Foo": 5,
		"Bar": 6,
		"foo": 7,
		"bar": 8,
	}
	expected = []any{2, 1, 3, 4, 6, 5, 8, 7}
	values, err = Values(in, in2, in3)
	require.NoError(t, err)
	assert.Equal(t, expected, values)
}

func TestAppend(t *testing.T) {
	out, err := Append(42, []any{})
	require.NoError(t, err)
	assert.Equal(t, []any{42}, out)

	out, err = Append(42, []any{4.9, false, "foo"})
	require.NoError(t, err)
	assert.Equal(t, []any{4.9, false, "foo", 42}, out)

	// a strange but valid use-cases, since we're converting to an []any
	out, err = Append(42, []string{"foo"})
	require.NoError(t, err)
	assert.Equal(t, []any{"foo", 42}, out)

	out, err = Append("baz", []string{"foo", "bar"})
	require.NoError(t, err)
	assert.Equal(t, []any{"foo", "bar", "baz"}, out)
}

func TestPrepend(t *testing.T) {
	out, err := Prepend(42, []any{})
	require.NoError(t, err)
	assert.Equal(t, []any{42}, out)

	out, err = Prepend(42, []any{4.9, false, "foo"})
	require.NoError(t, err)
	assert.Equal(t, []any{42, 4.9, false, "foo"}, out)

	// a strange but valid use-cases, since we're converting to an []any
	out, err = Prepend(42, []string{"foo"})
	require.NoError(t, err)
	assert.Equal(t, []any{42, "foo"}, out)

	out, err = Prepend("foo", []string{"bar", "baz"})
	require.NoError(t, err)
	assert.Equal(t, []any{"foo", "bar", "baz"}, out)
}

func TestUniq(t *testing.T) {
	out, err := Uniq([]any{1, 2, 3, 1, true, false, true, "1", 2})
	require.NoError(t, err)
	assert.Equal(t, []any{1, 2, 3, true, false, "1"}, out)

	out, err = Uniq([]string{"one", "two", "one", "three"})
	require.NoError(t, err)
	assert.Equal(t, []any{"one", "two", "three"}, out)
}

func TestReverse(t *testing.T) {
	out, err := Reverse([]any{})
	require.NoError(t, err)
	assert.Equal(t, []any{}, out)

	out, err = Reverse([]any{8})
	require.NoError(t, err)
	assert.Equal(t, []any{8}, out)

	out, err = Reverse([]any{1, 2, 3, 4})
	require.NoError(t, err)
	assert.Equal(t, []any{4, 3, 2, 1}, out)

	out, err = Reverse([]int{1, 2, 3, 4})
	require.NoError(t, err)
	assert.Equal(t, []any{4, 3, 2, 1}, out)
}

func TestMerge(t *testing.T) {
	dst := map[string]any{}
	src := map[string]any{}
	expected := map[string]any{}

	out, err := Merge(dst, src)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	dst = map[string]any{"a": 4, "c": 5}
	src = map[string]any{"a": 1, "b": 2, "c": 3}
	expected = map[string]any{
		"a": dst["a"], "b": src["b"], "c": dst["c"],
	}

	out, err = Merge(dst, src)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	dst = map[string]any{"a": 4, "c": 5}
	src = map[string]any{"a": 1, "b": 2, "c": 3}
	src2 := map[string]any{"a": 1, "b": 2, "c": 3, "d": 4}
	expected = map[string]any{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	dst = map[string]any{"a": false, "c": 5}
	src = map[string]any{"a": true, "b": 2, "c": 3}
	src2 = map[string]any{"a": true, "b": 2, "c": 3, "d": 4}
	expected = map[string]any{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	dst = map[string]any{"a": true, "c": 5}
	src = map[string]any{
		"a": false,
		"b": map[string]any{
			"ca": "foo",
		},
	}
	src2 = map[string]any{"a": false, "b": 2, "c": 3, "d": 4}
	expected = map[string]any{
		"a": dst["a"], "b": src["b"], "c": dst["c"], "d": src2["d"],
	}

	out, err = Merge(dst, src, src2)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	dst = map[string]any{
		"a": true,
		"b": map[string]any{
			"ca": "foo",
			"cb": "bar",
		},
		"c": 5,
	}
	src = map[string]any{
		"a": false,
		"b": map[string]any{
			"ca": 8,
		},
	}
	expected = map[string]any{
		"a": dst["a"], "b": map[string]any{
			"ca": "foo",
			"cb": "bar",
		}, "c": dst["c"],
	}

	out, err = Merge(dst, src)
	require.NoError(t, err)
	assert.Equal(t, expected, out)
}

type coords struct {
	X, Y int
}

func TestSameTypes(t *testing.T) {
	data := []struct {
		in  []any
		out bool
	}{
		{[]any{}, true},
		{[]any{"a", "b"}, true},
		{[]any{1.0, 3.14}, true},
		{[]any{1, 3}, true},
		{[]any{true, false}, true},
		{[]any{1, 3.0}, false},
		{[]any{"a", nil}, false},
		{[]any{"a", true}, false},
		{[]any{coords{2, 3}, coords{3, 4}}, true},
		{[]any{coords{2, 3}, &coords{3, 4}}, false},
	}

	for _, d := range data {
		assert.Equal(t, d.out, sameTypes(d.in))
	}
}

func TestLessThan(t *testing.T) {
	data := []struct {
		left, right any
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
		{left: map[string]any{"foo": 1}, right: map[string]any{"foo": 2}},
		{
			key:   "foo",
			left:  map[string]any{"foo": 1},
			right: map[string]any{"foo": 2},
			out:   true,
		},
		{
			key:   "bar",
			left:  map[string]any{"foo": 1},
			right: map[string]any{"foo": 2},
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
		in  any
		out []any
	}{
		{
			key: "",
			in:  []string{"b", "c", "a", "d"},
			out: []any{"a", "b", "c", "d"},
		},
		{
			key: "",
			in:  []any{"b", "c", "a", "d"},
			out: []any{"a", "b", "c", "d"},
		},
		{
			key: "",
			in:  []any{"c", "a", "b", 3, 1, 2},
			out: []any{"c", "a", "b", 3, 1, 2},
		},
		{
			key: "",
			in:  nil,
			out: nil,
		},

		{
			key: "",
			in: []map[string]any{
				{"name": "Bart", "age": 12},
				{"age": 1, "name": "Maggie"},
				{"name": "Lisa", "age": 6},
			},
			out: []any{
				map[string]any{"name": "Bart", "age": 12},
				map[string]any{"age": 1, "name": "Maggie"},
				map[string]any{"name": "Lisa", "age": 6},
			},
		},
		{
			key: "name",
			in: []map[string]any{
				{"name": "Bart", "age": 12},
				{"age": 1, "name": "Maggie"},
				{"name": "Lisa", "age": 6},
			},
			out: []any{
				map[string]any{"name": "Bart", "age": 12},
				map[string]any{"name": "Lisa", "age": 6},
				map[string]any{"age": 1, "name": "Maggie"},
			},
		},
		{
			key: "age",
			in: []map[string]any{
				{"name": "Bart", "age": 12},
				{"age": 1, "name": "Maggie"},
				{"name": "Lisa", "age": 6},
			},
			out: []any{
				map[string]any{"age": 1, "name": "Maggie"},
				map[string]any{"name": "Lisa", "age": 6},
				map[string]any{"name": "Bart", "age": 12},
			},
		},
		{
			key: "y",
			in: []map[string]int{
				{"x": 54, "y": 6},
				{"x": 13, "y": -8},
				{"x": 1, "y": 0},
			},
			out: []any{
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
			out: []any{
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
			out: []any{
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
			assert.Equal(t, d.out, out)
		})
	}
}

func TestFlatten(t *testing.T) {
	data := []struct {
		in       any
		expected []any
		depth    int
	}{
		{in: []int{1, 2, 3}, expected: []any{1, 2, 3}},
		{in: [3]int{1, 2, 3}, expected: []any{1, 2, 3}},
		{in: []any{[]string{}, []int{1, 2}, 3}, expected: []any{[]string{}, []int{1, 2}, 3}},
		{
			in:       []any{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []any{[]string{"one"}, [][]int{{1, 2}}, 3},
		},
		{depth: 1, in: []int{1, 2, 3}, expected: []any{1, 2, 3}},
		{depth: 1, in: [3]int{1, 2, 3}, expected: []any{1, 2, 3}},
		{depth: 1, in: []any{[]string{}, []int{1, 2}, 3}, expected: []any{1, 2, 3}},
		{
			depth:    1,
			in:       []any{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []any{"one", []int{1, 2}, 3},
		},
		{depth: 2, in: []int{1, 2, 3}, expected: []any{1, 2, 3}},
		{depth: 2, in: [3]int{1, 2, 3}, expected: []any{1, 2, 3}},
		{depth: 2, in: []any{[]string{}, []int{1, 2}, 3}, expected: []any{1, 2, 3}},
		{
			depth:    2,
			in:       []any{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []any{"one", 1, 2, 3},
		},
		{
			depth: 2,
			in: []any{
				[]string{"one"},
				[]any{
					[]any{
						[]int{1},
						[]any{2, []int{3}},
					},
					[]int{4, 5},
				},
				6,
			},
			expected: []any{"one", []int{1}, []any{2, []int{3}}, 4, 5, 6},
		},
		{depth: -1, in: []int{1, 2, 3}, expected: []any{1, 2, 3}},
		{depth: -1, in: [3]int{1, 2, 3}, expected: []any{1, 2, 3}},
		{depth: -1, in: []any{[]string{}, []int{1, 2}, 3}, expected: []any{1, 2, 3}},
		{
			depth:    -1,
			in:       []any{[]string{"one"}, [][]int{{1, 2}}, 3},
			expected: []any{"one", 1, 2, 3},
		},
		{
			depth: -1,
			in: []any{
				[]string{"one"},
				[]any{
					[]any{
						[]int{1},
						[]any{2, []int{3}},
					},
					[]int{4, 5},
				},
				6,
			},
			expected: []any{"one", 1, 2, 3, 4, 5, 6},
		},
	}

	for _, d := range data {
		out, err := Flatten(d.in, d.depth)
		require.NoError(t, err)
		assert.Equal(t, d.expected, out)
	}

	_, err := Flatten(42, -1)
	require.Error(t, err)
}

func BenchmarkFlatten(b *testing.B) {
	data := []any{
		[]int{1, 2, 3},
		[3]int{1, 2, 3},
		[]any{[]string{}, []int{1, 2}, 3},
		[]any{[]string{"one"}, [][]int{{1, 2}}, 3},
		[]any{
			[]string{"one"},
			[]any{
				[]any{
					[]int{1},
					[]any{2, []int{3}},
				},
				[]int{4, 5},
			},
			6,
		},
	}
	for depth := -1; depth <= 2; depth++ {
		for _, d := range data {
			b.Run(fmt.Sprintf("depth%d %T(%v)", depth, d, d), func(b *testing.B) {
				for b.Loop() {
					Flatten(d, depth)
				}
			})
		}
	}
}

func TestOmit(t *testing.T) {
	in := map[string]any{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	assert.Equal(t, in, Omit(in, "baz"))

	expected := map[string]any{
		"foo": "bar",
		"bar": true,
	}
	assert.Equal(t, expected, Omit(in, ""))

	expected = map[string]any{
		"": "baz",
	}
	assert.Equal(t, expected, Omit(in, "foo", "bar"))

	assert.Equal(t, map[string]any{}, Omit(in, "foo", "bar", ""))
}

func TestPick(t *testing.T) {
	in := map[string]any{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	expected := map[string]any{}
	assert.Equal(t, expected, Pick(in, "baz"))

	expected = map[string]any{
		"": "baz",
	}
	assert.Equal(t, expected, Pick(in, ""))

	expected = map[string]any{
		"foo": "bar",
		"bar": true,
	}
	assert.Equal(t, expected, Pick(in, "foo", "bar"))

	assert.Equal(t, in, Pick(in, "foo", "bar", ""))
}
