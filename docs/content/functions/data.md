---
title: data functions
menu:
  main:
    parent: functions
---

A collection of functions that retrieve, parse, and convert structured data.

## `datasource`

**Alias:** `ds`

Parses a given datasource (provided by the [`--datasource/-d`](../../usage/#--datasource-d) argument or [`defineDatasource`](#definedatasource)).

If the `alias` is undefined, but is a valid URL, `datasource` will dynamically read from that URL.

See [Datasources](../../datasources) for (much!) more information.

_Added in gomplate [v0.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v0.5.0)_
### Usage

```
datasource alias [subpath]
```

### Arguments

| name | description |
|------|-------------|
| `alias` | _(required)_ the datasource alias (or a URL for dynamic use) |
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
[`--datasource/-d`](../../usage/#--datasource-d) argument). This is intended mainly to allow
a template to be rendered differently whether or not a given datasource was
defined.

Note: this does _not_ verify if the datasource is reachable.

Useful when used in an `if`/`else` block.

_Added in gomplate [v1.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.3.0)_
### Usage

```
datasourceExists alias
```

### Arguments

| name | description |
|------|-------------|
| `alias` | _(required)_ the datasource alias |

### Examples

```console
$ echo '{{if (datasourceExists "test")}}{{datasource "test"}}{{else}}no worries{{end}}' | gomplate
no worries
```

## `datasourceReachable`

Tests whether or not a given datasource is defined and reachable, where the definition of "reachable" differs by datasource, but generally means the data is able to be read successfully.

Useful when used in an `if`/`else` block.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
datasourceReachable alias
```

### Arguments

| name | description |
|------|-------------|
| `alias` | _(required)_ the datasource alias |

### Examples

```console
$ gomplate -i '{{if (datasourceReachable "test")}}{{datasource "test"}}{{else}}no worries{{end}}' -d test=https://bogus.example.com/wontwork.json
no worries
```

## `listDatasources`

Lists all the datasources defined, list returned will be sorted in ascending order.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
listDatasources
```


### Examples

```console
$ gomplate -d person=env:///FOO -d bar=env:///BAR -i '{{range (listDatasources)}} Datasource-{{.}} {{end}}'
Datasource-bar
Datasource-person
```

## `defineDatasource`

Define a datasource alias with target URL inside the template. Overridden by the [`--datasource/-d`](../../usage/#--datasource-d) flag.

Note: once a datasource is defined, it can not be redefined (i.e. if this function is called twice with the same alias, only the first applies).

This function can provide a good way to set a default datasource when sharing templates.

See [Datasources](../../datasources) for (much!) more information.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
defineDatasource alias url
```

### Arguments

| name | description |
|------|-------------|
| `alias` | _(required)_ the datasource alias |
| `url` | _(required)_ the datasource's URL |

### Examples

_`person.json`:_
```json
{ "name": "Dave" }
```

```console
$ gomplate -i '{{ defineDatasource "person" "person.json" }}Hello {{ (ds "person").name }}'
Hello Dave
$ FOO='{"name": "Daisy"}' gomplate -d person=env:///FOO -i '{{ defineDatasource "person" "person.json" }}Hello {{ (ds "person").name }}'
Hello Daisy
```

## `include`

Includes the content of a given datasource (provided by the [`--datasource/-d`](../../usage/#--datasource-d) argument).

This is similar to [`datasource`](#datasource), except that the data is not parsed. There is no restriction on the type of data included, except that it should be textual.

_Added in gomplate [v1.8.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.8.0)_
### Usage

```
include alias [subpath]
```

### Arguments

| name | description |
|------|-------------|
| `alias` | _(required)_ the datasource alias, as provided by [`--datasource/-d`](../../usage/#--datasource-d) |
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

Converts a JSON string into an object. Works for JSON Objects, but will
also parse JSON Arrays. Will not parse other valid JSON types.

For more explicit JSON Array support, see [`data.JSONArray`](#datajsonarray).

#### Encrypted JSON support (EJSON)

If the input is in the [EJSON](https://github.com/Shopify/ejson) format (i.e. has a `_public_key` field), this function will attempt to decrypt the document first. A private key must be provided by one of these methods:

- set the `EJSON_KEY` environment variable to the private key's value
- set the `EJSON_KEY_FILE` environment variable to the path to a file containing the private key
- set the `EJSON_KEYDIR` environment variable to the path to a directory containing private keys (filename must be the public key), just like [`ejson decrypt`'s `--keydir`](https://github.com/Shopify/ejson/blob/master/man/man1/ejson.1.ronn) flag. Defaults to `/opt/ejson/keys`.

_Added in gomplate [v1.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.4.0)_
### Usage

```
data.JSON in
```
```
in | data.JSON
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

### Examples

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

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.JSONArray in
```
```
in | data.JSONArray
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

### Examples

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

Converts a YAML string into an object. Works for YAML Objects but will
also parse YAML Arrays. This can be used to access properties of YAML objects.

For more explicit YAML Array support, see [`data.JSONArray`](#datayamlarray).

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.YAML in
```
```
in | data.YAML
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

### Examples

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

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.YAMLArray in
```
```
in | data.YAMLArray
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

### Examples

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

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.TOML input
```
```
input | data.TOML
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the TOML document to parse |

### Examples

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

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.CSV [delim] input
```
```
input | data.CSV [delim]
```

### Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `input` | _(required)_ the CSV-format string to parse |

### Examples

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

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.CSVByRow [delim] [header] input
```
```
input | data.CSVByRow [delim] [header]
```

### Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `header` | _(optional)_ list of column names separated by `delim`, set to `""` to get auto-named columns (A-Z), defaults to using the first line of `input` |
| `input` | _(required)_ the CSV-format string to parse |

### Examples

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

Like [`csvByRow`](#datacsvbyrow), except that the data is presented as a columnar
(column-oriented) map.

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.CSVByColumn [delim] [header] input
```
```
input | data.CSVByColumn [delim] [header]
```

### Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `header` | _(optional)_ list of column names separated by `delim`, set to `""` to get auto-named columns (A-Z), defaults to using the first line of `input` |
| `input` | _(required)_ the CSV-format string to parse |

### Examples

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

## `data.CUE`_(unreleased)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._

**Alias:** `cue`

Converts a [CUE](https://cuelang.org/) document into an object. Any type
of CUE document is supported. This can be used to access properties of CUE
documents.

Note that the `import` statement is not yet supported, and will result in
an error (except for importing builtin packages).

### Usage

```
data.CUE input
```
```
input | data.CUE
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the CUE document to parse |

### Examples

```console
$ gomplate -i '{{ $t := `data: {
    hello: "world"
  }` -}}
  Hello {{ (cue $t).data.hello }}'
Hello world
```

## `data.ToJSON`

**Alias:** `toJSON`

Converts an object to a JSON document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.ToJSON obj
```
```
obj | data.ToJSON
```

### Arguments

| name | description |
|------|-------------|
| `obj` | _(required)_ the object to marshal |

### Examples

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
by a [`datasource`](../datasources).

The indent string must be provided as an argument.

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.ToJSONPretty indent obj
```
```
obj | data.ToJSONPretty indent
```

### Arguments

| name | description |
|------|-------------|
| `indent` | _(required)_ the string to use for indentation |
| `obj` | _(required)_ the object to marshal |

### Examples

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
or they could be provided by a [`datasource`](../datasources).

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.ToYAML obj
```
```
obj | data.ToYAML
```

### Arguments

| name | description |
|------|-------------|
| `obj` | _(required)_ the object to marshal |

### Examples

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

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.ToTOML obj
```
```
obj | data.ToTOML
```

### Arguments

| name | description |
|------|-------------|
| `obj` | _(required)_ the object to marshal as a TOML document |

### Examples

```console
$ gomplate -i '{{ `{"foo":"bar"}` | data.JSON | data.ToTOML }}'
foo = "bar"
```

## `data.ToCSV`

**Alias:** `toCSV`

Converts an object to a CSV document. The input object must be a 2-dimensional
array of strings (a `[][]string`). Objects produced by [`data.CSVByRow`](#datacsvbyrow)
and [`data.CSVByColumn`](#datacsvbycolumn) cannot yet be converted back to CSV documents.

**Note:** With the exception that a custom delimiter can be used, `data.ToCSV`
outputs according to the [RFC 4180](https://tools.ietf.org/html/rfc4180) format,
which means that line terminators are `CRLF` (Windows format, or `\r\n`). If
you require `LF` (UNIX format, or `\n`), the output can be piped through
[`strings.ReplaceAll`](../strings/#stringsreplaceall) to replace `"\r\n"` with `"\n"`.

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
data.ToCSV [delim] input
```
```
input | data.ToCSV [delim]
```

### Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `input` | _(required)_ the object to convert to a CSV |

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

## `data.ToCUE`_(unreleased)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._

**Alias:** `toCUE`

Converts an object to a [CUE](https://cuelang.org/) document in canonical
format. The input object can be of any type.

This is roughly equivalent to using the `cue export --out=cue <file>`
command to convert from other formats to CUE.

### Usage

```
data.ToCUE input
```
```
input | data.ToCUE
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the object to marshal as a CUE document |

### Examples

```console
$ gomplate -i '{{ `{"foo":"bar"}` | data.JSON | data.ToCUE }}'
{
	foo: "bar"
}
```
```console
$ gomplate -i '{{ toCUE "hello world" }}'
"hello world"
```
```console
$ gomplate -i '{{ coll.Slice 1 "two" true | data.ToCUE }}'
[1, "two", true]
```
