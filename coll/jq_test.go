package coll

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Equal(t, "red", out)

	out, err = JQ(ctx, ".store.bicycle.price", in)
	require.NoError(t, err)
	assert.InEpsilon(t, 19.95, out, 1e-12)

	out, err = JQ(ctx, ".store.bogus", in)
	require.NoError(t, err)
	assert.Nil(t, out)

	_, err = JQ(ctx, "{.store.unclosed", in)
	require.Error(t, err)

	out, err = JQ(ctx, ".store", in)
	require.NoError(t, err)
	assert.EqualValues(t, in["store"], out)

	out, err = JQ(ctx, ".store.book[].author", in)
	require.NoError(t, err)
	assert.Len(t, out, 4)
	assert.Contains(t, out, "Nigel Rees")
	assert.Contains(t, out, "Evelyn Waugh")
	assert.Contains(t, out, "Herman Melville")
	assert.Contains(t, out, "J. R. R. Tolkien")

	out, err = JQ(ctx, ".store.book[]|select(.price < 10.0 )", in)
	require.NoError(t, err)
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
	require.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Contains(t, out, map[string]interface{}{"aaaa": map[string]interface{}{"bar": 1234}})
	assert.Contains(t, out, true)
	assert.Contains(t, out, "baz")
}

func TestJQ_typeConversions(t *testing.T) {
	ctx := context.Background()

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

	out, err := JQ(ctx, ".Bicycle.Color", structIn)
	require.NoError(t, err)
	assert.Equal(t, "red", out)

	out, err = JQ(ctx, ".safe", structIn)
	require.NoError(t, err)
	assert.Nil(t, out)

	_, err = JQ(ctx, ".*", structIn)
	require.Error(t, err)

	// a type with an underlying type of map[string]interface{}, just like
	// gomplate.tmplctx
	type mapType map[string]interface{}

	out, err = JQ(ctx, ".foo", mapType{"foo": "bar"})
	require.NoError(t, err)
	assert.Equal(t, "bar", out)

	// sometimes it'll be a pointer...
	out, err = JQ(ctx, ".foo", &mapType{"foo": "bar"})
	require.NoError(t, err)
	assert.Equal(t, "bar", out)

	// underlying slice type
	type sliceType []interface{}

	out, err = JQ(ctx, ".[1]", sliceType{"foo", "bar"})
	require.NoError(t, err)
	assert.Equal(t, "bar", out)

	out, err = JQ(ctx, ".[2]", &sliceType{"foo", "bar", "baz"})
	require.NoError(t, err)
	assert.Equal(t, "baz", out)

	// other basic types
	out, err = JQ(ctx, ".", []byte("hello"))
	require.NoError(t, err)
	assert.EqualValues(t, "hello", out)

	out, err = JQ(ctx, ".", "hello")
	require.NoError(t, err)
	assert.EqualValues(t, "hello", out)

	out, err = JQ(ctx, ".", 1234)
	require.NoError(t, err)
	assert.EqualValues(t, 1234, out)

	out, err = JQ(ctx, ".", true)
	require.NoError(t, err)
	assert.EqualValues(t, true, out)

	out, err = JQ(ctx, ".", nil)
	require.NoError(t, err)
	assert.Nil(t, out)

	// underlying basic types
	type intType int
	out, err = JQ(ctx, ".", intType(1234))
	require.NoError(t, err)
	assert.EqualValues(t, 1234, out)

	type byteArrayType []byte
	out, err = JQ(ctx, ".", byteArrayType("hello"))
	require.NoError(t, err)
	assert.EqualValues(t, "hello", out)
}

func TestJQConvertType_passthroughTypes(t *testing.T) {
	// non-marshalable values, like recursive structs, can't be used
	type recursive struct{ Self *recursive }
	v := &recursive{}
	v.Self = v
	_, err := jqConvertType(v)
	require.Error(t, err)

	testdata := []interface{}{
		map[string]interface{}{"foo": 1234},
		[]interface{}{"foo", "bar", "baz", 1, 2, 3},
		"foo",
		[]byte("foo"),
		json.RawMessage(`{"foo": "bar"}`),
		true,
		nil,
		int(1234), int8(123), int16(123), int32(123), int64(123),
		uint(123), uint8(123), uint16(123), uint32(123), uint64(123),
		float32(123.45), float64(123.45),
	}

	for _, d := range testdata {
		out, err := jqConvertType(d)
		require.NoError(t, err)
		assert.Equal(t, d, out)
	}
}
