package data

import (
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/stretchr/testify/assert"
)

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
