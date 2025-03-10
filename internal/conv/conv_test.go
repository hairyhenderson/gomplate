package conv

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestInterfaceSlice(t *testing.T) {
	data := []struct {
		in, expected any
	}{
		{[]int{1, 2, 3}, []any{1, 2, 3}},
		{[3]int{1, 2, 3}, []any{1, 2, 3}},
		{[]string{"foo", "bar", "baz"}, []any{"foo", "bar", "baz"}},
		{[3]string{"foo", "bar", "baz"}, []any{"foo", "bar", "baz"}},
		{[]any{[]string{}, []int{1, 2}, 3}, []any{[]string{}, []int{1, 2}, 3}},
		{[3]any{[]string{}, []int{1, 2}, 3}, []any{[]string{}, []int{1, 2}, 3}},
	}

	for _, d := range data {
		out, err := InterfaceSlice(d.in)
		assert.NilError(t, err)
		assert.DeepEqual(t, d.expected, out)
	}

	_, err := InterfaceSlice(42)
	assert.ErrorContains(t, err, "")
}

func BenchmarkInterfaceSlice(b *testing.B) {
	data := []any{
		[]int{1, 2, 3},
		[3]int{1, 2, 3},
		[]string{"foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz"},
		[12]string{"foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz"},
		[]any{[]string{}, []int{1, 2}, 3},
		[3]any{[]string{}, []int{1, 2}, 3},
	}

	for _, d := range data {
		b.Run(fmt.Sprintf("%T(%v)", d, d), func(b *testing.B) {
			for b.Loop() {
				InterfaceSlice(d)
			}
		})
	}
}
