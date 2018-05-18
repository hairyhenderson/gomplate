---
title: data functions
menu:
  main:
    parent: functions
---

A collection of functions that retrieve, parse, and convert structured data.

## `datasource`

Parses a given datasource (provided by the [`--datasource/-d`](#--datasource-d) argument).

Currently, `file://`, `stdin://`, `http://`, `https://`, `vault://`, and `boltdb://` URLs are supported.

Currently-supported formats are JSON, YAML, TOML, and CSV. Plain-text datasources can also be specified, but can only be safely accessed with the [`include`](#include) function.

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

### Providing datasources on standard input (`stdin`)

Normally `stdin` is used as the input for the template, but it can also be used
to provide datasources. To do this, specify a URL with the `stdin:` scheme:

```console
$ echo 'foo: bar' | gomplate -i '{{(ds "data").foo}}' -d data=stdin:///foo.yaml
bar
```

Note that the URL must have a file name with a supported extension in order for
the input to be correctly parsed. If no parsing is required (i.e. if the data
is being included verbatim with the include function), just `stdin:` is enough:

```console
$ echo 'foo' | gomplate -i '{{ include "data" }}' -d data=stdin:
foo
```

### Overriding the MIME type

On occasion it's necessary to override the MIME type a datasource is parsed with.
To accomplish this, gomplate supports a `type` query string parameter on
datasource URLs. This can contain the same value as a standard
[HTTP Content-Type](https://tools.ietf.org/html/rfc7231#section-3.1.1.1)
header.

For example, to force a file named `data.txt` to be parsed as a JSON document:

```console
$ echo '{"foo": "bar"}' > /tmp/data.txt
$ gomplate -d data=file:///tmp/data.txt?type=application/json -i '{{ (ds "data").foo }}'
bar
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
| `CONSUL_VAULT_ROLE` | Set to the name of the role to use for authenticating to Consul with [Vault's Consul secret backend](https://www.vaultproject.io/docs/secrets/consul/index.html). |
| `CONSUL_VAULT_MOUNT` | Used to override the mount-point when using Vault's Consul secret backend for authentication. Defaults to `consul`. |

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

Instead of using a non-authenticated Consul connection or connecting using the token set with the
`CONSUL_HTTP_TOKEN` environment variable, it is possible to authenticate using a dynamically generated
token fetched from Vault. This requires Vault to be configured to use the [Consul secret backend](https://www.vaultproject.io/docs/secrets/consul/index.html) and
is enabled by passing the name of the role to use in the `CONSUL_VAULT_ROLE` environment variable.

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

### Usage with AWS Systems Manager Parameter data

The special `aws+smp://` URL scheme can be used to retrieve data from the
[AWS Systems Manager]((https://aws.amazon.com/systems-manager/) (née AWS EC2 Simple 
Systems Manager) [Parameter Store](https://aws.amazon.com/systems-manager/features/#Parameter_Store).
This hierarchically organised key/value store allows you to store text, lists or encrypted
secrets for easy retrieval by AWS resources.

#### Arguments for aws+smp datasource

```go
datasource alias [subpath]
```

| name   | description |
|--------|-------|
| `alias` | the datasource alias, as provided by [`--datasource/-d`](../usage/#datasource-d) |
| `subpath` | _(optional)_ the subpath to use, using path-join semantics (add a '/' if none at start/end of path-in-the-url and subpath) |

You must grant the gomplate process IAM credentials via the AWS golang SDK default
methods (e.g. environment args, ~/.aws/* files, instance profiles) for the
`ssm.GetParameter` action.

#### Output of aws+smp datasource

The output will be a single Parameter object from the
[AWS golang SDK](https://docs.aws.amazon.com/sdk-for-go/api/service/ssm/#Parameter):

| name   | description |
|--------|-------|
| `Name` | full Parameter name |
| `Type` | `String`, `StringList` or `SecureString` |
| `Value` | textual value, comma-separated single string if StringList |
| `Version` | incrementing integer version |

If the Parameter key specified is not found (or not allowed to be read due to
missing `ssm.GetParameter` permission) an error will be generated.
There is no default.

#### Examples

Given your [AWS account's Parameter Store](https://eu-west-1.console.aws.amazon.com/ec2/v2/home#Parameters:sort=Name) has the following data:

* /foo/first/others - "Bill,Ben" (a StringList)
* /foo/first/password - "super-secret" (a SecureString)
* /foo/second/p1 - "aaa"

```console
$ echo '{{ ds "foo" }}' | gomplate -d foo=aws+smp:///foo/first/password
map[Name:/foo/first/password Type:SecureString Value:super-secret Version:1]

$ echo '{{ (ds "foo").Value }}' | gomplate -d foo=aws+smp:///foo/first/password
super-secret

$ echo '{{ (ds "foo" "/foo/first/others").Value }}' | gomplate -d foo=aws+smp:
Bill,Ben

$ echo '{{ (ds "foo" "/second/p1").Value }}' | gomplate -d foo=aws+smp:///foo/
aaa
```

### Usage with Vault data

The special `vault://` URL scheme can be used to retrieve data from [Hashicorp
Vault](https://vaultproject.io). To use this, you must either provide the Vault
server's hostname and port in the URL, or put the Vault server's URL in the
`$VAULT_ADDR` environment variable.

The `vault+http://` URL scheme can be used to indicate that request must be sent
over regular unencrypted HTTP, while `vault+https://` and `vault://` are equivalent,
and indicate that requests must be sent over HTTPS.

List support is also available when the URL ends with a `/` character. In order for this to work correctly, the authenticated token must have permission to use the [`list` capability](https://www.vaultproject.io/docs/concepts/policies.html#list) for the given path.

This table describes the currently-supported authentication mechanisms and how to use them, in order of precedence:

| auth backend | configuration |
|-------------: |---------------|
| [`approle`](https://www.vaultproject.io/docs/auth/approle.html) | Environment variables `$VAULT_ROLE_ID` and `$VAULT_SECRET_ID` must be set to the appropriate values.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_APPROLE_MOUNT`. |
| [`app-id`](https://www.vaultproject.io/docs/auth/app-id.html) | Environment variables `$VAULT_APP_ID` and `$VAULT_USER_ID` must be set to the appropriate values.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_APP_ID_MOUNT`. |
| [`github`](https://www.vaultproject.io/docs/auth/github.html) | Environment variable `$VAULT_AUTH_GITHUB_TOKEN` must be set to an appropriate value.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_GITHUB_MOUNT`. |
| [`userpass`](https://www.vaultproject.io/docs/auth/userpass.html) | Environment variables `$VAULT_AUTH_USERNAME` and `$VAULT_AUTH_PASSWORD` must be set to the appropriate values.<br/> If the backend is mounted to a different location, set `$VAULT_AUTH_USERPASS_MOUNT`. |
| [`token`](https://www.vaultproject.io/docs/auth/token.html) | Determined from either the `$VAULT_TOKEN` environment variable, or read from the file `~/.vault-token` |
| [`aws`](https://www.vaultproject.io/docs/auth/aws.html) | As a final option authentication will be attempted using the AWS auth backend. See below for more details. |

_**Note:**_ The secret values listed in the above table can either be set in environment
variables or provided in files. This can increase security when using
[Docker Swarm Secrets](https://docs.docker.com/engine/swarm/secrets/), for example.
To use files, specify the filename by appending `_FILE` to the environment variable,
(i.e. `VAULT_USER_ID_FILE`). If the non-file variable is set, this will override
any `_FILE` variable and the secret file will be ignored.

To use a Vault datasource with a single secret, just use a URL of
`vault:///secret/mysecret`. Note the 3 `/`s - the host portion of the URL is left
empty in this example.

```console
$ echo 'My voice is my passport. {{(datasource "vault").value}}' \
  | gomplate -d vault=vault:///secret/sneakers
My voice is my passport. Verify me.
```

You can also specify the secret path in the template by omitting the path portion
of the URL:

```console
$ echo 'My voice is my passport. {{(datasource "vault" "secret/sneakers").value}}' \
  | gomplate -d vault=vault:///
My voice is my passport. Verify me.
```

And the two can be mixed to scope secrets to a specific namespace:

```console
$ echo 'db_password={{(datasource "vault" "db/pass").value}}' \
  | gomplate -d vault=vault:///secret/production
db_password=prodsecret
```

If you are unable to set the `VAULT_ADDR` environment variable, or need to
specify multiple Vault datasources connecting to different servers, you can set
the address as part of the URL:

```console
$ gomplate -d v=vault://vaultserver.com/secret/foo -i '{{ (ds "v").value }}'
bar
```

It is also possible to use dynamic secrets by using the write capability of the datasource. To use,
add a URL query to the optional path (i.e. `"key?name=value&name=value"`). These values are then
included within the JSON body of the request.

```console
$ echo 'otp={{(datasource "vault" "ssh/creds/test?ip=10.1.2.3&username=user").key}}' \
  | gomplate -d vault=vault:///
otp=604a4bd5-7afd-30a2-d2d8-80c4aebc6183
```

#### Authentication using AWS details

If running on an EC2 instance authentication will be attempted using the AWS auth backend. The
optional `VAULT_AUTH_AWS_MOUNT` environment variable can be used to set the mount point to use if
it differs from the default of `aws`. Additionally `AWS_TIMEOUT` can be set (in seconds) to a value
to wait for AWS to respond before skipping the attempt.

If set, the `VAULT_AUTH_AWS_ROLE` environment variable will be used to specify the role to authenticate
using. If not set the AMI ID of the EC2 instance will be used by Vault.

If you want to allow multiple authentications using AWS EC2 auth (i.e. run gomplate multiple times) you
will need to pass the same nonce each time. This can be sent using `VAULT_AUTH_AWS_NONCE`. If not set once
will automatically be generated by AWS. The nonce used can be stored by setting `VAULT_AUTH_AWS_NONCE_OUTPUT`
to a filename. If the file doesn't exist it is created with 0600 permission.

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
