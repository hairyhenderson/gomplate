---
title: data functions
menu:
  main:
    parent: functions
---

A collection of functions that retrieve, parse, and convert structured data.

## `datasource`

Parses a given datasource (provided by the [`--datasource/-d`](#--datasource-d) argument).

See [Datasources](../../datasources) for (much!) more information.

### Usage

```go
datasource alias [subpath]
```

### Arguments

| name   | description |
|--------|-------|
| `alias` | the datasource alias, as provided by [`--datasource`/`-d`](../usage/#datasource-d) |
| `subpath` | _(optional)_ the subpath to use, if supported by the datasource |

### Examples

_`person.json`:_
```json
{ "name": "Dave" }
```

```console
$ gomplate -d person.json -i 'Hello {{ (datasource "person").name }}'
Hello Dave
```

## `datasourceExists`

Tests whether or not a given datasource was defined on the commandline (with the
[`--datasource/-d`](#--datasource-d) argument). This is intended mainly to allow
a template to be rendered differently whether or not a given datasource was
defined.

Note: this does _not_ verify if the datasource is reachable.

Useful when used in an `if`/`else` block.

```console
$ echo '{{if (datasourceExists "test")}}{{datasource "test"}}{{else}}no worries{{end}}' | gomplate
no worries
```

## `datasourceReachable`

Tests whether or not a given datasource is defined and reachable, where the definition of "reachable" differs by datasource, but generally means the data is able to be read successfully.

Useful when used in an `if`/`else` block.

```console
$ gomplate -i '{{if (datasourceReachable "test")}}{{datasource "test"}}{{else}}no worries{{end}}' -d test=https://bogus.example.com/wontwork.json
no worries
```

## `ds`

Alias to [`datasource`](#datasource)

## `include`

Includes the content of a given datasource (provided by the [`--datasource/-d`](../usage/#datasource-d) argument).

This is similar to [`datasource`](#datasource), except that the data is not parsed. There is no restriction on the type of data included, except that it should be textual.

### Usage

```go
include alias [subpath]
```

### Arguments

| name   | description |
|--------|-------|
| `alias` | the datasource alias, as provided by [`--datasource/-d`](../usage/#datasource-d) |
| `subpath` | _(optional)_ the subpath to use, if supported by the datasource |

### Examples

_`person.json`:_
```json
{ "name": "Dave" }
```

_`input.tmpl`:_
```go
{
  "people": [
    {{ include "person" }}
  ]
}
```

```console
$ gomplate -d person.json -f input.tmpl
{
  "people": [
    { "name": "Dave" }
  ]
}
```

## `data.JSON`

**Alias:** `json`

Converts a JSON string into an object. Only works for JSON Objects (not Arrays or other valid JSON types). This can be used to access properties of JSON objects.

#### Example

_`input.tmpl`:_
```
Hello {{ (getenv "FOO" | json).hello }}
```

```console
$ export FOO='{"hello":"world"}'
$ gomplate < input.tmpl
Hello world
```

## `data.JSONArray`

**Alias:** `jsonArray`

Converts a JSON string into a slice. Only works for JSON Arrays.

#### Example

_`input.tmpl`:_
```
Hello {{ index (getenv "FOO" | jsonArray) 1 }}
```

```console
$ export FOO='[ "you", "world" ]'
$ gomplate < input.tmpl
Hello world
```

## `data.YAML`

**Alias:** `yaml`

Converts a YAML string into an object. Only works for YAML Objects (not Arrays or other valid YAML types). This can be used to access properties of YAML objects.

#### Example

_`input.tmpl`:_
```
Hello {{ (getenv "FOO" | yaml).hello }}
```

```console
$ export FOO='hello: world'
$ gomplate < input.tmpl
Hello world
```

## `data.YAMLArray`

**Alias:** `yamlArray`

Converts a YAML string into a slice. Only works for YAML Arrays.

#### Example

_`input.tmpl`:_
```
Hello {{ index (getenv "FOO" | yamlArray) 1 }}
```

```console
$ export FOO='[ "you", "world" ]'
$ gomplate < input.tmpl
Hello world
```

## `data.TOML`

**Alias:** `toml`

Converts a [TOML](https://github.com/toml-lang/toml) document into an object.
This can be used to access properties of TOML documents.

Compatible with [TOML v0.4.0](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md).

### Usage

```go
toml input
```

Can also be used in a pipeline:
```go
input | toml
```

### Arguments

| name   | description |
|--------|-------|
| `input` | the TOML document to parse |

#### Example

_`input.tmpl`:_
```
{{ $t := `[data]
hello = "world"` -}}
Hello {{ (toml $t).hello }}
```

```console
$ gomplate -f input.tmpl
Hello world
```

## `data.CSV`

**Alias:** `csv`

Converts a CSV-format string into a 2-dimensional string array.

By default, the [RFC 4180](https://tools.ietf.org/html/rfc4180) format is
supported, but any single-character delimiter can be specified.

### Usage

```go
csv [delim] input
```

Can also be used in a pipeline:
```go
input | csv [delim]
```

### Arguments

| name   | description |
|--------|-------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `input` | the CSV-format string to parse |

### Example

_`input.tmpl`:_
```
{{ $c := `C,32
Go,25
COBOL,357` -}}
{{ range ($c | csv) -}}
{{ index . 0 }} has {{ index . 1 }} keywords.
{{ end }}
```

```console
$ gomplate < input.tmpl
C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords.
```

## `data.CSVByRow`

**Alias:** `csvByRow`

Converts a CSV-format string into a slice of maps.

By default, the [RFC 4180](https://tools.ietf.org/html/rfc4180) format is
supported, but any single-character delimiter can be specified.

Also by default, the first line of the string will be assumed to be the header,
but this can be overridden by providing an explicit header, or auto-indexing
can be used.


### Usage

```go
csv [delim] [header] input
```

Can also be used in a pipeline:
```go
input | csv [delim] [header]
```

### Arguments

| name   | description |
|--------|-------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `header`| _(optional)_ comma-separated list of column names, set to `""` to get auto-named columns (A-Z), defaults to using the first line of `input` |
| `input` | the CSV-format string to parse |

### Example

_`input.tmpl`:_
```
{{ $c := `lang,keywords
C,32
Go,25
COBOL,357` -}}
{{ range ($c | csvByRow) -}}
{{ .lang }} has {{ .keywords }} keywords.
{{ end }}
```

```console
$ gomplate < input.tmpl
C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords.
```

## `data.CSVByColumn`

**Alias:** `csvByColumn`

Like [`csvByRow`](#csvByRow), except that the data is presented as a columnar
(column-oriented) map.

### Example

_`input.tmpl`:_
```
{{ $c := `C;32
Go;25
COBOL;357` -}}
{{ $langs := ($c | csvByColumn ";" "lang,keywords").lang -}}
{{ range $langs }}{{ . }}
{{ end -}}
```

```console
$ gomplate < input.tmpl
C
Go
COBOL
```

## `data.ToJSON`

**Alias:** `toJSON`

Converts an object to a JSON document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.

#### Example

_This is obviously contrived - `json` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | json).foo | toJSON }}
```

```console
$ gomplate < input.tmpl
{"hello":"world"}
```

## `data.ToJSONPretty`

**Alias:** `toJSONPretty`

Converts an object to a pretty-printed (or _indented_) JSON document.
Input objects may be the result of functions like `data.JSON`, `data.YAML`,
`data.JSONArray`, or `data.YAMLArray` functions, or they could be provided
by a [`datasource`](../general/datasource).

The indent string must be provided as an argument.

#### Example

_`input.tmpl`:_
```
{{ `{"hello":"world"}` | data.JSON | data.ToJSONPretty "  " }}
```

```console
$ gomplate < input.tmpl
{
  "hello": "world"
}
```

## `data.ToYAML`

**Alias:** `toYAML`

Converts an object to a YAML document. Input objects may be the result of
`data.JSON`, `data.YAML`, `data.JSONArray`, or `data.YAMLArray` functions,
or they could be provided by a [`datasource`](../general/datasource).

#### Example

_This is obviously contrived - `data.JSON` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | data.JSON).foo | data.ToYAML }}
```

```console
$ gomplate < input.tmpl
hello: world
```

## `data.ToTOML`

**Alias:** `toTOML`

Converts an object to a [TOML](https://github.com/toml-lang/toml) document.

### Usage

```go
data.ToTOML obj
```

Can also be used in a pipeline:
```go
obj | data.ToTOML
```

### Arguments

| name   | description |
|--------|-------|
| `obj`  | the object to marshal as a TOML document |

#### Example

```console
$ gomplate -i '{{ `{"foo":"bar"}` | data.JSON | data.ToTOML }}'
foo = "bar"
```

## `data.ToCSV`

**Alias:** `toCSV`

Converts an object to a CSV document. The input object must be a 2-dimensional
array of strings (a `[][]string`). Objects produced by [`data.CSVByRow`](#conv-csvbyrow)
and [`data.CSVByColumn`](#conv-csvbycolumn) cannot yet be converted back to CSV documents.

**Note:** With the exception that a custom delimiter can be used, `data.ToCSV`
outputs according to the [RFC 4180](https://tools.ietf.org/html/rfc4180) format,
which means that line terminators are `CRLF` (Windows format, or `\r\n`). If
you require `LF` (UNIX format, or `\n`), the output can be piped through
[`strings.ReplaceAll`](../strings/#strings-replaceall) to replace `"\r\n"` with `"\n"`.

### Usage

```go
data.ToCSV [delim] input
```

Can also be used in a pipeline:
```go
input | data.ToCSV [delim]
```

### Arguments

| name   | description |
|--------|-------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `input` | the object to convert to a CSV |

### Examples

_`input.tmpl`:_
```go
{{ $rows := (jsonArray `[["first","second"],["1","2"],["3","4"]]`) -}}
{{ data.ToCSV ";" $rows }}
```

```console
$ gomplate -f input.tmpl
first,second
1,2
3,4
```
