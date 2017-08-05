package data

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"log"
	"strings"

	// XXX: replace once https://github.com/BurntSushi/toml/pull/179 is merged
	"github.com/hairyhenderson/toml"
	"github.com/ugorji/go/codec"
	yaml "gopkg.in/yaml.v2"
)

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
func JSON(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// JSONArray - Unmarshal a JSON Array
func JSONArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// YAML - Unmarshal a YAML Object
func YAML(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// YAMLArray - Unmarshal a YAML Array
func YAMLArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// TOML - Unmarshal a TOML Object
func TOML(in string) interface{} {
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
func CSV(args ...string) [][]string {
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
func CSVByRow(args ...string) (rows []map[string]string) {
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
func CSVByColumn(args ...string) (cols map[string][]string) {
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
func ToCSV(args ...interface{}) string {
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
func ToJSON(in interface{}) string {
	return string(toJSONBytes(in))
}

// ToJSONPretty - Stringify a struct as JSON (indented)
func ToJSONPretty(indent string, in interface{}) string {
	out := new(bytes.Buffer)
	b := toJSONBytes(in)
	err := json.Indent(out, b, "", indent)
	if err != nil {
		log.Fatalf("Unable to indent JSON %s: %v", b, err)
	}

	return string(out.Bytes())
}

// ToYAML - Stringify a struct as YAML
func ToYAML(in interface{}) string {
	return marshalObj(in, yaml.Marshal)
}

// ToTOML - Stringify a struct as TOML
func ToTOML(in interface{}) string {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(in)
	if err != nil {
		log.Fatalf("Unable to marshal %s: %v", in, err)
	}
	return string(buf.Bytes())
}
