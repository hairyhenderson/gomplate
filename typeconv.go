package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	// XXX: replace once https://github.com/BurntSushi/toml/pull/179 is merged
	"github.com/hairyhenderson/toml"
	"github.com/ugorji/go/codec"
)

// TypeConv - type conversion function
type TypeConv struct {
}

// Bool converts a string to a boolean value, using strconv.ParseBool under the covers.
// Possible true values are: 1, t, T, TRUE, true, True
// All other values are considered false.
func (t *TypeConv) Bool(in string) bool {
	if b, err := strconv.ParseBool(in); err == nil {
		return b
	}
	return false
}

func unmarshalObj(obj map[string]interface{}, in string, f func([]byte, interface{}) error) map[string]interface{} {
	err := f([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal object %s: %v", in, err)
	}
	return obj
}

func unmarshalArray(obj []interface{}, in string, f func([]byte, interface{}) error) []interface{} {
	err := f([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal array %s: %v", in, err)
	}
	return obj
}

// JSON - Unmarshal a JSON Object
func (t *TypeConv) JSON(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// JSONArray - Unmarshal a JSON Array
func (t *TypeConv) JSONArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// YAML - Unmarshal a YAML Object
func (t *TypeConv) YAML(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// YAMLArray - Unmarshal a YAML Array
func (t *TypeConv) YAMLArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// TOML - Unmarshal a TOML Object
func (t *TypeConv) TOML(in string) interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, toml.Unmarshal)
}

func parseCSV(args ...string) (records [][]string, hdr []string) {
	delim := ","
	var in string
	if len(args) == 1 {
		in = args[0]
	}
	if len(args) == 2 {
		in = args[1]
		if len(args[0]) == 1 {
			delim = args[0]
		} else if len(args[0]) == 0 {
			hdr = []string{}
		} else {
			hdr = strings.Split(args[0], delim)
		}
	}
	if len(args) == 3 {
		delim = args[0]
		hdr = strings.Split(args[1], delim)
		in = args[2]
	}
	c := csv.NewReader(strings.NewReader(in))
	c.Comma = rune(delim[0])
	records, err := c.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	if hdr == nil {
		hdr = records[0]
		records = records[1:]
	} else if len(hdr) == 0 {
		hdr = make([]string, len(records[0]))
		for i := range hdr {
			hdr[i] = autoIndex(i)
		}
	}
	return records, hdr
}

// autoIndex - calculates a default string column name given a numeric value
func autoIndex(i int) string {
	s := ""
	for n := 0; n <= i/26; n++ {
		s += string('A' + i%26)
	}
	return s
}

// CSV - Unmarshal CSV
// parameters:
//  delim - (optional) the (single-character!) field delimiter, defaults to ","
//     in - the CSV-format string to parse
// returns:
//  an array of rows, which are arrays of cells (strings)
func (t *TypeConv) CSV(args ...string) [][]string {
	records, hdr := parseCSV(args...)
	records = append(records, nil)
	copy(records[1:], records)
	records[0] = hdr
	return records
}

// CSVByRow - Unmarshal CSV in a row-oriented form
// parameters:
//  delim - (optional) the (single-character!) field delimiter, defaults to ","
//    hdr - (optional) comma-separated list of column names,
//          set to "" to get auto-named columns (A-Z), omit
//          to use the first line
//     in - the CSV-format string to parse
// returns:
//  an array of rows, indexed by the header name
func (t *TypeConv) CSVByRow(args ...string) (rows []map[string]string) {
	records, hdr := parseCSV(args...)
	for _, record := range records {
		m := make(map[string]string)
		for i, v := range record {
			m[hdr[i]] = v
		}
		rows = append(rows, m)
	}
	return rows
}

// CSVByColumn - Unmarshal CSV in a Columnar form
// parameters:
//  delim - (optional) the (single-character!) field delimiter, defaults to ","
//    hdr - (optional) comma-separated list of column names,
//          set to "" to get auto-named columns (A-Z), omit
//          to use the first line
//     in - the CSV-format string to parse
// returns:
//  a map of columns, indexed by the header name. values are arrays of strings
func (t *TypeConv) CSVByColumn(args ...string) (cols map[string][]string) {
	records, hdr := parseCSV(args...)
	cols = make(map[string][]string)
	for _, record := range records {
		for i, v := range record {
			cols[hdr[i]] = append(cols[hdr[i]], v)
		}
	}
	return cols
}

// ToCSV -
func (t *TypeConv) ToCSV(args ...interface{}) string {
	delim := ","
	var in [][]string
	if len(args) == 2 {
		d, ok := args[0].(string)
		if ok {
			delim = d
		} else {
			log.Fatalf("Can't parse ToCSV delimiter (%v) - must be string (is a %T)", args[0], args[0])
		}
		in, ok = args[1].([][]string)
		if !ok {
			log.Fatal("Can't parse ToCSV input - must be of type [][]string")
		}
	}
	if len(args) == 1 {
		var ok bool
		in, ok = args[0].([][]string)
		if !ok {
			log.Fatal("Can't parse ToCSV input - must be of type [][]string")
		}
	}
	b := &bytes.Buffer{}
	c := csv.NewWriter(b)
	c.Comma = rune(delim[0])
	// We output RFC4180 CSV, so force this to CRLF
	c.UseCRLF = true
	err := c.WriteAll(in)
	if err != nil {
		log.Fatal(err)
	}
	return string(b.Bytes())
}

func marshalObj(obj interface{}, f func(interface{}) ([]byte, error)) string {
	b, err := f(obj)
	if err != nil {
		log.Fatalf("Unable to marshal object %s: %v", obj, err)
	}

	return string(b)
}

func toJSONBytes(in interface{}) []byte {
	h := &codec.JsonHandle{}
	h.Canonical = true
	buf := new(bytes.Buffer)
	err := codec.NewEncoder(buf, h).Encode(in)
	if err != nil {
		log.Fatalf("Unable to marshal %s: %v", in, err)
	}
	return buf.Bytes()
}

// ToJSON - Stringify a struct as JSON
func (t *TypeConv) ToJSON(in interface{}) string {
	return string(toJSONBytes(in))
}

// ToJSONPretty - Stringify a struct as JSON (indented)
func (t *TypeConv) toJSONPretty(indent string, in interface{}) string {
	out := new(bytes.Buffer)
	b := toJSONBytes(in)
	err := json.Indent(out, b, "", indent)
	if err != nil {
		log.Fatalf("Unable to indent JSON %s: %v", b, err)
	}

	return string(out.Bytes())
}

// ToYAML - Stringify a struct as YAML
func (t *TypeConv) ToYAML(in interface{}) string {
	return marshalObj(in, yaml.Marshal)
}

// ToTOML - Stringify a struct as TOML
func (t *TypeConv) ToTOML(in interface{}) string {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(in)
	if err != nil {
		log.Fatalf("Unable to marshal %s: %v", in, err)
	}
	return string(buf.Bytes())
}

// Slice creates a slice from a bunch of arguments
func (t *TypeConv) Slice(args ...interface{}) []interface{} {
	return args
}

// Indent - indent each line of the string with the given indent string
func (t *TypeConv) indent(indent, s string) string {
	var res []byte
	bol := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if bol && c != '\n' {
			res = append(res, indent...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return string(res)
}

// Join concatenates the elements of a to create a single string.
// The separator string sep is placed between elements in the resulting string.
//
// This is functionally identical to strings.Join, except that each element is
// coerced to a string first
func (t *TypeConv) Join(in interface{}, sep string) string {
	s, ok := in.([]string)
	if ok {
		return strings.Join(s, sep)
	}

	var a []interface{}
	a, ok = in.([]interface{})
	if ok {
		b := make([]string, len(a))
		for i := range a {
			b[i] = toString(a[i])
		}
		return strings.Join(b, sep)
	}

	log.Fatal("Input to Join must be an array")
	return ""
}

// Has determines whether or not a given object has a property with the given key
func (t *TypeConv) Has(in interface{}, key string) bool {
	av := reflect.ValueOf(in)
	kv := reflect.ValueOf(key)

	if av.Kind() == reflect.Map {
		return av.MapIndex(kv).IsValid()
	}

	return false
}

func toString(in interface{}) string {
	if s, ok := in.(string); ok {
		return s
	}
	if s, ok := in.(fmt.Stringer); ok {
		return s.String()
	}
	if i, ok := in.(int); ok {
		return strconv.Itoa(i)
	}
	if u, ok := in.(uint64); ok {
		return strconv.FormatUint(u, 10)
	}
	if f, ok := in.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	if b, ok := in.(bool); ok {
		return strconv.FormatBool(b)
	}
	if in == nil {
		return "nil"
	}
	return fmt.Sprintf("%s", in)
}
