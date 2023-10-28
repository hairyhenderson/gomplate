package funcs

import (
	"context"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/data"
)

// DataNS -
//
// Deprecated: don't use
func DataNS() *DataFuncs {
	return &DataFuncs{}
}

// AddDataFuncs -
//
// Deprecated: use [CreateDataFuncs] instead
func AddDataFuncs(f map[string]interface{}, d *data.Data) {
	for k, v := range CreateDataFuncs(context.Background(), d) {
		f[k] = v
	}
}

// CreateDataFuncs -
//
//nolint:staticcheck
func CreateDataFuncs(ctx context.Context, d *data.Data) map[string]interface{} {
	f := map[string]interface{}{}
	f["datasource"] = d.Datasource
	f["ds"] = d.Datasource
	f["datasourceExists"] = d.DatasourceExists
	f["datasourceReachable"] = d.DatasourceReachable
	f["defineDatasource"] = d.DefineDatasource
	f["include"] = d.Include
	f["listDatasources"] = d.ListDatasources

	ns := &DataFuncs{ctx}

	f["data"] = func() interface{} { return ns }

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
func (f *DataFuncs) JSON(in interface{}) (map[string]interface{}, error) {
	return data.JSON(conv.ToString(in))
}

// JSONArray -
func (f *DataFuncs) JSONArray(in interface{}) ([]interface{}, error) {
	return data.JSONArray(conv.ToString(in))
}

// YAML -
func (f *DataFuncs) YAML(in interface{}) (map[string]interface{}, error) {
	return data.YAML(conv.ToString(in))
}

// YAMLArray -
func (f *DataFuncs) YAMLArray(in interface{}) ([]interface{}, error) {
	return data.YAMLArray(conv.ToString(in))
}

// TOML -
func (f *DataFuncs) TOML(in interface{}) (interface{}, error) {
	return data.TOML(conv.ToString(in))
}

// CSV -
func (f *DataFuncs) CSV(args ...string) ([][]string, error) {
	return data.CSV(args...)
}

// CSVByRow -
func (f *DataFuncs) CSVByRow(args ...string) (rows []map[string]string, err error) {
	return data.CSVByRow(args...)
}

// CSVByColumn -
func (f *DataFuncs) CSVByColumn(args ...string) (cols map[string][]string, err error) {
	return data.CSVByColumn(args...)
}

// CUE -
func (f *DataFuncs) CUE(in interface{}) (interface{}, error) {
	return data.CUE(conv.ToString(in))
}

// ToCSV -
func (f *DataFuncs) ToCSV(args ...interface{}) (string, error) {
	return data.ToCSV(args...)
}

// ToCUE -
func (f *DataFuncs) ToCUE(in interface{}) (string, error) {
	return data.ToCUE(in)
}

// ToJSON -
func (f *DataFuncs) ToJSON(in interface{}) (string, error) {
	return data.ToJSON(in)
}

// ToJSONPretty -
func (f *DataFuncs) ToJSONPretty(indent string, in interface{}) (string, error) {
	return data.ToJSONPretty(indent, in)
}

// ToYAML -
func (f *DataFuncs) ToYAML(in interface{}) (string, error) {
	return data.ToYAML(in)
}

// ToTOML -
func (f *DataFuncs) ToTOML(in interface{}) (string, error) {
	return data.ToTOML(in)
}
