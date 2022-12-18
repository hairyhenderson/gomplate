package data

import (
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"os"

	"gotest.tools/v3/fs"
)

func TestUnmarshalObj(t *testing.T) {
	expected := map[string]interface{}{
		"foo":     map[string]interface{}{"bar": "baz"},
		"one":     1.0,
		"true":    true,
		"escaped": "\"/\\\b\f\n\r\t∞",
	}

	test := func(actual interface{}, err error) {
		t.Helper()
		assert.NoError(t, err)
		assert.IsType(t, make(map[string]interface{}), actual)

		actualObj := actual.(map[string]interface{})

		assert.Equal(t, expected["foo"], actualObj["foo"], "foo")
		assert.Equal(t, expected["one"], actualObj["one"], "one")
		assert.Equal(t, expected["true"], actualObj["true"], "true")
		assert.Equal(t, expected["escaped"], actualObj["escaped"], "escaped")
	}
	test(YAML(`{"foo":{"bar":"baz"},"one":1.0,"true":true,"escaped":"\"\/\\\b\f\n\r\t\u221e"}`))
	test(YAML(`foo:
  bar: baz
one: 1.0
'true': true
escaped: "\"\/\\\b\f\n\r\t\u221e"
`))
	test(YAML(`anchor: &anchor
  bar: baz
foo:
  <<: *anchor
one: 1.0
'true': true
escaped: "\"\/\\\b\f\n\r\t\u221e"
`))
	test(YAML(`# this comment marks an empty (nil!) document
---
# this one too, for good measure
---
foo:
  bar: baz
one: 1.0
'true': true
escaped: "\"\/\\\b\f\n\r\t\u221e"
`))
}

func TestUnmarshalArray(t *testing.T) {
	expected := []interface{}{"foo", "bar",
		map[string]interface{}{
			"baz":   map[string]interface{}{"qux": true},
			"quux":  map[string]interface{}{"42": 18},
			"corge": map[string]interface{}{"false": "blah"},
		}}

	test := func(actual interface{}, err error) {
		assert.NoError(t, err)
		assert.EqualValues(t, expected, actual)
	}
	test(YAML(`["foo","bar",{"baz":{"qux": true},"quux":{"42":18},"corge":{"false":"blah"}}]`))
	test(YAML(`
- foo
- bar
- baz:
    qux: true
  quux:
    "42": 18
  corge:
    "false": blah
`))
	test(YAML(`---
# blah blah blah ignore this!
---
- foo
- bar
- baz:
    qux: true
  quux:
    "42": 18
  corge:
    "false": blah
---
this shouldn't be reached
`))

	actual, err := YAML(`---
- foo: &foo
    bar: baz
- qux:
    <<: *foo
    quux: corge
- baz:
    qux: true
    42: 18
    false: blah
`)
	assert.NoError(t, err)
	assert.EqualValues(t,
		[]interface{}{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			map[string]interface{}{
				"qux": map[string]interface{}{
					"bar":  "baz",
					"quux": "corge",
				},
			},
			map[string]interface{}{
				"baz": map[string]interface{}{
					"qux":   true,
					"42":    18,
					"false": "blah",
				},
			},
		},
		actual)
}

func TestMarshalObj(t *testing.T) {
	expected := "foo"
	actual, err := marshalObj(nil, func(in interface{}) ([]byte, error) {
		return []byte("foo"), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
	_, err = marshalObj(nil, func(in interface{}) ([]byte, error) {
		return nil, errors.New("fail")
	})
	assert.Error(t, err)
}

func TestToJSONBytes(t *testing.T) {
	expected := []byte("null")
	actual, err := toJSONBytes(nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	_, err = toJSONBytes(&badObject{})
	assert.Error(t, err)
}

type badObject struct {
}

func (b *badObject) CodecEncodeSelf(e *codec.Encoder) {
	panic("boom")
}

func (b *badObject) CodecDecodeSelf(e *codec.Decoder) {

}

func TestToJSON(t *testing.T) {
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
	out, err := ToJSON(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)

	_, err = ToJSON(&badObject{})
	assert.Error(t, err)
}

func TestToJSONPretty(t *testing.T) {
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
	out, err := ToJSONPretty("  ", in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)

	_, err = ToJSONPretty("  ", &badObject{})
	assert.Error(t, err)
}

func TestToYAML(t *testing.T) {
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
	out, err := ToYAML(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}

func TestCSV(t *testing.T) {
	expected := [][]string{
		{"first", "second", "third"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	testdata := []struct {
		args []string
		out  [][]string
	}{
		{[]string{"first,second,third\n1,2,3\n4,5,6"}, expected},
		{[]string{";", "first;second;third\r\n1;2;3\r\n4;5;6\r\n"}, expected},

		{[]string{""}, [][]string{nil}},
		{[]string{"\n"}, [][]string{nil}},
		{[]string{"foo"}, [][]string{{"foo"}}},
	}
	for _, d := range testdata {
		out, err := CSV(d.args...)
		assert.NoError(t, err)
		assert.Equal(t, d.out, out)
	}
}

func TestCSVByRow(t *testing.T) {
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
	testdata := []struct {
		args []string
		out  []map[string]string
	}{
		{[]string{in}, expected},
		{[]string{"first,second,third", "1,2,3\n4,5,6"}, expected},
		{[]string{";", "first;second;third", "1;2;3\n4;5;6"}, expected},
		{[]string{";", "first;second;third\r\n1;2;3\r\n4;5;6"}, expected},
		{[]string{"", "1,2,3\n4,5,6"}, []map[string]string{
			{"A": "1", "B": "2", "C": "3"},
			{"A": "4", "B": "5", "C": "6"},
		}},
		{[]string{"", "1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1"}, []map[string]string{
			{"A": "1", "B": "1", "C": "1", "D": "1", "E": "1", "F": "1", "G": "1", "H": "1", "I": "1", "J": "1", "K": "1", "L": "1", "M": "1", "N": "1", "O": "1", "P": "1", "Q": "1", "R": "1", "S": "1", "T": "1", "U": "1", "V": "1", "W": "1", "X": "1", "Y": "1", "Z": "1", "AA": "1", "BB": "1", "CC": "1", "DD": "1"},
		}},
	}
	for _, d := range testdata {
		out, err := CSVByRow(d.args...)
		assert.NoError(t, err)
		assert.Equal(t, d.out, out)
	}
}

func TestCSVByColumn(t *testing.T) {
	expected := map[string][]string{
		"first":  {"1", "4"},
		"second": {"2", "5"},
		"third":  {"3", "6"},
	}

	testdata := []struct {
		out  map[string][]string
		args []string
	}{
		{expected, []string{"first,second,third\n1,2,3\n4,5,6"}},
		{expected, []string{"first,second,third", "1,2,3\n4,5,6"}},
		{expected, []string{";", "first;second;third", "1;2;3\n4;5;6"}},
		{expected, []string{";", "first;second;third\r\n1;2;3\r\n4;5;6"}},
		{map[string][]string{
			"A": {"1", "4"},
			"B": {"2", "5"},
			"C": {"3", "6"},
		}, []string{"", "1,2,3\n4,5,6"}},
	}
	for _, d := range testdata {
		out, err := CSVByColumn(d.args...)
		assert.NoError(t, err)
		assert.Equal(t, d.out, out)
	}
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
	in := [][]string{
		{"first", "second", "third"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	expected := "first,second,third\r\n1,2,3\r\n4,5,6\r\n"

	out, err := ToCSV(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)

	expected = "first;second;third\r\n1;2;3\r\n4;5;6\r\n"

	out, err = ToCSV(";", in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)

	_, err = ToCSV(42, [][]int{{1, 2}})
	assert.Error(t, err)

	_, err = ToCSV([][]int{{1, 2}})
	assert.Error(t, err)

	expected = "first,second,third\r\n1,2,3\r\n4,5,6\r\n"
	out, err = ToCSV([][]interface{}{
		{"first", "second", "third"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, out)

	expected = "first|second|third\r\n1|2|3\r\n4|5|6\r\n"
	out, err = ToCSV("|", []interface{}{
		[]interface{}{"first", "second", "third"},
		[]interface{}{1, "2", 3},
		[]interface{}{"4", 5, "6"},
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}

func TestTOML(t *testing.T) {
	in := `# This is a TOML document. Boom.

title = "TOML Example"

[owner]
name = "Tom Preston-Werner"
organization = "GitHub"
bio = "GitHub Cofounder & CEO\nLikes tater tots and beer."
dob = 1979-05-27T07:32:00Z # First class dates? Why not?

[database]
server = "192.168.1.1"
ports = [ 8001, 8001, 8002 ]
connection_max = 5000
enabled = true

[servers]

  # You can indent as you please. Tabs or spaces. TOML don't care.
  [servers.alpha]
  ip = "10.0.0.1"
  dc = "eqdc10"

  [servers.beta]
  ip = "10.0.0.2"
  dc = "eqdc10"

[clients]
data = [ ["gamma", "delta"], [1, 2] ] # just an update to make sure parsers support it

# Line breaks are OK when inside arrays
hosts = [
  "alpha",
  "omega"
]
`
	expected := map[string]interface{}{
		"title": "TOML Example",
		"owner": map[string]interface{}{
			"name":         "Tom Preston-Werner",
			"organization": "GitHub",
			"bio":          "GitHub Cofounder & CEO\nLikes tater tots and beer.",
			"dob":          time.Date(1979, time.May, 27, 7, 32, 0, 0, time.UTC),
		},
		"database": map[string]interface{}{
			"server":         "192.168.1.1",
			"ports":          []interface{}{int64(8001), int64(8001), int64(8002)},
			"connection_max": int64(5000),
			"enabled":        true,
		},
		"servers": map[string]interface{}{
			"alpha": map[string]interface{}{
				"ip": "10.0.0.1",
				"dc": "eqdc10",
			},
			"beta": map[string]interface{}{
				"ip": "10.0.0.2",
				"dc": "eqdc10",
			},
		},
		"clients": map[string]interface{}{
			"data": []interface{}{
				[]interface{}{"gamma", "delta"},
				[]interface{}{int64(1), int64(2)},
			},
			"hosts": []interface{}{"alpha", "omega"},
		},
	}

	out, err := TOML(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}

func TestToTOML(t *testing.T) {
	expected := `foo = "bar"
one = 1
true = true

[down]
  [down.the]
    [down.the.rabbit]
      hole = true
`
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
	out, err := ToTOML(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}

func TestDecryptEJSON(t *testing.T) {
	privateKey := "e282d979654f88267f7e6c2d8268f1f4314b8673579205ed0029b76de9c8223f"
	publicKey := "6e05ec625bcdca34864181cc43e6fcc20a57732a453bc2f4a2e117ffdf1a6762"
	expected := map[string]interface{}{
		"password":     "supersecret",
		"_unencrypted": "notsosecret",
	}
	in := `{
		"_public_key": "` + publicKey + `",
		"password": "EJ[1:yJ7n4UorqxkJZMoKevIA1dJeDvaQhkbgENIVZW18jig=:0591iW+paVSh4APOytKBVW/ZcxHO/5wO:TssnpVtkiXmpDIxPlXSiYdgnWyd44stGcwG1]",
		"_unencrypted": "notsosecret"
	}`

	os.Setenv("EJSON_KEY", privateKey)
	defer os.Unsetenv("EJSON_KEY")
	actual, err := decryptEJSON(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)

	actualJson, err := YAML(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actualJson)

	tmpDir := fs.NewDir(t, "gomplate-ejsontest",
		fs.WithFile(publicKey, privateKey),
	)
	defer tmpDir.Remove()

	os.Unsetenv("EJSON_KEY")
	os.Setenv("EJSON_KEY_FILE", tmpDir.Join(publicKey))
	defer os.Unsetenv("EJSON_KEY_FILE")
	actual, err = decryptEJSON(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)

	os.Unsetenv("EJSON_KEY")
	os.Unsetenv("EJSON_KEY_FILE")
	os.Setenv("EJSON_KEYDIR", tmpDir.Path())
	defer os.Unsetenv("EJSON_KEYDIR")
	actual, err = decryptEJSON(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestDotEnv(t *testing.T) {
	in := `FOO=a regular unquoted value
export BAR=another value, exports are ignored

# comments are totally ignored, as are blank lines
FOO.BAR = "values can be double-quoted, and shell\nescapes are supported"

BAZ = "variable expansion: ${FOO}"
QUX='single quotes ignore $variables'
`
	expected := map[string]interface{}{
		"FOO":     "a regular unquoted value",
		"BAR":     "another value, exports are ignored",
		"FOO.BAR": "values can be double-quoted, and shell\nescapes are supported",
		"BAZ":     "variable expansion: a regular unquoted value",
		"QUX":     "single quotes ignore $variables",
	}
	out, err := dotEnv(in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)
}

func TestStringifyYAMLArrayMapKeys(t *testing.T) {
	cases := []struct {
		input    []interface{}
		want     []interface{}
		replaced bool
	}{
		{
			[]interface{}{map[interface{}]interface{}{"a": 1, "b": 2}},
			[]interface{}{map[string]interface{}{"a": 1, "b": 2}},
			false,
		},
		{
			[]interface{}{map[interface{}]interface{}{"a": []interface{}{1, map[interface{}]interface{}{"b": 2}}}},
			[]interface{}{map[string]interface{}{"a": []interface{}{1, map[string]interface{}{"b": 2}}}},
			false,
		},
		{
			[]interface{}{map[interface{}]interface{}{true: 1, "b": false}},
			[]interface{}{map[string]interface{}{"true": 1, "b": false}},
			false,
		},
		{
			[]interface{}{map[interface{}]interface{}{1: "a", 2: "b"}},
			[]interface{}{map[string]interface{}{"1": "a", "2": "b"}},
			false,
		},
		{
			[]interface{}{map[interface{}]interface{}{"a": map[interface{}]interface{}{"b": 1}}},
			[]interface{}{map[string]interface{}{"a": map[string]interface{}{"b": 1}}},
			false,
		},
		{
			[]interface{}{map[string]interface{}{"a": map[string]interface{}{"b": 1}}},
			[]interface{}{map[string]interface{}{"a": map[string]interface{}{"b": 1}}},
			false,
		},
		{
			[]interface{}{map[interface{}]interface{}{1: "a", 2: "b"}},
			[]interface{}{map[string]interface{}{"1": "a", "2": "b"}},
			false,
		},
	}

	for _, c := range cases {
		actual, _ := stringifyMapKeys(c.input)
		assert.EqualValues(t, c.want, actual)
	}
}

func TestStringifyYAMLMapMapKeys(t *testing.T) {
	cases := []struct {
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			map[string]interface{}{"root": map[interface{}]interface{}{"a": 1, "b": 2}},
			map[string]interface{}{"root": map[string]interface{}{"a": 1, "b": 2}},
		},
		{
			map[string]interface{}{"root": map[interface{}]interface{}{"a": []interface{}{1, map[interface{}]interface{}{"b": 2}}}},
			map[string]interface{}{"root": map[string]interface{}{"a": []interface{}{1, map[string]interface{}{"b": 2}}}},
		},
		{
			map[string]interface{}{"root": map[interface{}]interface{}{true: 1, "b": false}},
			map[string]interface{}{"root": map[string]interface{}{"true": 1, "b": false}},
		},
		{
			map[string]interface{}{"root": map[interface{}]interface{}{1: "a", 2: "b"}},
			map[string]interface{}{"root": map[string]interface{}{"1": "a", "2": "b"}},
		},
		{
			map[string]interface{}{"root": map[interface{}]interface{}{"a": map[interface{}]interface{}{"b": 1}}},
			map[string]interface{}{"root": map[string]interface{}{"a": map[string]interface{}{"b": 1}}},
		},
		{
			map[string]interface{}{"a": map[string]interface{}{"b": 1}},
			map[string]interface{}{"a": map[string]interface{}{"b": 1}},
		},
		{
			map[string]interface{}{"root": []interface{}{map[interface{}]interface{}{1: "a", 2: "b"}}},
			map[string]interface{}{"root": []interface{}{map[string]interface{}{"1": "a", "2": "b"}}},
		},
	}

	for _, c := range cases {
		actual, _ := stringifyMapKeys(c.input)
		assert.EqualValues(t, c.want, actual)
	}
}
