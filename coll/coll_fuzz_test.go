package coll

import (
	"reflect"
	"testing"

	"github.com/hairyhenderson/gomplate/v5/internal/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fuzzList parses in as a YAML array, twice - one copy to pass to the
// function under test, and one to check that the function didn't modify it.
func fuzzList(t *testing.T, in string) (list, orig []any) {
	t.Helper()

	list, err := parsers.YAMLArray(in)
	if err != nil || len(list) > 100 {
		t.Skip()
	}

	orig, err = parsers.YAMLArray(in)
	require.NoError(t, err)

	// NaN isn't equal to itself, so changes couldn't be detected
	if !reflect.DeepEqual(list, orig) {
		t.Skip()
	}

	return list, orig
}

func FuzzListFuncs(f *testing.F) {
	f.Add("[3, 1, 2]", "", 1)
	f.Add("[b, a, c, a]", "", -1)
	f.Add("[[1, [2, [3]]], 4]", "", 2)
	f.Add("[{a: 2}, {a: 1}]", "a", 0)
	f.Add("[{a: 1}, {b: 2}]", "a", 0)
	f.Add("[1, 1.5, true, null]", "", 0)
	f.Add("[null, a]", "", 0)
	f.Add(" ", "", 0)
	f.Add("[{a: 1}, {a: 1.5}, {b: 2}]", "a", 0)
	f.Add("[.nan, [.nan]]", "", 0)

	f.Fuzz(func(t *testing.T, in, key string, depth int) {
		list, orig := fuzzList(t, in)

		sorted, err := Sort(key, list)
		if err == nil {
			assert.ElementsMatch(t, orig, sorted)
		}

		uniq, err := Uniq(list)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(uniq), len(list))
		assert.Subset(t, list, uniq)

		rev, err := Reverse(list)
		require.NoError(t, err)
		require.Len(t, rev, len(orig))
		for i := range rev {
			assert.Equal(t, orig[len(orig)-1-i], rev[i])
		}

		if len(list) > 0 {
			omitted, oerr := OmitSlice(list, list[0])
			require.NoError(t, oerr)
			assert.NotContains(t, omitted, list[0])
		}

		_, err = Flatten(list, depth)
		require.NoError(t, err)

		// none of these should modify the input
		assert.True(t, reflect.DeepEqual(orig, list), "input was modified: %v != %v", orig, list)
	})
}

// fuzzMap parses in as a YAML map, twice - see fuzzList.
func fuzzMap(t *testing.T, in string) (m, orig map[string]any) {
	t.Helper()

	m, err := parsers.YAML(in)
	if err != nil || m == nil {
		t.Skip()
	}

	orig, err = parsers.YAML(in)
	require.NoError(t, err)

	if !reflect.DeepEqual(m, orig) {
		t.Skip()
	}

	return m, orig
}

func FuzzMapFuncs(f *testing.F) {
	f.Add("a: 1\nb: {c: 2}\n", "b: {d: 3}\ne: 4\n", "a")
	f.Add("a: {b: 1}\n", "a: 2\n", "b")
	f.Add("a: 1\n", "a: {b: 1}\n", "a")
	f.Add("a: .nan\n", "b: [.nan]\n", "a")

	f.Fuzz(func(t *testing.T, dstIn, srcIn, key string) {
		dst, dstOrig := fuzzMap(t, dstIn)
		src, srcOrig := fuzzMap(t, srcIn)

		merged, err := Merge(dst, src)
		require.NoError(t, err)

		// every key from both maps is in the result, and dst wins
		for k, v := range dstOrig {
			assert.Contains(t, merged, k)
			if _, isMap := v.(map[string]any); !isMap {
				assert.Equal(t, v, merged[k], k)
			}
		}
		for k := range srcOrig {
			assert.Contains(t, merged, k)
		}

		picked := Pick(dst, key)
		omitted := Omit(dst, key)
		assert.Equal(t, len(dst), len(picked)+len(omitted))

		// none of these should modify the input
		assert.True(t, reflect.DeepEqual(dstOrig, dst), "dst was modified: %v != %v", dstOrig, dst)
		assert.True(t, reflect.DeepEqual(srcOrig, src), "src was modified: %v != %v", srcOrig, src)
	})
}
