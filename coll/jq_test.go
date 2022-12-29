package coll

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJQ(t *testing.T) {
	ctx := context.Background()
	in := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				map[string]interface{}{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				map[string]interface{}{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				map[string]interface{}{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": map[string]interface{}{
				"color": "red",
				"price": 19.95,
			},
		},
	}
	out, err := JQ(ctx, ".store.bicycle.color", in)
	assert.NoError(t, err)
	assert.Equal(t, "red", out)

	out, err = JQ(ctx, ".store.bicycle.price", in)
	assert.NoError(t, err)
	assert.Equal(t, 19.95, out)

	out, err = JQ(ctx, ".store.bogus", in)
	assert.NoError(t, err)
	assert.Nil(t, out)

	_, err = JQ(ctx, "{.store.unclosed", in)
	assert.Error(t, err)

	out, err = JQ(ctx, ".store", in)
	assert.NoError(t, err)
	assert.EqualValues(t, in["store"], out)

	out, err = JQ(ctx, ".store.book[].author", in)
	assert.NoError(t, err)
	assert.Len(t, out, 4)
	assert.Contains(t, out, "Nigel Rees")
	assert.Contains(t, out, "Evelyn Waugh")
	assert.Contains(t, out, "Herman Melville")
	assert.Contains(t, out, "J. R. R. Tolkien")

	out, err = JQ(ctx, ".store.book[]|select(.price < 10.0 )", in)
	assert.NoError(t, err)
	expected := []interface{}{
		map[string]interface{}{
			"category": "reference",
			"author":   "Nigel Rees",
			"title":    "Sayings of the Century",
			"price":    8.95,
		},
		map[string]interface{}{
			"category": "fiction",
			"author":   "Herman Melville",
			"title":    "Moby Dick",
			"isbn":     "0-553-21311-3",
			"price":    8.99,
		},
	}
	assert.EqualValues(t, expected, out)

	in = map[string]interface{}{
		"a": map[string]interface{}{
			"aa": map[string]interface{}{
				"foo": map[string]interface{}{
					"aaa": map[string]interface{}{
						"aaaa": map[string]interface{}{
							"bar": 1234,
						},
					},
				},
			},
			"ab": map[string]interface{}{
				"aba": map[string]interface{}{
					"foo": map[string]interface{}{
						"abaa": true,
						"abab": "baz",
					},
				},
			},
		},
	}
	out, err = JQ(ctx, `tostream|select((.[0]|index("foo")) and (.[0][-1]!="foo") and (.[1])) as $s|($s[0]|index("foo")+1) as $ind|($ind|truncate_stream($s)) as $newstream|$newstream|reduce . as [$p,$v] ({};setpath($p;$v))|add`, in)
	assert.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Contains(t, out, map[string]interface{}{"aaaa": map[string]interface{}{"bar": 1234}})
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

	// TODO: Check if this is a valid test case (taken from jsonpath_test.go) since the struct
	// had to be converted to JSON and parsed from it again to be able to process using gojq.
	v := map[string]interface{}{}
	b, err := json.Marshal(structIn)
	assert.NoError(t, err)
	err = json.Unmarshal(b, &v)
	assert.NoError(t, err)
	out, err = JQ(ctx, ".Bicycle.Color", v)
	assert.NoError(t, err)
	assert.Equal(t, "red", out)

	_, err = JQ(ctx, ".safe", structIn)
	assert.Error(t, err)

	_, err = JQ(ctx, ".*", structIn)
	assert.Error(t, err)
}
