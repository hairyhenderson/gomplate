// Package data contains functions that parse and produce data structures in
// different formats.
//
// Supported formats are: JSON, YAML, TOML, and CSV.
package data

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/joho/godotenv"

	"github.com/Shopify/ejson"
	ejsonJson "github.com/Shopify/ejson/json"
	"github.com/hairyhenderson/gomplate/env"

	// XXX: replace once https://github.com/BurntSushi/toml/pull/179 is merged
	"github.com/hairyhenderson/toml"
	"github.com/pkg/errors"
	"github.com/ugorji/go/codec"

	// XXX: replace once https://github.com/go-yaml/yaml/issues/139 is solved
	yaml "gopkg.in/hairyhenderson/yaml.v2"
)

func init() {
	// XXX: remove once https://github.com/go-yaml/yaml/issues/139 is solved
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
}

func unmarshalObj(obj map[string]interface{}, in string, f func([]byte, interface{}) error) (map[string]interface{}, error) {
	err := f([]byte(in), &obj)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal object %s", in)
	}
	return obj, nil
}

func unmarshalArray(obj []interface{}, in string, f func([]byte, interface{}) error) ([]interface{}, error) {
	err := f([]byte(in), &obj)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal array %s", in)
	}
	return obj, nil
}

// JSON - Unmarshal a JSON Object. Can be ejson-encrypted.
func JSON(in string) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	out, err := unmarshalObj(obj, in, yaml.Unmarshal)
	if err != nil {
		return out, err
	}

	_, ok := out[ejsonJson.PublicKeyField]
	if ok {
		out, err = decryptEJSON(in)
	}
	return out, err
}

// decryptEJSON - decrypts an ejson input, and unmarshals it, stripping the _public_key field.
func decryptEJSON(in string) (map[string]interface{}, error) {
	keyDir := env.Getenv("EJSON_KEYDIR", "/opt/ejson/keys")
	key := env.Getenv("EJSON_KEY")

	rIn := bytes.NewBufferString(in)
	rOut := &bytes.Buffer{}
	err := ejson.Decrypt(rIn, rOut, keyDir, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obj := make(map[string]interface{})
	out, err := unmarshalObj(obj, rOut.String(), yaml.Unmarshal)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	delete(out, ejsonJson.PublicKeyField)
	return out, nil
}

// JSONArray - Unmarshal a JSON Array
func JSONArray(in string) ([]interface{}, error) {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// YAML - Unmarshal a YAML Object
func YAML(in string) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// YAMLArray - Unmarshal a YAML Array
func YAMLArray(in string) ([]interface{}, error) {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// TOML - Unmarshal a TOML Object
func TOML(in string) (interface{}, error) {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, toml.Unmarshal)
}

// dotEnv - Unmarshal a dotenv file
func dotEnv(in string) (interface{}, error) {
	env, err := godotenv.Unmarshal(in)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{})
	for k, v := range env {
		out[k] = v
	}
	return out, nil
}

func parseCSV(args ...string) ([][]string, []string, error) {
	in, delim, hdr := csvParseArgs(args...)
	c := csv.NewReader(strings.NewReader(in))
	c.Comma = rune(delim[0])
	records, err := c.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) > 0 {
		if hdr == nil {
			hdr = records[0]
			records = records[1:]
		} else if len(hdr) == 0 {
			hdr = make([]string, len(records[0]))
			for i := range hdr {
				hdr[i] = autoIndex(i)
			}
		}
	}
	return records, hdr, nil
}

func csvParseArgs(args ...string) (in, delim string, hdr []string) {
	delim = ","
	switch len(args) {
	case 1:
		in = args[0]
	case 2:
		in = args[1]
		switch len(args[0]) {
		case 1:
			delim = args[0]
		case 0:
			hdr = []string{}
		default:
			hdr = strings.Split(args[0], delim)
		}
	case 3:
		delim = args[0]
		hdr = strings.Split(args[1], delim)
		in = args[2]
	}
	return in, delim, hdr
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
func CSV(args ...string) ([][]string, error) {
	records, hdr, err := parseCSV(args...)
	if err != nil {
		return nil, err
	}
	records = append(records, nil)
	copy(records[1:], records)
	records[0] = hdr
	return records, nil
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
func CSVByRow(args ...string) (rows []map[string]string, err error) {
	records, hdr, err := parseCSV(args...)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		m := make(map[string]string)
		for i, v := range record {
			m[hdr[i]] = v
		}
		rows = append(rows, m)
	}
	return rows, nil
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
func CSVByColumn(args ...string) (cols map[string][]string, err error) {
	records, hdr, err := parseCSV(args...)
	if err != nil {
		return nil, err
	}
	cols = make(map[string][]string)
	for _, record := range records {
		for i, v := range record {
			cols[hdr[i]] = append(cols[hdr[i]], v)
		}
	}
	return cols, nil
}

// ToCSV -
func ToCSV(args ...interface{}) (string, error) {
	delim := ","
	var in [][]string
	if len(args) == 2 {
		var ok bool
		delim, ok = args[0].(string)
		if !ok {
			return "", errors.Errorf("Can't parse ToCSV delimiter (%v) - must be string (is a %T)", args[0], args[0])
		}
		args = args[1:]
	}
	if len(args) == 1 {
		var ok bool
		in, ok = args[0].([][]string)
		if !ok {
			return "", errors.Errorf("Can't parse ToCSV input - must be of type [][]string")
		}
	}
	b := &bytes.Buffer{}
	c := csv.NewWriter(b)
	c.Comma = rune(delim[0])
	// We output RFC4180 CSV, so force this to CRLF
	c.UseCRLF = true
	err := c.WriteAll(in)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func marshalObj(obj interface{}, f func(interface{}) ([]byte, error)) (string, error) {
	b, err := f(obj)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to marshal object %s", obj)
	}

	return string(b), nil
}

func toJSONBytes(in interface{}) ([]byte, error) {
	h := &codec.JsonHandle{}
	h.Canonical = true
	buf := new(bytes.Buffer)
	err := codec.NewEncoder(buf, h).Encode(in)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to marshal %s", in)
	}
	return buf.Bytes(), nil
}

// ToJSON - Stringify a struct as JSON
func ToJSON(in interface{}) (string, error) {
	s, err := toJSONBytes(in)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

// ToJSONPretty - Stringify a struct as JSON (indented)
func ToJSONPretty(indent string, in interface{}) (string, error) {
	out := new(bytes.Buffer)
	b, err := toJSONBytes(in)
	if err != nil {
		return "", err
	}
	err = json.Indent(out, b, "", indent)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to indent JSON %s", b)
	}

	return out.String(), nil
}

// ToYAML - Stringify a struct as YAML
func ToYAML(in interface{}) (string, error) {
	return marshalObj(in, yaml.Marshal)
}

// ToTOML - Stringify a struct as TOML
func ToTOML(in interface{}) (string, error) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(in)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to marshal %s", in)
	}
	return buf.String(), nil
}
