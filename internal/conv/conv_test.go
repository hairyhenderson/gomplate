package conv

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestInterfaceSlice(t *testing.T) {
	data := []struct {
		in, expected interface{}
	}{
		{[]int{1, 2, 3}, []interface{}{1, 2, 3}},
		{[3]int{1, 2, 3}, []interface{}{1, 2, 3}},
		{[]string{"foo", "bar", "baz"}, []interface{}{"foo", "bar", "baz"}},
		{[3]string{"foo", "bar", "baz"}, []interface{}{"foo", "bar", "baz"}},
		{[]interface{}{[]string{}, []int{1, 2}, 3}, []interface{}{[]string{}, []int{1, 2}, 3}},
		{[3]interface{}{[]string{}, []int{1, 2}, 3}, []interface{}{[]string{}, []int{1, 2}, 3}},
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
	data := []interface{}{
		[]int{1, 2, 3},
		[3]int{1, 2, 3},
		[]string{"foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz"},
		[12]string{"foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz", "foo", "bar", "baz"},
		[]interface{}{[]string{}, []int{1, 2}, 3},
		[3]interface{}{[]string{}, []int{1, 2}, 3},
	}

	for _, d := range data {
		b.Run(fmt.Sprintf("%T(%v)", d, d), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				InterfaceSlice(d)
			}
		})
	}
}
