package parsers

import (
	"encoding/csv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// marshalAll checks that parsed data can be written back out in each
// supported format without panicking. Errors are fine - not every value can
// be represented in every format.
func marshalAll(v any) {
	_, _ = ToJSON(v)
	_, _ = ToJSONPretty("  ", v)
	_, _ = ToYAML(v)
	_, _ = ToTOML(v)
}

func FuzzJSON(f *testing.F) {
	f.Add(`{"foo": {"bar": "baz"}, "one": 1, "true": true}`)
	f.Add(`[1, "two", {"three": 3.0}]`)
	f.Add(`{"_public_key": "abc", "a": "b"}`)
	f.Add(``)

	f.Fuzz(func(_ *testing.T, in string) {
		if out, err := JSON(in); err == nil {
			marshalAll(out)
		}
		if out, err := JSONArray(in); err == nil {
			marshalAll(out)
		}
	})
}

func FuzzYAML(f *testing.F) {
	f.Add("foo: bar\none: 1\n")
	f.Add("- a\n- b: c\n")
	f.Add("? [a, b]\n: c\n")
	f.Add("1: 2\ntrue: false\n")
	f.Add("---\n---\nfoo: bar\n")
	f.Add("a: &x 1\nb: *x\n")

	f.Fuzz(func(_ *testing.T, in string) {
		if out, err := YAML(in); err == nil {
			marshalAll(out)
		}
		if out, err := YAMLArray(in); err == nil {
			marshalAll(out)
		}
	})
}

func FuzzTOML(f *testing.F) {
	f.Add("foo = \"bar\"\n[table]\none = 1\n")
	f.Add("[[arr]]\na = 1\n[[arr]]\na = 2\n")
	f.Add("d = 1979-05-27T07:32:00Z\n")

	f.Fuzz(func(_ *testing.T, in string) {
		if out, err := TOML(in); err == nil {
			marshalAll(out)
		}
	})
}

func FuzzDotEnv(f *testing.F) {
	f.Add("FOO=bar\nBAZ=\"qux\"\n")
	f.Add("export A='b'\n# comment\nC=${A}\n")

	f.Fuzz(func(_ *testing.T, in string) {
		if out, err := DotEnv(in); err == nil {
			marshalAll(out)
		}
	})
}

func FuzzCUE(f *testing.F) {
	f.Add(`foo: "bar"`)
	f.Add(`[1, 2, 3]`)
	f.Add(`'bytes'`)
	f.Add(`a: int | *1`)
	f.Add(`null`)

	f.Fuzz(func(_ *testing.T, in string) {
		if out, err := CUE(in); err == nil {
			marshalAll(out)
			_, _ = ToCUE(out)
		}
	})
}

func FuzzCSV(f *testing.F) {
	f.Add(uint8(1), ",", "a,b", "first,second\n1,2\n")
	f.Add(uint8(2), ";", "", "a;b\n1;2\n")
	f.Add(uint8(2), "", "", "a,b\n1,2\n")
	f.Add(uint8(2), "a,b", "", "1,2,3\n")
	f.Add(uint8(3), "§", "a§b", "1§2§3\n")
	f.Add(uint8(3), "", "a,b", "1,2\n")

	f.Fuzz(func(_ *testing.T, nargs uint8, delim, hdr, in string) {
		var args []string
		switch nargs % 3 {
		case 0:
			args = []string{in}
		case 1:
			args = []string{delim, in}
		case 2:
			args = []string{delim, hdr, in}
		}

		// arbitrary input should give an error, not a panic
		_, _ = CSV(args...)
		_, _ = CSVByRow(args...)
		_, _ = CSVByColumn(args...)
	})
}

// FuzzCSVRoundTrip checks that records written by ToCSV with a given
// delimiter can be read back by CSV with the same delimiter.
func FuzzCSVRoundTrip(f *testing.F) {
	f.Add(",", uint8(2), "a\x00b\x00c\x00d")
	f.Add(";", uint8(3), "has;delim\x00has \"quotes\"\x00has\nnewline")
	f.Add("§", uint8(1), "one\x00two")

	f.Fuzz(func(t *testing.T, delim string, ncols uint8, fields string) {
		comma, _ := utf8.DecodeRuneInString(delim)
		if utf8.RuneCountInString(delim) != 1 || !validCSVDelim(comma) {
			t.Skip()
		}

		cols := int(ncols%8) + 1
		cells := strings.Split(fields, "\x00")
		for len(cells)%cols != 0 {
			cells = append(cells, "x")
		}

		records := [][]string{}
		for i := 0; i < len(cells); i += cols {
			row := cells[i : i+cols]
			for j, c := range row {
				// CSV can't preserve a bare CR - it's normalized to LF on read
				row[j] = strings.ReplaceAll(c, "\r", "")
			}
			// a record with one empty field is written as an empty line,
			// which the reader skips
			if cols == 1 && row[0] == "" {
				row[0] = "x"
			}
			records = append(records, row)
		}

		out, err := ToCSV(delim, records)
		require.NoError(t, err)

		got, err := CSV(delim, out)
		require.NoError(t, err, "CSV failed to parse ToCSV output %q", out)
		assert.Equal(t, records, got)
	})
}

// validCSVDelim reports whether encoding/csv accepts r as a delimiter.
func validCSVDelim(r rune) bool {
	w := csv.NewWriter(&strings.Builder{})
	w.Comma = r
	return w.Write([]string{"a"}) == nil
}
