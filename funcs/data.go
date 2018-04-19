package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/data"
)

var (
	dataNS     *DataFuncs
	dataNSInit sync.Once
)

// DataNS -
func DataNS() *DataFuncs {
	dataNSInit.Do(func() { dataNS = &DataFuncs{} })
	return dataNS
}

// AddDataFuncs -
func AddDataFuncs(f map[string]interface{}, d *data.Data) {
	f["datasource"] = d.Datasource
	f["ds"] = d.Datasource
	f["datasourceExists"] = d.DatasourceExists
	f["include"] = d.Include

	f["data"] = DataNS

	f["json"] = DataNS().JSON
	f["jsonArray"] = DataNS().JSONArray
	f["yaml"] = DataNS().YAML
	f["yamlArray"] = DataNS().YAMLArray
	f["toml"] = DataNS().TOML
	f["csv"] = DataNS().CSV
	f["csvByRow"] = DataNS().CSVByRow
	f["csvByColumn"] = DataNS().CSVByColumn
	f["toJSON"] = DataNS().ToJSON
	f["toJSONPretty"] = DataNS().ToJSONPretty
	f["toYAML"] = DataNS().ToYAML
	f["toTOML"] = DataNS().ToTOML
	f["toCSV"] = DataNS().ToCSV
}

// DataFuncs -
type DataFuncs struct{}

// JSON -
func (f *DataFuncs) JSON(in interface{}) map[string]interface{} {
	return data.JSON(conv.ToString(in))
}

// JSONArray -
func (f *DataFuncs) JSONArray(in interface{}) []interface{} {
	return data.JSONArray(conv.ToString(in))
}

// YAML -
func (f *DataFuncs) YAML(in interface{}) map[string]interface{} {
	return data.YAML(conv.ToString(in))
}

// YAMLArray -
func (f *DataFuncs) YAMLArray(in interface{}) []interface{} {
	return data.YAMLArray(conv.ToString(in))
}

// TOML -
func (f *DataFuncs) TOML(in interface{}) interface{} {
	return data.TOML(conv.ToString(in))
}

// CSV -
func (f *DataFuncs) CSV(args ...string) [][]string {
	return data.CSV(args...)
}

// CSVByRow -
func (f *DataFuncs) CSVByRow(args ...string) (rows []map[string]string) {
	return data.CSVByRow(args...)
}

// CSVByColumn -
func (f *DataFuncs) CSVByColumn(args ...string) (cols map[string][]string) {
	return data.CSVByColumn(args...)
}

// ToCSV -
func (f *DataFuncs) ToCSV(args ...interface{}) string {
	return data.ToCSV(args...)
}

// ToJSON -
func (f *DataFuncs) ToJSON(in interface{}) string {
	return data.ToJSON(in)
}

// ToJSONPretty -
func (f *DataFuncs) ToJSONPretty(indent string, in interface{}) string {
	return data.ToJSONPretty(indent, in)
}

// ToYAML -
func (f *DataFuncs) ToYAML(in interface{}) string {
	return data.ToYAML(in)
}

// ToTOML -
func (f *DataFuncs) ToTOML(in interface{}) string {
	return data.ToTOML(in)
}
