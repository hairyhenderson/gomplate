package data

import (
	"github.com/hairyhenderson/gomplate/v4/internal/parsers"
)

// temporary aliases for parser functions while I figure out if they need to be
// exported from the internal parsers package

// JSON - Unmarshal a JSON Object. Can be ejson-encrypted.
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var JSON = parsers.JSON

// JSONArray - Unmarshal a JSON Array
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var JSONArray = parsers.JSONArray

// YAML - Unmarshal a YAML Object
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var YAML = parsers.YAML

// YAMLArray - Unmarshal a YAML Array
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var YAMLArray = parsers.YAMLArray

// TOML - Unmarshal a TOML Object
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var TOML = parsers.TOML

// CSV - Unmarshal CSV
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var CSV = parsers.CSV

// CSVByRow - Unmarshal CSV in a row-oriented form
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var CSVByRow = parsers.CSVByRow

// CSVByColumn - Unmarshal CSV in a Columnar form
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var CSVByColumn = parsers.CSVByColumn

// ToCSV -
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var ToCSV = parsers.ToCSV

// ToJSON - Stringify a struct as JSON
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var ToJSON = parsers.ToJSON

// ToJSONPretty - Stringify a struct as JSON (indented)
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var ToJSONPretty = parsers.ToJSONPretty

// ToYAML - Stringify a struct as YAML
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var ToYAML = parsers.ToYAML

// ToTOML - Stringify a struct as TOML
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var ToTOML = parsers.ToTOML

// CUE - Unmarshal a CUE expression into the appropriate type
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var CUE = parsers.CUE

// ToCUE - Stringify a struct as CUE
//
// Deprecated: will be removed in a future version of gomplate. If you have a
// need for this, please open an issue!
var ToCUE = parsers.ToCUE
