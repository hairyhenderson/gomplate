package coll

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	m  = map[string]any
	ar = []any
)

func TestJSONPath(t *testing.T) {
	in := m{
		"store": m{
			"book": ar{
				m{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				m{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				m{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				m{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": m{
				"color": "red",
				"price": 19.95,
			},
		},
	}
	out, err := JSONPath(".store.bicycle.color", in)
	require.NoError(t, err)
	assert.Equal(t, "red", out)

	out, err = JSONPath(".store.bicycle.price", in)
	require.NoError(t, err)
	assert.InEpsilon(t, 19.95, out, 1e-12)

	_, err = JSONPath(".store.bogus", in)
	require.Error(t, err)

	_, err = JSONPath("{.store.unclosed", in)
	require.Error(t, err)

	out, err = JSONPath(".store", in)
	require.NoError(t, err)
	assert.Equal(t, in["store"], out)

	out, err = JSONPath("$.store.book[*].author", in)
	require.NoError(t, err)
	assert.Len(t, out, 4)
	assert.Contains(t, out, "Nigel Rees")
	assert.Contains(t, out, "Evelyn Waugh")
	assert.Contains(t, out, "Herman Melville")
	assert.Contains(t, out, "J. R. R. Tolkien")

	out, err = JSONPath("$..book[?( @.price < 10.0 )]", in)
	require.NoError(t, err)
	expected := ar{
		m{
			"category": "reference",
			"author":   "Nigel Rees",
			"title":    "Sayings of the Century",
			"price":    8.95,
		},
		m{
			"category": "fiction",
			"author":   "Herman Melville",
			"title":    "Moby Dick",
			"isbn":     "0-553-21311-3",
			"price":    8.99,
		},
	}
	assert.Equal(t, expected, out)

	in = m{
		"a": m{
			"aa": m{
				"foo": m{
					"aaa": m{
						"aaaa": m{
							"bar": 1234,
						},
					},
				},
			},
			"ab": m{
				"aba": m{
					"foo": m{
						"abaa": true,
						"abab": "baz",
					},
				},
			},
		},
	}
	out, err = JSONPath("..foo.*", in)
	require.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Contains(t, out, m{"aaaa": m{"bar": 1234}})
	assert.Contains(t, out, true)
	assert.Contains(t, out, "baz")

	type bicycleType struct {
		Color string
	}
	type storeType struct {
		Bicycle *bicycleType
		safe    any
	}

	structIn := &storeType{
		Bicycle: &bicycleType{
			Color: "red",
		},
		safe: "hidden",
	}

	out, err = JSONPath(".Bicycle.Color", structIn)
	require.NoError(t, err)
	assert.Equal(t, "red", out)

	_, err = JSONPath(".safe", structIn)
	require.Error(t, err)

	_, err = JSONPath(".*", structIn)
	require.Error(t, err)
}

func TestMissingKeyFromContext(t *testing.T) {
	assert.Equal(t, "error", MissingKeyFromContext(context.Background()))
	assert.Equal(t, "error", MissingKeyFromContext(context.TODO()))
	assert.Equal(t, "default", MissingKeyFromContext(ContextWithMissingKey(context.Background(), "default")))
	assert.Equal(t, "zero", MissingKeyFromContext(ContextWithMissingKey(context.Background(), "zero")))
	assert.Equal(t, "invalid", MissingKeyFromContext(ContextWithMissingKey(context.Background(), "invalid")))
	assert.Equal(t, "error", MissingKeyFromContext(ContextWithMissingKey(context.Background(), "error")))
	assert.Equal(t, "error", MissingKeyFromContext(ContextWithMissingKey(context.Background(), "")))
}

func TestJSONPathWithContext(t *testing.T) {
	in := m{
		"store": m{
			"book": ar{
				m{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				m{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
			},
			"bicycle": m{
				"color": "red",
				"price": 19.95,
			},
		},
	}

	type bicycleType struct {
		Color string
	}
	type storeType struct {
		Bicycle *bicycleType
		safe    any
	}
	structIn := &storeType{
		Bicycle: &bicycleType{
			Color: "red",
		},
		safe: "hidden",
	}

	testCases := []struct {
		name       string
		missingKey string
	}{
		{"missing-key=default", "default"},
		{"missing-key=zero", "zero"},
		{"missing-key=invalid", "invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ContextWithMissingKey(context.Background(), tc.missingKey)

			out, err := JSONPathWithContext(ctx, ".store.bicycle.color", in)
			require.NoError(t, err)
			assert.Equal(t, "red", out)

			out, err = JSONPathWithContext(ctx, ".Bicycle.Color", structIn)
			require.NoError(t, err)
			assert.Equal(t, "red", out)

			out, err = JSONPathWithContext(ctx, ".bogus", in)
			require.NoError(t, err)
			assert.Nil(t, out)

			out, err = JSONPathWithContext(ctx, ".store.bogus", in)
			require.NoError(t, err)
			assert.Nil(t, out)

			out, err = JSONPathWithContext(ctx, ".store.bogus.nested.field", in)
			require.NoError(t, err)
			assert.Nil(t, out)

			out, err = JSONPathWithContext(ctx, ".store.book[99]", in)
			require.NoError(t, err)
			assert.Nil(t, out)

			out, err = JSONPathWithContext(ctx, ".store.book[99].title", in)
			require.NoError(t, err)
			assert.Nil(t, out)

			out, err = JSONPathWithContext(ctx, ".store.bogus[0]", in)
			require.NoError(t, err)
			assert.Nil(t, out)

			out, err = JSONPathWithContext(ctx, ".Bicycle.NonExistent", structIn)
			require.NoError(t, err)
			assert.Nil(t, out)

			_, err = JSONPathWithContext(ctx, ".safe", structIn)
			require.Error(t, err)

			_, err = JSONPathWithContext(ctx, "{.store.unclosed", in)
			require.Error(t, err)
		})
	}

	t.Run("missing-key=error", func(t *testing.T) {
		ctx := ContextWithMissingKey(context.Background(), "error")

		out, err := JSONPathWithContext(ctx, ".store.bicycle.color", in)
		require.NoError(t, err)
		assert.Equal(t, "red", out)

		_, err = JSONPathWithContext(ctx, ".store.bogus", in)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bogus is not found")

		_, err = JSONPathWithContext(ctx, ".store.bogus.nested", in)
		require.Error(t, err)

		_, err = JSONPathWithContext(ctx, ".store.book[99]", in)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "array index out of bounds")

		_, err = JSONPathWithContext(ctx, ".Bicycle.NonExistent", structIn)
		require.Error(t, err)
	})

	t.Run("default context acts as error", func(t *testing.T) {
		ctx := context.Background()

		out, err := JSONPathWithContext(ctx, ".store.bicycle.color", in)
		require.NoError(t, err)
		assert.Equal(t, "red", out)

		_, err = JSONPathWithContext(ctx, ".store.bogus", in)
		require.Error(t, err)

		_, err = JSONPathWithContext(ctx, ".store.book[99]", in)
		require.Error(t, err)
	})
}
