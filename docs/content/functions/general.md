---
title: other functions
menu:
  main:
    parent: functions
---

## `bool`

Converts a true-ish string to a boolean. Can be used to simplify conditional statements based on environment variables or other text input.

#### Example

_`input.tmpl`:_
```
{{if bool (getenv "FOO")}}foo{{else}}bar{{end}}
```

```console
$ gomplate < input.tmpl
bar
$ FOO=true gomplate < input.tmpl
foo
```

## `slice`

Creates a slice. Useful when needing to `range` over a bunch of variables.

#### Example

_`input.tmpl`:_
```
{{range slice "Bart" "Lisa" "Maggie"}}
Hello, {{.}}
{{- end}}
```

```console
$ gomplate < input.tmpl
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `urlParse`

Parses a string as a URL for later use. Equivalent to [url.Parse](https://golang.org/pkg/net/url/#Parse)

#### Example

_`input.tmpl`:_
```
{{ $u := urlParse "https://example.com:443/foo/bar" }}
The scheme is {{ $u.Scheme }}
The host is {{ $u.Host }}
The path is {{ $u.Path }}
```

```console
$ gomplate < input.tmpl
The scheme is https
The host is example.com:443
The path is /foo/bar
```

## `has`

Has reports whether or not a given object has a property with the given key. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

#### Example

_Let's say we're using a Vault datasource..._

_`input.tmpl`:_
```
{{ $secret := datasource "vault" "mysecret" -}}
The secret is '
{{- if (has $secret "value") }}
{{- $secret.value }}
{{- else }}
{{- $secret | toYAML }}
{{- end }}'
```

If the `secret/foo/mysecret` secret in Vault has a property named `value` set to `supersecret`:

```console
$ gomplate -d vault:///secret/foo < input.tmpl
The secret is 'supersecret'
```

On the other hand, if there is no `value` property:

```console
$ gomplate -d vault:///secret/foo < input.tmpl
The secret is 'foo: bar'
```

## `join`

Concatenates the elements of an array to create a string. The separator string sep is placed between elements in the resulting string.

#### Example

_`input.tmpl`_
```
{{ $a := `[1, 2, 3]` | jsonArray }}
{{ join $a "-" }}
```

```console
$ gomplate -f input.tmpl
1-2-3
```

## `json`

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

## `jsonArray`

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

## `yaml`

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

## `yamlArray`

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

## `toml`

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

## `csv`

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

## `csvByRow`

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

## `csvByColumn`

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

## `toJSON`

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

## `toJSONPretty`

Converts an object to a pretty-printed (or _indented_) JSON document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.

The indent string must be provided as an argument.

#### Example

_`input.tmpl`:_
```
{{ `{"hello":"world"}` | json | toJSONPretty "  " }}
```

```console
$ gomplate < input.tmpl
{
  "hello": "world"
}
```

## `toYAML`

Converts an object to a YAML document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.

#### Example

_This is obviously contrived - `json` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | json).foo | toYAML }}
```

```console
$ gomplate < input.tmpl
hello: world
```

## `toTOML`

Converts an object to a [TOML](https://github.com/toml-lang/toml) document.

### Usage

```go
toTOML obj
```

Can also be used in a pipeline:
```go
obj | toTOML
```

### Arguments

| name   | description |
|--------|-------|
| `obj`  | the object to marshal as a TOML document |

#### Example

```console
$ gomplate -i '{{ `{"foo":"bar"}` | json | toTOML }}'
foo = "bar"
```

## `toCSV`

Converts an object to a CSV document. The input object must be a 2-dimensional
array of strings (a `[][]string`). Objects produced by [`csvByRow`](#csvByRow)
and [`csvByColumn`](#csvByColumn) cannot yet be converted back to CSV documents.

**Note:** With the exception that a custom delimiter can be used, `toCSV`
outputs according to the [RFC 4180](https://tools.ietf.org/html/rfc4180) format,
which means that line terminators are `CRLF` (Windows format, or `\r\n`). If
you require `LF` (UNIX format, or `\n`), the output can be piped through
[`replaceAll`](#replaceAll) to replace `"\r\n"` with `"\n"`.

### Usage

```go
toCSV [delim] input
```

Can also be used in a pipeline:
```go
input | toCSV [delim]
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
{{ toCSV ";" $rows }}
```

```console
$ gomplate -f input.tmpl
first,second
1,2
3,4
```

## `datasource`

Parses a given datasource (provided by the [`--datasource/-d`](#--datasource-d) argument).

Currently, `file://`, `http://`, `https://`, and `vault://` URLs are supported.

Currently-supported formats are JSON, YAML, TOML, and CSV.

### Basic usage

_`person.json`:_
```json
{
  "name": "Dave"
}
```

_`input.tmpl`:_
```
Hello {{ (datasource "person").name }}
```

```console
$ gomplate -d person.json < input.tmpl
Hello Dave
```

### Usage with HTTP data

```console
$ echo 'Hello there, {{(datasource "foo").headers.Host}}...' | gomplate -d foo=https://httpbin.org/get
Hello there, httpbin.org...
```

Additional headers can be provided with the `--datasource-header`/`-H` option:

```console
$ gomplate -d foo=https://httpbin.org/get -H 'foo=Foo: bar' -i '{{(datasource "foo").headers.Foo}}'
bar
```

### Usage with Consul data

There are three supported URL schemes to retrieve data from [Consul](https://consul.io/).
The `consul://` (or `consul+http://`) scheme can optionally be used with a hostname and port to specify a server (e.g. `consul://localhost:8500`).
By default HTTP will be used, but the `consul+https://` form can be used to use HTTPS, alternatively `$CONSUL_HTTP_SSL` can be used.

If the server address isn't part of the datasource URL, `$CONSUL_HTTP_ADDR` will be checked.

The following optional environment variables can be set:

| name | usage |
|------|-------|
| `CONSUL_HTTP_ADDR` | Hostname and optional port for connecting to Consul. Defaults to `http://localhost:8500` |
| `CONSUL_TIMEOUT` | Timeout (in seconds) when communicating to Consul. Defaults to 10 seconds. |
| `CONSUL_HTTP_TOKEN` | The Consul token to use when connecting to the server. |
| `CONSUL_HTTP_AUTH` | Should be specified as `<username>:<password>`. Used to authenticate to the server. |
| `CONSUL_HTTP_SSL` | Force HTTPS if set to `true` value. Disables if set to `false`. Any value acceptable to [`strconv.ParseBool`](https://golang.org/pkg/strconv/#ParseBool) can be provided. |
| `CONSUL_TLS_SERVER_NAME` | The server name to use as the SNI host when connecting to Consul via TLS. |
| `CONSUL_CACERT` | Path to CA file for verifying Consul server using TLS. |
| `CONSUL_CAPATH` | Path to directory of CA files for verifying Consul server using TLS. |
| `CONSUL_CLIENT_CERT` | Client certificate file for certificate authentication. If this is set, `$CONSUL_CLIENT_KEY` must also be set. |
| `CONSUL_CLIENT_KEY` | Client key file for certificate authentication. If this is set, `$CONSUL_CLIENT_CERT` must also be set. |
| `CONSUL_HTTP_SSL_VERIFY` | Set to `false` to disable Consul TLS certificate checking. Any value acceptable to [`strconv.ParseBool`](https://golang.org/pkg/strconv/#ParseBool) can be provided. <br/> _Recommended only for testing and development scenarios!_ |

If a path is included it is used as a prefix for all uses of the datasource.

#### Example

```console
$ gomplate -d consul=consul:// -i '{{(datasource "consul" "foo")}}'
value for foo key
```

```console
$ gomplate -d consul=consul+https://my-consul-server.com:8533/foo -i '{{(datasource "consul" "bar")}}'
value for foo/bar key
```

```console
$ gomplate -d consul=consul:///foo -i '{{(datasource "consul" "bar/baz")}}'
value for foo/bar/baz key
```

### Usage with BoltDB data

[BoltDB](https://github.com/boltdb/bolt) is a simple local key/value store used
by many Go tools. The `boltdb://` scheme can be used to access values stored in
a BoltDB database file. The full path is provided in the URL, and the bucket name
can be specified using a URL fragment (e.g. `boltdb:///tmp/database.db#bucket`).

Access is implemented through [libkv](https://github.com/docker/libkv), and as
such, the first 8 bytes of all values are used as an incrementing last modified
index value. All values must therefore be at least 9 bytes long, with the first
8 being ignored.

The following environment variables can be set:

| name | usage |
|------|-------|
| `BOLTDB_TIMEOUT` | Timeout (in seconds) to wait for a lock on the database file when opening. |
| `BOLTDB_PERSIST` | If set keep the database open instead of closing after each read. Any value acceptable to [`strconv.ParseBool`](https://golang.org/pkg/strconv/#ParseBool) can be provided. |

### Example

```console
$ gomplate -d config=boltdb:///tmp/config.db#Bucket1 -i '{{(datasource "config" "foo")}}'
bar
```

### Usage with Vault data

The special `vault://` URL scheme can be used to retrieve data from [Hashicorp
Vault](https://vaultproject.io). To use this, you must put the Vault server's
URL in the `$VAULT_ADDR` environment variable.

This table describes the currently-supported authentication mechanisms and how to use them, in order of precedence:

| auth backend | configuration |
|-------------: |---------------|
| [`approle`](https://www.vaultproject.io/docs/auth/approle.html) | Environment variables `$VAULT_ROLE_ID` and `$VAULT_SECRET_ID` must be set to the appropriate values.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_APPROLE_MOUNT`. |
| [`app-id`](https://www.vaultproject.io/docs/auth/app-id.html) | Environment variables `$VAULT_APP_ID` and `$VAULT_USER_ID` must be set to the appropriate values.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_APP_ID_MOUNT`. |
| [`github`](https://www.vaultproject.io/docs/auth/github.html) | Environment variable `$VAULT_AUTH_GITHUB_TOKEN` must be set to an appropriate value.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_GITHUB_MOUNT`. |
| [`userpass`](https://www.vaultproject.io/docs/auth/userpass.html) | Environment variables `$VAULT_AUTH_USERNAME` and `$VAULT_AUTH_PASSWORD` must be set to the appropriate values.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_USERPASS_MOUNT`. |
| [`token`](https://www.vaultproject.io/docs/auth/token.html) | Determined from either the `$VAULT_TOKEN` environment variable, or read from the file `~/.vault-token` |

_**Note:**_ The secret values listed in the above table can either be set in environment
variables or provided in files. This can increase security when using
[Docker Swarm Secrets](https://docs.docker.com/engine/swarm/secrets/), for example.
To use files, specify the filename by appending `_FILE` to the environment variable,
(i.e. `VAULT_USER_ID_FILE`). If the non-file variable is set, this will override
any `_FILE` variable and the secret file will be ignored.

To use a Vault datasource with a single secret, just use a URL of
`vault:///secret/mysecret`. Note the 3 `/`s - the host portion of the URL is left
empty.

```console
$ echo 'My voice is my passport. {{(datasource "vault").value}}' \
  | gomplate -d vault=vault:///secret/sneakers
My voice is my passport. Verify me.
```

You can also specify the secret path in the template by using a URL of `vault://`
(or `vault:///`, or `vault:`):
```console
$ echo 'My voice is my passport. {{(datasource "vault" "secret/sneakers").value}}' \
  | gomplate -d vault=vault://
My voice is my passport. Verify me.
```

And the two can be mixed to scope secrets to a specific namespace:

```console
$ echo 'db_password={{(datasource "vault" "db/pass").value}}' \
  | gomplate -d vault=vault:///secret/production
db_password=prodsecret
```

It is also possible to use dynamic secrets by using the write capibility of the datasource. To use
add an additional query string style section to the optional key name (i.e.
`"key?name=value&name=value"`). These values are then included within the JSON body of the request.

```console
$ echo 'otp={{(datasource "vault" "ssh/creds/test?ip=10.1.2.3&username=user").key}}' \
  | gomplate -d vault=vault:///
otp=604a4bd5-7afd-30a2-d2d8-80c4aebc6183
```

## `datasourceExists`

Tests whether or not a given datasource was defined on the commandline (with the
[`--datasource/-d`](#--datasource-d) argument). This is intended mainly to allow
a template to be rendered differently whether or not a given datasource was
defined.

Note: this does _not_ verify if the datasource is reachable.

Useful when used in an `if`/`else` block

```console
$ echo '{{if (datasourceExists "test")}}{{datasource "test"}}{{else}}no worries{{end}}' | gomplate
no worries
```

## `ds`

Alias to [`datasource`](#datasource)

## `include`

Includes the content of a given datasource (provided by the [`--datasource/-d`](../usage/#datasource-d) argument).

This is similar to [`datasource`](#datasource),
except that the data is not parsed.

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
