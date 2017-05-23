package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	ty := &TypeConv{}
	assert.False(t, ty.Bool(""))
	assert.False(t, ty.Bool("asdf"))
	assert.False(t, ty.Bool("1234"))
	assert.False(t, ty.Bool("False"))
	assert.False(t, ty.Bool("0"))
	assert.False(t, ty.Bool("false"))
	assert.False(t, ty.Bool("F"))
	assert.False(t, ty.Bool("f"))
	assert.True(t, ty.Bool("true"))
	assert.True(t, ty.Bool("True"))
	assert.True(t, ty.Bool("t"))
	assert.True(t, ty.Bool("T"))
	assert.True(t, ty.Bool("1"))
}

func TestUnmarshalObj(t *testing.T) {
	ty := new(TypeConv)
	expected := map[string]interface{}{
		"foo":  map[interface{}]interface{}{"bar": "baz"},
		"one":  1.0,
		"true": true,
	}

	test := func(actual map[string]interface{}) {
		assert.Equal(t, expected["foo"], actual["foo"])
		assert.Equal(t, expected["one"], actual["one"])
		assert.Equal(t, expected["true"], actual["true"])
	}
	test(ty.JSON(`{"foo":{"bar":"baz"},"one":1.0,"true":true}`))
	test(ty.YAML(`foo:
  bar: baz
one: 1.0
true: true
`))
}

func TestUnmarshalArray(t *testing.T) {
	ty := new(TypeConv)

	expected := []string{"foo", "bar"}

	test := func(actual []interface{}) {
		assert.Equal(t, expected[0], actual[0])
		assert.Equal(t, expected[1], actual[1])
	}
	test(ty.JSONArray(`["foo","bar"]`))
	test(ty.YAMLArray(`
- foo
- bar
`))
}

func TestToJSON(t *testing.T) {
	ty := new(TypeConv)
	expected := `{"down":{"the":{"rabbit":{"hole":true}}},"foo":"bar","one":1,"true":true}`
	in := map[string]interface{}{
		"foo":  "bar",
		"one":  1,
		"true": true,
		"down": map[interface{}]interface{}{
			"the": map[interface{}]interface{}{
				"rabbit": map[interface{}]interface{}{
					"hole": true,
				},
			},
		},
	}
	assert.Equal(t, expected, ty.ToJSON(in))
}

func TestToJSONPretty(t *testing.T) {
	ty := new(TypeConv)
	expected := `{
  "down": {
    "the": {
      "rabbit": {
        "hole": true
      }
    }
  },
  "foo": "bar",
  "one": 1,
  "true": true
}`
	in := map[string]interface{}{
		"foo":  "bar",
		"one":  1,
		"true": true,
		"down": map[string]interface{}{
			"the": map[string]interface{}{
				"rabbit": map[string]interface{}{
					"hole": true,
				},
			},
		},
	}
	assert.Equal(t, expected, ty.toJSONPretty("  ", in))
}

func TestToYAML(t *testing.T) {
	ty := new(TypeConv)
	expected := `d: 2006-01-02T15:04:05.999999999-07:00
foo: bar
? |-
  multi
  line
  key
: hello: world
one: 1
"true": true
`
	mst, _ := time.LoadLocation("MST")
	in := map[string]interface{}{
		"foo":  "bar",
		"one":  1,
		"true": true,
		`multi
line
key`: map[string]interface{}{
			"hello": "world",
		},
		"d": time.Date(2006, time.January, 2, 15, 4, 5, 999999999, mst),
	}
	assert.Equal(t, expected, ty.ToYAML(in))
}

func TestSlice(t *testing.T) {
	ty := new(TypeConv)
	expected := []string{"foo", "bar"}
	actual := ty.Slice("foo", "bar")
	assert.Equal(t, expected[0], actual[0])
	assert.Equal(t, expected[1], actual[1])
}

func TestJoin(t *testing.T) {
	ty := new(TypeConv)

	assert.Equal(t, "foo,bar", ty.Join([]interface{}{"foo", "bar"}, ","))
	assert.Equal(t, "foo,\nbar", ty.Join([]interface{}{"foo", "bar"}, ",\n"))
	// Join handles all kinds of scalar types too...
	assert.Equal(t, "42-18446744073709551615", ty.Join([]interface{}{42, uint64(18446744073709551615)}, "-"))
	assert.Equal(t, "1,,true,3.14,foo,nil", ty.Join([]interface{}{1, "", true, 3.14, "foo", nil}, ","))
	// and best-effort with weird types
	assert.Equal(t, "[foo],bar", ty.Join([]interface{}{[]string{"foo"}, "bar"}, ","))
}

func TestHas(t *testing.T) {
	ty := new(TypeConv)

	in := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]interface{}{
			"qux": "quux",
		},
	}

	assert.True(t, ty.Has(in, "foo"))
	assert.False(t, ty.Has(in, "bar"))
	assert.True(t, ty.Has(in["baz"], "qux"))
}

func TestIndent(t *testing.T) {
	ty := new(TypeConv)
	actual := "hello\nworld\n!"
	expected := "  hello\n  world\n  !"
	assert.Equal(t, expected, ty.indent("  ", actual))

	assert.Equal(t, "\n", ty.indent("  ", "\n"))

	assert.Equal(t, "  foo\n", ty.indent("  ", "foo\n"))

	assert.Equal(t, "   foo", ty.indent("   ", "foo"))
}

func TestCSV(t *testing.T) {
	ty := new(TypeConv)
	in := "first,second,third\n1,2,3\n4,5,6"
	expected := [][]string{
		{"first", "second", "third"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	assert.Equal(t, expected, ty.CSV(in))

	in = "first;second;third\r\n1;2;3\r\n4;5;6\r\n"
	assert.Equal(t, expected, ty.CSV(";", in))
}

func TestCSVByRow(t *testing.T) {
	ty := new(TypeConv)
	in := "first,second,third\n1,2,3\n4,5,6"
	expected := []map[string]string{
		{
			"first":  "1",
			"second": "2",
			"third":  "3",
		},
		{
			"first":  "4",
			"second": "5",
			"third":  "6",
		},
	}
	assert.Equal(t, expected, ty.CSVByRow(in))

	in = "1,2,3\n4,5,6"
	assert.Equal(t, expected, ty.CSVByRow("first,second,third", in))

	in = "1;2;3\n4;5;6"
	assert.Equal(t, expected, ty.CSVByRow(";", "first;second;third", in))

	in = "first;second;third\r\n1;2;3\r\n4;5;6"
	assert.Equal(t, expected, ty.CSVByRow(";", in))

	expected = []map[string]string{
		{"A": "1", "B": "2", "C": "3"},
		{"A": "4", "B": "5", "C": "6"},
	}

	in = "1,2,3\n4,5,6"
	assert.Equal(t, expected, ty.CSVByRow("", in))

	expected = []map[string]string{
		{"A": "1", "B": "1", "C": "1", "D": "1", "E": "1", "F": "1", "G": "1", "H": "1", "I": "1", "J": "1", "K": "1", "L": "1", "M": "1", "N": "1", "O": "1", "P": "1", "Q": "1", "R": "1", "S": "1", "T": "1", "U": "1", "V": "1", "W": "1", "X": "1", "Y": "1", "Z": "1", "AA": "1", "BB": "1", "CC": "1", "DD": "1"},
	}

	in = "1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1"
	assert.Equal(t, expected, ty.CSVByRow("", in))
}

func TestCSVByColumn(t *testing.T) {
	ty := new(TypeConv)
	in := "first,second,third\n1,2,3\n4,5,6"
	expected := map[string][]string{
		"first":  {"1", "4"},
		"second": {"2", "5"},
		"third":  {"3", "6"},
	}
	assert.Equal(t, expected, ty.CSVByColumn(in))

	in = "1,2,3\n4,5,6"
	assert.Equal(t, expected, ty.CSVByColumn("first,second,third", in))

	in = "1;2;3\n4;5;6"
	assert.Equal(t, expected, ty.CSVByColumn(";", "first;second;third", in))

	in = "first;second;third\r\n1;2;3\r\n4;5;6"
	assert.Equal(t, expected, ty.CSVByColumn(";", in))

	expected = map[string][]string{
		"A": {"1", "4"},
		"B": {"2", "5"},
		"C": {"3", "6"},
	}

	in = "1,2,3\n4,5,6"
	assert.Equal(t, expected, ty.CSVByColumn("", in))
}

func TestAutoIndex(t *testing.T) {
	assert.Equal(t, "A", autoIndex(0))
	assert.Equal(t, "B", autoIndex(1))
	assert.Equal(t, "Z", autoIndex(25))
	assert.Equal(t, "AA", autoIndex(26))
	assert.Equal(t, "ZZ", autoIndex(51))
	assert.Equal(t, "AAA", autoIndex(52))
	assert.Equal(t, "YYYYY", autoIndex(128))
}

func TestToCSV(t *testing.T) {
	ty := new(TypeConv)
	in := [][]string{
		{"first", "second", "third"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	expected := "first,second,third\r\n1,2,3\r\n4,5,6\r\n"

	assert.Equal(t, expected, ty.ToCSV(in))

	expected = "first;second;third\r\n1;2;3\r\n4;5;6\r\n"

	assert.Equal(t, expected, ty.ToCSV(";", in))
}
