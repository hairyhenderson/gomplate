package funcs

import (
	"context"

	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/data"
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
func CreateDataFuncs(ctx context.Context,
	//nolint:staticcheck
	d *data.Data) map[string]interface{} {
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

	f["json"] = ns.YAML
	f["jsonArray"] = ns.YAML
	f["yaml"] = ns.YAML
	f["yamlArray"] = ns.YAML
	f["toml"] = ns.TOML
	f["csv"] = ns.CSV
	f["csvByRow"] = ns.CSVByRow
	f["csvByColumn"] = ns.CSVByColumn
	f["toJSON"] = ns.ToJSON
	f["toJSONPretty"] = ns.ToJSONPretty
	f["toYAML"] = ns.ToYAML
	f["toTOML"] = ns.ToTOML
	f["toCSV"] = ns.ToCSV
	return f
}

// DataFuncs -
type DataFuncs struct {
	ctx context.Context
}

// YAML -
func (f *DataFuncs) YAML(in interface{}) (interface{}, error) {
	return data.YAML(conv.ToString(in))
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

// ToCSV -
func (f *DataFuncs) ToCSV(args ...interface{}) (string, error) {
	return data.ToCSV(args...)
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
