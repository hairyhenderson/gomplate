// Package parsers has internal parsers for various formats.
package parsers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"github.com/Shopify/ejson"
	ejsonJson "github.com/Shopify/ejson/json"
	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/joho/godotenv"

	// XXX: replace once https://github.com/BurntSushi/toml/pull/179 is merged
	"github.com/hairyhenderson/toml"
	"github.com/ugorji/go/codec"

	"github.com/hairyhenderson/yaml"
)

func unmarshalObj(obj map[string]interface{}, in string, f func([]byte, interface{}) error) (map[string]interface{}, error) {
	err := f([]byte(in), &obj)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal object %s: %w", in, err)
	}
	return obj, nil
}

func unmarshalArray(obj []interface{}, in string, f func([]byte, interface{}) error) ([]interface{}, error) {
	err := f([]byte(in), &obj)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal array %s: %w", in, err)
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
	keyDir := getenv("EJSON_KEYDIR", "/opt/ejson/keys")
	key := getenv("EJSON_KEY")

	rIn := bytes.NewBufferString(in)
	rOut := &bytes.Buffer{}
	err := ejson.Decrypt(rIn, rOut, keyDir, key)
	if err != nil {
		return nil, err
	}
	obj := make(map[string]interface{})
	out, err := unmarshalObj(obj, rOut.String(), yaml.Unmarshal)
	if err != nil {
		return nil, err
	}
	delete(out, ejsonJson.PublicKeyField)
	return out, nil
}

// a reimplementation of env.Getenv to avoid import cycles
func getenv(key string, def ...string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	p := os.Getenv(key + "_FILE")
	if p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return ""
		}

		val = strings.TrimSpace(string(b))
	}

	if val == "" && len(def) > 0 {
		return def[0]
	}

	return val
}

// JSONArray - Unmarshal a JSON Array
func JSONArray(in string) ([]interface{}, error) {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// YAML - Unmarshal a YAML Object
func YAML(in string) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	s := strings.NewReader(in)
	d := yaml.NewDecoder(s)
	for {
		err := d.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if obj != nil {
			break
		}
	}

	err := stringifyYAMLMapMapKeys(obj)
	return obj, err
}

// YAMLArray - Unmarshal a YAML Array
func YAMLArray(in string) ([]interface{}, error) {
	obj := make([]interface{}, 1)
	s := strings.NewReader(in)
	d := yaml.NewDecoder(s)
	for {
		err := d.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if obj != nil {
			break
		}
	}
	err := stringifyYAMLArrayMapKeys(obj)
	return obj, err
}

// stringifyYAMLArrayMapKeys recurses into the input array and changes all
// non-string map keys to string map keys. Modifies the input array.
func stringifyYAMLArrayMapKeys(in []interface{}) error {
	if _, changed := stringifyMapKeys(in); changed {
		return fmt.Errorf("stringifyYAMLArrayMapKeys: output type did not match input type, this should be impossible")
	}
	return nil
}

// stringifyYAMLMapMapKeys recurses into the input map and changes all
// non-string map keys to string map keys. Modifies the input map.
func stringifyYAMLMapMapKeys(in map[string]interface{}) error {
	if _, changed := stringifyMapKeys(in); changed {
		return fmt.Errorf("stringifyYAMLMapMapKeys: output type did not match input type, this should be impossible")
	}
	return nil
}

// stringifyMapKeys recurses into in and changes all instances of
// map[interface{}]interface{} to map[string]interface{}. This is useful to
// work around the impedance mismatch between JSON and YAML unmarshaling that's
// described here: https://github.com/go-yaml/yaml/issues/139
//
// Taken and modified from https://github.com/gohugoio/hugo/blob/cdfd1c99baa22d69e865294dfcd783811f96c880/parser/metadecoders/decoder.go#L257, Apache License 2.0
// Originally inspired by https://github.com/stripe/stripe-mock/blob/24a2bb46a49b2a416cfea4150ab95781f69ee145/mapstr.go#L13, MIT License
func stringifyMapKeys(in interface{}) (interface{}, bool) {
	switch in := in.(type) {
	case []interface{}:
		for i, v := range in {
			if vv, replaced := stringifyMapKeys(v); replaced {
				in[i] = vv
			}
		}
	case map[string]interface{}:
		for k, v := range in {
			if vv, changed := stringifyMapKeys(v); changed {
				in[k] = vv
			}
		}
	case map[interface{}]interface{}:
		res := make(map[string]interface{})

		for k, v := range in {
			ks := conv.ToString(k)
			if vv, replaced := stringifyMapKeys(v); replaced {
				res[ks] = vv
			} else {
				res[ks] = v
			}
		}
		return res, true
	}

	return nil, false
}

// TOML - Unmarshal a TOML Object
func TOML(in string) (interface{}, error) {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, toml.Unmarshal)
}

// DotEnv - Unmarshal a dotenv file
func DotEnv(in string) (interface{}, error) {
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
	s := &strings.Builder{}
	for n := 0; n <= i/26; n++ {
		s.WriteRune('A' + rune(i%26))
	}
	return s.String()
}

// CSV - Unmarshal CSV
// parameters:
//
//	delim - (optional) the (single-character!) field delimiter, defaults to ","
//	   in - the CSV-format string to parse
//
// returns:
//
//	an array of rows, which are arrays of cells (strings)
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
//
//	delim - (optional) the (single-character!) field delimiter, defaults to ","
//	  hdr - (optional) list of column names separated by `delim`,
//	        set to "" to get auto-named columns (A-Z), omit
//	        to use the first line
//	   in - the CSV-format string to parse
//
// returns:
//
//	an array of rows, indexed by the header name
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
//
//	delim - (optional) the (single-character!) field delimiter, defaults to ","
//	  hdr - (optional) list of column names separated by `delim`,
//	        set to "" to get auto-named columns (A-Z), omit
//	        to use the first line
//	   in - the CSV-format string to parse
//
// returns:
//
//	a map of columns, indexed by the header name. values are arrays of strings
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
			return "", fmt.Errorf("can't parse ToCSV delimiter (%v) - must be string (is a %T)", args[0], args[0])
		}
		args = args[1:]
	}
	if len(args) == 1 {
		switch a := args[0].(type) {
		case [][]string:
			in = a
		case [][]interface{}:
			in = make([][]string, len(a))
			for i, v := range a {
				in[i] = conv.ToStrings(v...)
			}
		case []interface{}:
			in = make([][]string, len(a))
			for i, v := range a {
				ar, ok := v.([]interface{})
				if !ok {
					return "", fmt.Errorf("can't parse ToCSV input - must be a two-dimensional array (like [][]string or [][]interface{}) (was %T)", args[0])
				}
				in[i] = conv.ToStrings(ar...)
			}
		default:
			return "", fmt.Errorf("can't parse ToCSV input - must be a two-dimensional array (like [][]string or [][]interface{}) (was %T)", args[0])
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
		return "", fmt.Errorf("unable to marshal object %s: %w", obj, err)
	}

	return string(b), nil
}

func toJSONBytes(in interface{}) ([]byte, error) {
	h := &codec.JsonHandle{}
	h.Canonical = true
	buf := new(bytes.Buffer)
	err := codec.NewEncoder(buf, h).Encode(in)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal %s: %w", in, err)
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
		return "", fmt.Errorf("unable to indent JSON %s: %w", b, err)
	}

	return out.String(), nil
}

// ToYAML - Stringify a struct as YAML
func ToYAML(in interface{}) (string, error) {
	// I'd use yaml.Marshal, but between v2 and v3 the indent has changed from
	// 2 to 4. This explicitly sets it back to 2.
	marshal := func(in interface{}) (out []byte, err error) {
		buf := &bytes.Buffer{}
		e := yaml.NewEncoder(buf)
		e.SetIndent(2)
		defer e.Close()
		err = e.Encode(in)
		return buf.Bytes(), err
	}

	return marshalObj(in, marshal)
}

// ToTOML - Stringify a struct as TOML
func ToTOML(in interface{}) (string, error) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(in)
	if err != nil {
		return "", fmt.Errorf("unable to marshal %s: %w", in, err)
	}
	return buf.String(), nil
}

// CUE - Unmarshal a CUE expression into the appropriate type
func CUE(in string) (interface{}, error) {
	cuectx := cuecontext.New()
	val := cuectx.CompileString(in)

	if val.Err() != nil {
		return nil, fmt.Errorf("unable to process CUE: %w", val.Err())
	}

	switch val.Kind() {
	case cue.StructKind:
		out := map[string]interface{}{}
		err := val.Decode(&out)
		return out, err
	case cue.ListKind:
		out := []interface{}{}
		err := val.Decode(&out)
		return out, err
	case cue.BytesKind:
		out := []byte{}
		err := val.Decode(&out)
		return out, err
	case cue.StringKind:
		out := ""
		err := val.Decode(&out)
		return out, err
	case cue.IntKind:
		out := 0
		err := val.Decode(&out)
		return out, err
	case cue.NumberKind, cue.FloatKind:
		out := 0.0
		err := val.Decode(&out)
		return out, err
	case cue.BoolKind:
		out := false
		err := val.Decode(&out)
		return out, err
	case cue.NullKind:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported CUE type %q", val.Kind())
	}
}

func ToCUE(in interface{}) (string, error) {
	cuectx := cuecontext.New()
	v := cuectx.Encode(in)
	if v.Err() != nil {
		return "", v.Err()
	}

	bs, err := format.Node(v.Syntax())
	if err != nil {
		return "", err
	}

	return string(bs), nil
}
