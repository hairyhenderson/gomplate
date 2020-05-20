// Package data contains functions that parse and produce data structures in
// different formats.
//
// Supported formats are: JSON, YAML, TOML, and CSV.
package data

import (
	"github.com/hairyhenderson/gomplate/v3/internal/dataconv"
)

// JSON - Unmarshal a JSON Object. Can be ejson-encrypted.
func JSON(in string) (map[string]interface{}, error) {
	return dataconv.JSON(in)
}

// JSONArray - Unmarshal a JSON Array
func JSONArray(in string) ([]interface{}, error) {
	return dataconv.JSONArray(in)
}

// YAML - Unmarshal a YAML Object
func YAML(in string) (map[string]interface{}, error) {
	return dataconv.YAML(in)
}

// YAMLArray - Unmarshal a YAML Array
func YAMLArray(in string) ([]interface{}, error) {
	return dataconv.YAMLArray(in)
}

// TOML - Unmarshal a TOML Object
func TOML(in string) (interface{}, error) {
	return dataconv.TOML(in)
}

// dotEnv - Unmarshal a dotenv file
func dotEnv(in string) (interface{}, error) {
	return dataconv.DotEnv(in)
}

// CSV - Unmarshal CSV
// parameters:
//  delim - (optional) the (single-character!) field delimiter, defaults to ","
//     in - the CSV-format string to parse
// returns:
//  an array of rows, which are arrays of cells (strings)
func CSV(args ...string) ([][]string, error) {
	return dataconv.CSV(args...)
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
	return dataconv.CSVByRow(args...)
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
	return dataconv.CSVByColumn(args...)
}

// ToCSV -
func ToCSV(args ...interface{}) (string, error) {
	return dataconv.ToCSV(args...)
}

// ToJSON - Stringify a struct as JSON
func ToJSON(in interface{}) (string, error) {
	return dataconv.ToJSON(in)
}

// ToJSONPretty - Stringify a struct as JSON (indented)
func ToJSONPretty(indent string, in interface{}) (string, error) {
	return dataconv.ToJSONPretty(indent, in)
}

// ToYAML - Stringify a struct as YAML
func ToYAML(in interface{}) (string, error) {
	return dataconv.ToYAML(in)
}

// ToTOML - Stringify a struct as TOML
func ToTOML(in interface{}) (string, error) {
	return dataconv.ToTOML(in)
}
