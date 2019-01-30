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
	f["datasourceReachable"] = d.DatasourceReachable
	f["defineDatasource"] = d.DefineDatasource
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
