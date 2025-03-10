package funcs

import (
	"context"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/internal/parsers"
)

// CreateDataFuncs -
func CreateDataFuncs(ctx context.Context) map[string]any {
	f := map[string]any{}

	ns := &DataFuncs{ctx}

	f["data"] = func() any { return ns }

	f["json"] = ns.JSON
	f["jsonArray"] = ns.JSONArray
	f["yaml"] = ns.YAML
	f["yamlArray"] = ns.YAMLArray
	f["toml"] = ns.TOML
	f["csv"] = ns.CSV
	f["csvByRow"] = ns.CSVByRow
	f["csvByColumn"] = ns.CSVByColumn
	f["cue"] = ns.CUE
	f["toJSON"] = ns.ToJSON
	f["toJSONPretty"] = ns.ToJSONPretty
	f["toYAML"] = ns.ToYAML
	f["toTOML"] = ns.ToTOML
	f["toCSV"] = ns.ToCSV
	f["toCUE"] = ns.ToCUE
	return f
}

// DataFuncs -
type DataFuncs struct {
	ctx context.Context
}

// JSON -
func (f *DataFuncs) JSON(in any) (map[string]any, error) {
	return parsers.JSON(conv.ToString(in))
}

// JSONArray -
func (f *DataFuncs) JSONArray(in any) ([]any, error) {
	return parsers.JSONArray(conv.ToString(in))
}

// YAML -
func (f *DataFuncs) YAML(in any) (map[string]any, error) {
	return parsers.YAML(conv.ToString(in))
}

// YAMLArray -
func (f *DataFuncs) YAMLArray(in any) ([]any, error) {
	return parsers.YAMLArray(conv.ToString(in))
}

// TOML -
func (f *DataFuncs) TOML(in any) (any, error) {
	return parsers.TOML(conv.ToString(in))
}

// CSV -
func (f *DataFuncs) CSV(args ...string) ([][]string, error) {
	return parsers.CSV(args...)
}

// CSVByRow -
func (f *DataFuncs) CSVByRow(args ...string) (rows []map[string]string, err error) {
	return parsers.CSVByRow(args...)
}

// CSVByColumn -
func (f *DataFuncs) CSVByColumn(args ...string) (cols map[string][]string, err error) {
	return parsers.CSVByColumn(args...)
}

// CUE -
func (f *DataFuncs) CUE(in any) (any, error) {
	return parsers.CUE(conv.ToString(in))
}

// ToCSV -
func (f *DataFuncs) ToCSV(args ...any) (string, error) {
	return parsers.ToCSV(args...)
}

// ToCUE -
func (f *DataFuncs) ToCUE(in any) (string, error) {
	return parsers.ToCUE(in)
}

// ToJSON -
func (f *DataFuncs) ToJSON(in any) (string, error) {
	return parsers.ToJSON(in)
}

// ToJSONPretty -
func (f *DataFuncs) ToJSONPretty(indent string, in any) (string, error) {
	return parsers.ToJSONPretty(indent, in)
}

// ToYAML -
func (f *DataFuncs) ToYAML(in any) (string, error) {
	return parsers.ToYAML(in)
}

// ToTOML -
func (f *DataFuncs) ToTOML(in any) (string, error) {
	return parsers.ToTOML(in)
}
