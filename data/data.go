// Package data contains functions that parse and produce data structures in
// different formats.
//
// Supported formats are: JSON, YAML, TOML, and CSV.
package data

import (
	"github.com/hairyhenderson/gomplate/v3/internal/datafuncs"
)

// The contents of this file have been moved to internal/datafuncs to avoid
// circular imports... They may move back in a future refactoring!

var (
	// JSON - Unmarshal a JSON Object. Can be ejson-encrypted.
	JSON = datafuncs.JSON

	// JSONArray - Unmarshal a JSON Array
	JSONArray = datafuncs.JSONArray

	// YAML - Unmarshal a YAML Object
	YAML = datafuncs.YAML

	// YAMLArray - Unmarshal a YAML Array
	YAMLArray = datafuncs.YAMLArray

	// TOML - Unmarshal a TOML Object
	TOML = datafuncs.TOML

	// dotEnv - Unmarshal a dotenv file
	dotEnv = datafuncs.DotEnv

	// CSV - Unmarshal CSV
	// parameters:
	//  delim - (optional) the (single-character!) field delimiter, defaults to ","
	//     in - the CSV-format string to parse
	// returns:
	//  an array of rows, which are arrays of cells (strings)
	CSV = datafuncs.CSV

	// CSVByRow - Unmarshal CSV in a row-oriented form
	// parameters:
	//  delim - (optional) the (single-character!) field delimiter, defaults to ","
	//    hdr - (optional) comma-separated list of column names,
	//          set to "" to get auto-named columns (A-Z), omit
	//          to use the first line
	//     in - the CSV-format string to parse
	// returns:
	//  an array of rows, indexed by the header name
	CSVByRow = datafuncs.CSVByRow

	// CSVByColumn - Unmarshal CSV in a Columnar form
	// parameters:
	//  delim - (optional) the (single-character!) field delimiter, defaults to ","
	//    hdr - (optional) comma-separated list of column names,
	//          set to "" to get auto-named columns (A-Z), omit
	//          to use the first line
	//     in - the CSV-format string to parse
	// returns:
	//  a map of columns, indexed by the header name. values are arrays of strings
	CSVByColumn = datafuncs.CSVByColumn

	// ToCSV -
	ToCSV = datafuncs.ToCSV

	// ToJSON - Stringify a struct as JSON
	ToJSON = datafuncs.ToJSON

	// ToJSONPretty - Stringify a struct as JSON (indented)
	ToJSONPretty = datafuncs.ToJSONPretty

	// ToYAML - Stringify a struct as YAML
	ToYAML = datafuncs.ToYAML

	// ToTOML - Stringify a struct as TOML
	ToTOML = datafuncs.ToTOML
)
