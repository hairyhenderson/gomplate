package main

import (
	"net/url"
	"text/template"

	"github.com/hairyhenderson/gomplate/funcs"
)

// initFuncs - The function mappings are defined here!
func initFuncs(data *Data) template.FuncMap {
	env := &Env{}
	typeconv := &TypeConv{}

	f := template.FuncMap{
		"getenv":           env.Getenv,
		"bool":             typeconv.Bool,
		"has":              typeconv.Has,
		"json":             typeconv.JSON,
		"jsonArray":        typeconv.JSONArray,
		"yaml":             typeconv.YAML,
		"yamlArray":        typeconv.YAMLArray,
		"toml":             typeconv.TOML,
		"csv":              typeconv.CSV,
		"csvByRow":         typeconv.CSVByRow,
		"csvByColumn":      typeconv.CSVByColumn,
		"slice":            typeconv.Slice,
		"indent":           typeconv.indent,
		"join":             typeconv.Join,
		"toJSON":           typeconv.ToJSON,
		"toJSONPretty":     typeconv.toJSONPretty,
		"toYAML":           typeconv.ToYAML,
		"toTOML":           typeconv.ToTOML,
		"toCSV":            typeconv.ToCSV,
		"urlParse":         url.Parse,
		"datasource":       data.Datasource,
		"ds":               data.Datasource,
		"datasourceExists": data.DatasourceExists,
		"include":          data.include,
	}
	funcs.AWSFuncs(f)
	funcs.AddBase64Funcs(f)
	funcs.AddNetFuncs(f)
	funcs.AddReFuncs(f)
	funcs.AddStringFuncs(f)
	return f
}
