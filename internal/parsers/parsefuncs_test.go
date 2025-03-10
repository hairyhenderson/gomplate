package parsers

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gotest.tools/v3/fs"
)

func TestUnmarshalObj(t *testing.T) {
	expected := map[string]any{
		"foo":     map[string]any{"bar": "baz"},
		"one":     1.0,
		"true":    true,
		"escaped": "\"/\\\b\f\n\r\t∞",
	}

	test := func(actual map[string]any, err error) {
		t.Helper()
		require.NoError(t, err)
		assert.Equal(t, expected["foo"], actual["foo"], "foo")
		assert.Equal(t, expected["one"], actual["one"], "one")
		assert.Equal(t, expected["true"], actual["true"], "true")
		assert.Equal(t, expected["escaped"], actual["escaped"], "escaped")
	}
	test(JSON(`{"foo":{"bar":"baz"},"one":1.0,"true":true,"escaped":"\"\/\\\b\f\n\r\t\u221e"}`))
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

	obj := make(map[string]any)
	_, err := unmarshalObj(obj, "SOMETHING", func(_ []byte, _ any) error {
		return fmt.Errorf("fail")
	})
	assert.EqualError(t, err, "unable to unmarshal object SOMETHING: fail")
}

func TestUnmarshalArray(t *testing.T) {
	expected := []any{
		"foo", "bar",
		map[string]any{
			"baz":   map[string]any{"qux": true},
			"quux":  map[string]any{"42": 18},
			"corge": map[string]any{"false": "blah"},
		},
	}

	test := func(actual []any, err error) {
		require.NoError(t, err)
		assert.EqualValues(t, expected, actual)
	}
	test(JSONArray(`["foo","bar",{"baz":{"qux": true},"quux":{"42":18},"corge":{"false":"blah"}}]`))
	test(YAMLArray(`
- foo
- bar
- baz:
    qux: true
  quux:
    "42": 18
  corge:
    "false": blah
`))
	test(YAMLArray(`---
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

	actual, err := YAMLArray(`---
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
	require.NoError(t, err)
	assert.EqualValues(t,
		[]any{
			map[string]any{
				"foo": map[string]any{
					"bar": "baz",
				},
			},
			map[string]any{
				"qux": map[string]any{
					"bar":  "baz",
					"quux": "corge",
				},
			},
			map[string]any{
				"baz": map[string]any{
					"qux":   true,
					"42":    18,
					"false": "blah",
				},
			},
		},
		actual)

	obj := make([]any, 1)
	_, err = unmarshalArray(obj, "SOMETHING", func(_ []byte, _ any) error {
		return fmt.Errorf("fail")
	})
	assert.EqualError(t, err, "unable to unmarshal array SOMETHING: fail")
}

func TestMarshalObj(t *testing.T) {
	expected := "foo"
	actual, err := marshalObj(nil, func(_ any) ([]byte, error) {
		return []byte("foo"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
	_, err = marshalObj(nil, func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("fail")
	})
	require.Error(t, err)
}

func TestToJSONBytes(t *testing.T) {
	expected := []byte("null")
	actual, err := toJSONBytes(nil)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	_, err = toJSONBytes(&badObject{})
	require.Error(t, err)
}

type badObject struct{}

func (b *badObject) CodecEncodeSelf(_ *codec.Encoder) {
	panic("boom")
}

func (b *badObject) CodecDecodeSelf(_ *codec.Decoder) {
}

func TestToJSON(t *testing.T) {
	expected := `{"down":{"the":{"rabbit":{"hole":true}}},"foo":"bar","one":1,"true":true}`
	in := map[string]any{
		"foo":  "bar",
		"one":  1,
		"true": true,
		"down": map[any]any{
			"the": map[any]any{
				"rabbit": map[any]any{
					"hole": true,
				},
			},
		},
	}
	out, err := ToJSON(in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	_, err = ToJSON(&badObject{})
	require.Error(t, err)
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
	in := map[string]any{
		"foo":  "bar",
		"one":  1,
		"true": true,
		"down": map[string]any{
			"the": map[string]any{
				"rabbit": map[string]any{
					"hole": true,
				},
			},
		},
	}
	out, err := ToJSONPretty("  ", in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	_, err = ToJSONPretty("  ", &badObject{})
	require.Error(t, err)
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
	in := map[string]any{
		"foo":  "bar",
		"one":  1,
		"true": true,
		`multi
line
key`: map[string]any{
			"hello": "world",
		},
		"d": time.Date(2006, time.January, 2, 15, 4, 5, 999999999, mst),
	}
	out, err := ToYAML(in)
	require.NoError(t, err)
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
		require.NoError(t, err)
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
		require.NoError(t, err)
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
		require.NoError(t, err)
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
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	expected = "first;second;third\r\n1;2;3\r\n4;5;6\r\n"

	out, err = ToCSV(";", in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	_, err = ToCSV(42, [][]int{{1, 2}})
	require.Error(t, err)

	_, err = ToCSV([][]int{{1, 2}})
	require.Error(t, err)

	expected = "first,second,third\r\n1,2,3\r\n4,5,6\r\n"
	out, err = ToCSV([][]any{
		{"first", "second", "third"},
		{"1", "2", "3"},
		{"4", "5", "6"},
	})
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	expected = "first|second|third\r\n1|2|3\r\n4|5|6\r\n"
	out, err = ToCSV("|", []any{
		[]any{"first", "second", "third"},
		[]any{1, "2", 3},
		[]any{"4", 5, "6"},
	})
	require.NoError(t, err)
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
	expected := map[string]any{
		"title": "TOML Example",
		"owner": map[string]any{
			"name":         "Tom Preston-Werner",
			"organization": "GitHub",
			"bio":          "GitHub Cofounder & CEO\nLikes tater tots and beer.",
			"dob":          time.Date(1979, time.May, 27, 7, 32, 0, 0, time.UTC),
		},
		"database": map[string]any{
			"server":         "192.168.1.1",
			"ports":          []any{int64(8001), int64(8001), int64(8002)},
			"connection_max": int64(5000),
			"enabled":        true,
		},
		"servers": map[string]any{
			"alpha": map[string]any{
				"ip": "10.0.0.1",
				"dc": "eqdc10",
			},
			"beta": map[string]any{
				"ip": "10.0.0.2",
				"dc": "eqdc10",
			},
		},
		"clients": map[string]any{
			"data": []any{
				[]any{"gamma", "delta"},
				[]any{int64(1), int64(2)},
			},
			"hosts": []any{"alpha", "omega"},
		},
	}

	out, err := TOML(in)
	require.NoError(t, err)
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
	in := map[string]any{
		"foo":  "bar",
		"one":  1,
		"true": true,
		"down": map[any]any{
			"the": map[any]any{
				"rabbit": map[any]any{
					"hole": true,
				},
			},
		},
	}
	out, err := ToTOML(in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)
}

func TestDecryptEJSON(t *testing.T) {
	privateKey := "e282d979654f88267f7e6c2d8268f1f4314b8673579205ed0029b76de9c8223f"
	publicKey := "6e05ec625bcdca34864181cc43e6fcc20a57732a453bc2f4a2e117ffdf1a6762"
	expected := map[string]any{
		"password":     "supersecret",
		"_unencrypted": "notsosecret",
	}
	in := `{
		"_public_key": "` + publicKey + `",
		"password": "EJ[1:yJ7n4UorqxkJZMoKevIA1dJeDvaQhkbgENIVZW18jig=:0591iW+paVSh4APOytKBVW/ZcxHO/5wO:TssnpVtkiXmpDIxPlXSiYdgnWyd44stGcwG1]",
		"_unencrypted": "notsosecret"
	}`

	t.Setenv("EJSON_KEY", privateKey)
	actual, err := decryptEJSON(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, actual)

	actual, err = JSON(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, actual)

	tmpDir := fs.NewDir(t, "gomplate-ejsontest",
		fs.WithFile(publicKey, privateKey),
	)
	t.Cleanup(tmpDir.Remove)

	os.Unsetenv("EJSON_KEY")
	t.Setenv("EJSON_KEY_FILE", tmpDir.Join(publicKey))
	actual, err = decryptEJSON(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, actual)

	os.Unsetenv("EJSON_KEY")
	os.Unsetenv("EJSON_KEY_FILE")
	t.Setenv("EJSON_KEYDIR", tmpDir.Path())
	actual, err = decryptEJSON(in)
	require.NoError(t, err)
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
	expected := map[string]any{
		"FOO":     "a regular unquoted value",
		"BAR":     "another value, exports are ignored",
		"FOO.BAR": "values can be double-quoted, and shell\nescapes are supported",
		"BAZ":     "variable expansion: a regular unquoted value",
		"QUX":     "single quotes ignore $variables",
	}
	out, err := DotEnv(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)
}

func TestStringifyYAMLArrayMapKeys(t *testing.T) {
	cases := []struct {
		input    []any
		want     []any
		replaced bool
	}{
		{
			[]any{map[any]any{"a": 1, "b": 2}},
			[]any{map[string]any{"a": 1, "b": 2}},
			false,
		},
		{
			[]any{map[any]any{"a": []any{1, map[any]any{"b": 2}}}},
			[]any{map[string]any{"a": []any{1, map[string]any{"b": 2}}}},
			false,
		},
		{
			[]any{map[any]any{true: 1, "b": false}},
			[]any{map[string]any{"true": 1, "b": false}},
			false,
		},
		{
			[]any{map[any]any{1: "a", 2: "b"}},
			[]any{map[string]any{"1": "a", "2": "b"}},
			false,
		},
		{
			[]any{map[any]any{"a": map[any]any{"b": 1}}},
			[]any{map[string]any{"a": map[string]any{"b": 1}}},
			false,
		},
		{
			[]any{map[string]any{"a": map[string]any{"b": 1}}},
			[]any{map[string]any{"a": map[string]any{"b": 1}}},
			false,
		},
		{
			[]any{map[any]any{1: "a", 2: "b"}},
			[]any{map[string]any{"1": "a", "2": "b"}},
			false,
		},
	}

	for _, c := range cases {
		err := stringifyYAMLArrayMapKeys(c.input)
		require.NoError(t, err)
		assert.EqualValues(t, c.want, c.input)
	}
}

func TestStringifyYAMLMapMapKeys(t *testing.T) {
	cases := []struct {
		input map[string]any
		want  map[string]any
	}{
		{
			map[string]any{"root": map[any]any{"a": 1, "b": 2}},
			map[string]any{"root": map[string]any{"a": 1, "b": 2}},
		},
		{
			map[string]any{"root": map[any]any{"a": []any{1, map[any]any{"b": 2}}}},
			map[string]any{"root": map[string]any{"a": []any{1, map[string]any{"b": 2}}}},
		},
		{
			map[string]any{"root": map[any]any{true: 1, "b": false}},
			map[string]any{"root": map[string]any{"true": 1, "b": false}},
		},
		{
			map[string]any{"root": map[any]any{1: "a", 2: "b"}},
			map[string]any{"root": map[string]any{"1": "a", "2": "b"}},
		},
		{
			map[string]any{"root": map[any]any{"a": map[any]any{"b": 1}}},
			map[string]any{"root": map[string]any{"a": map[string]any{"b": 1}}},
		},
		{
			map[string]any{"a": map[string]any{"b": 1}},
			map[string]any{"a": map[string]any{"b": 1}},
		},
		{
			map[string]any{"root": []any{map[any]any{1: "a", 2: "b"}}},
			map[string]any{"root": []any{map[string]any{"1": "a", "2": "b"}}},
		},
	}

	for _, c := range cases {
		err := stringifyYAMLMapMapKeys(c.input)
		require.NoError(t, err)
		assert.EqualValues(t, c.want, c.input)
	}
}

func TestCUE(t *testing.T) {
	in := `package foo
import "regexp"
matches: regexp.FindSubmatch(#"^([^:]*):(\d+)$"#, "localhost:443")
one: 1
two: 2
// A field using quotes.
"two-and-a-half": 2.5
list: [ 1, 2, 3 ]
`

	expected := map[string]any{
		"matches": []any{
			"localhost:443",
			"localhost",
			"443",
		},
		"one":            int64(1),
		"two":            int64(2),
		"two-and-a-half": 2.5,
		"list":           []any{int64(1), int64(2), int64(3)},
	}

	out, err := CUE(in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	out, err = CUE(`[1,2,3]`)
	require.NoError(t, err)
	assert.Equal(t, []any{int64(1), int64(2), int64(3)}, out)

	out, err = CUE(`"hello world"`)
	require.NoError(t, err)
	assert.Equal(t, "hello world", out)

	out, err = CUE(`true`)
	require.NoError(t, err)
	assert.Equal(t, true, out)

	out, err = CUE(`'\x00\x01\x02\x03\x04'`)
	require.NoError(t, err)
	assert.Equal(t, []byte{0, 1, 2, 3, 4}, out)

	out, err = CUE(`42`)
	require.NoError(t, err)
	assert.EqualValues(t, int64(42), out)

	out, err = CUE(`42.0`)
	require.NoError(t, err)
	// assert.EqualValues(t, 42.0, out)
	assert.InEpsilon(t, 42.0, out, 1e-12)

	out, err = CUE(`null`)
	require.NoError(t, err)
	assert.Nil(t, out)

	_, err = CUE(`>=0 & <=7 & >=3 & <=10`)
	require.Error(t, err)
}

func TestToCUE(t *testing.T) {
	in := map[string]any{
		"matches": []any{
			"localhost:443",
			"localhost",
			"443",
		},
		"one":            1,
		"two":            2,
		"two-and-a-half": 2.5,
		"list":           []any{1, 2, 3},
	}

	expected := `{
	list: [1, 2, 3]
	matches: ["localhost:443", "localhost", "443"]
	one:              1
	two:              2
	"two-and-a-half": 2.5
}`

	out, err := ToCUE(in)
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	out, err = ToCUE([]any{1, 2, 3})
	require.NoError(t, err)
	assert.EqualValues(t, `[1, 2, 3]`, out)

	out, err = ToCUE("hello world")
	require.NoError(t, err)
	assert.EqualValues(t, `"hello world"`, out)

	out, err = ToCUE(true)
	require.NoError(t, err)
	assert.EqualValues(t, `true`, out)

	out, err = ToCUE([]byte{0, 1, 2, 3, 4})
	require.NoError(t, err)
	assert.EqualValues(t, `'\x00\x01\x02\x03\x04'`, out)

	out, err = ToCUE(42)
	require.NoError(t, err)
	assert.EqualValues(t, `42`, out)

	out, err = ToCUE(42.0)
	require.NoError(t, err)
	assert.EqualValues(t, `42.0`, out)

	out, err = ToCUE(nil)
	require.NoError(t, err)
	assert.EqualValues(t, `null`, out)

	out, err = ToCUE(struct{}{})
	require.NoError(t, err)
	assert.EqualValues(t, `{}`, out)
}
