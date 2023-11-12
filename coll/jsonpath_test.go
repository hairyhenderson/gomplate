package coll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	m  = map[string]interface{}
	ar = []interface{}
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
	assert.NoError(t, err)
	assert.Equal(t, "red", out)

	out, err = JSONPath(".store.bicycle.price", in)
	assert.NoError(t, err)
	assert.Equal(t, 19.95, out)

	_, err = JSONPath(".store.bogus", in)
	assert.Error(t, err)

	_, err = JSONPath("{.store.unclosed", in)
	assert.Error(t, err)

	out, err = JSONPath(".store", in)
	assert.NoError(t, err)
	assert.EqualValues(t, in["store"], out)

	out, err = JSONPath("$.store.book[*].author", in)
	assert.NoError(t, err)
	assert.Len(t, out, 4)
	assert.Contains(t, out, "Nigel Rees")
	assert.Contains(t, out, "Evelyn Waugh")
	assert.Contains(t, out, "Herman Melville")
	assert.Contains(t, out, "J. R. R. Tolkien")

	out, err = JSONPath("$..book[?( @.price < 10.0 )]", in)
	assert.NoError(t, err)
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
	assert.EqualValues(t, expected, out)

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
	assert.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Contains(t, out, m{"aaaa": m{"bar": 1234}})
	assert.Contains(t, out, true)
	assert.Contains(t, out, "baz")

	type bicycleType struct {
		Color string
	}
	type storeType struct {
		Bicycle *bicycleType
		safe    interface{}
	}

	structIn := &storeType{
		Bicycle: &bicycleType{
			Color: "red",
		},
		safe: "hidden",
	}

	out, err = JSONPath(".Bicycle.Color", structIn)
	assert.NoError(t, err)
	assert.Equal(t, "red", out)

	_, err = JSONPath(".safe", structIn)
	assert.Error(t, err)

	_, err = JSONPath(".*", structIn)
	assert.Error(t, err)
}
