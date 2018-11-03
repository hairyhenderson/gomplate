---
title: Datasources
weight: 13
menu: main
---

Datasources are an optional, but central concept in gomplate. While the basic flow of template rendering is taking an input template and rendering it into an output, there often is need to include data from one or more sources external to the template itself.

Some common use-cases include injecting sensitive material like passwords (which should not be stored in source-control with the templates), or providing simplified configuration formats that can be fed to a template to provide a much more complex output.

Datasources can be defined with the [`--datasource`/`-d`][] command-line flag or the [`defineDatasource`][] function, and referenced via an _alias_ inside the template, using a function such as [`datasource`][] or [`include`][].

Since datasources are defined separately from the template, the same templates can be used with different datasources and even different datasource types. For example, gomplate could be run on a developer machine with a `file` datasource pointing to a JSON file containing test data, where the same template could be used in a production environment using a `consul` datasource with the real production data.

## URL Format

All datasources are defined with a [URL][]. As a refresher, a URL is made up of the following components:

```pre
  foo://example.com:8042/over/there?name=ferret#nose
  \_/   \______________/\_________/ \_________/ \__/
   |           |            |            |        |
scheme     authority       path        query   fragment
```

For our purposes, the _scheme_ and the _path_ components are especially important, though the other components are used by certain datasources for particular purposes.

| component | purpose |
|-----------|---------|
| _scheme_ | Identifies which [datasource](#supported-datasources) to access. All datasources require a scheme (except for `file` when using relative paths), and some datasources allow multiple different schemes to clarify access modes, such as `consul+https` |
| _authority_ | Used only by networked datasources, and can be omitted in some of those cases |
| _path_ | Can be omitted, but usually used as the basis of the locator for the datasource. If the path ends with a `/` character, [directory](#directory-datasources) semantics are used. |
| _query_ | Used rarely for datasources where information must be provided in order to get a reasonable reply (such as generating dynamic secrets with Vault), or for [overriding MIME types](#overriding-mime-types) |
| _fragment_ | Used rarely for accessing a subset of the given path (such as a bucket name in a BoltDB database) |

## Supported datasources

Gomplate supports a number of datasources, each specified with a particular URL scheme. The table below describes these datasources. The names in the _Type_ column link to further documentation for each specific datasource.

| Type | URL Scheme(s) | Description |
|------|---------------|-------------|
| [AWS Systems Manager Parameter Store](#using-aws-smp-datasources) | `aws+smp` | [AWS Systems Manager Parameter Store][AWS SMP] is a hierarchically-organized key/value store which allows storage of text, lists, or encrypted secrets for retrieval by AWS resources |
| [BoltDB](#using-boltdb-datasources) | `boltdb` | [BoltDB][] is a simple local key/value store used by many Go tools |
| [Consul](#using-consul-datasources) | `consul`, `consul+http`, `consul+https` | [HashiCorp Consul][] provides (among many other features) a key/value store |
| [Environment](#using-env-datasources) | `env` | Environment variables can be used as datasources - useful for testing |
| [File](#using-file-datasources) | `file` | Files can be read in any of the [supported formats](#mime-types), including by piping through standard input (`Stdin`). [Directories](#directory-datasources) are also supported. |
| [HTTP](#using-http-datasources) | `http`, `https` | Data can be sourced from HTTP/HTTPS sites in many different formats. Arbitrary HTTP headers can be set with the [`--datasource-header`/`-H`][] flag |
| [Stdin](#using-stdin-datasources) | `stdin` | A special case of the `file` datasource; allows piping through standard input (`Stdin`) |
| [Vault](#using-vault-datasources) | `vault`, `vault+http`, `vault+https` | [HashiCorp Vault][] is an industry-leading open-source secret management tool. [List support](#directory-datasources) is also available. |

## Directory Datasources

When the _path_ component of the URL ends with a `/` character, the datasource is read with _directory_ semantics. Not all datasource types support this, and for those that don't support the notion of a directory, the behaviour is currently undefined. See each documentation section for details.
 
Currently the following datasources support directory semantics:

- [File](#using-file-datasources)
- [Vault](#using-vault-datasources) - translates to Vault's [LIST](https://www.vaultproject.io/api/index.html#reading-writing-and-listing-secrets) method

When accessing a directory datasource, an array of key names is returned, and can be iterated through to access each individual value contained within.

For example, a group of configuration key/value pairs (named `one`, `two`, and `three`, with values `v1`, `v2`, and `v3` respectively) could be rendered like this: 

_template.tmpl:_
```
{{ range (datasource "config") -}}
{{ . }} = {{ (datasource "config" .).value }}
{{- end }}
```

```console
$ gomplate -d config=vault:///secret/configs/ -f template.tmpl
one = v1
two = v2
three = v3
```

## MIME Types

Gomplate will read and parse a number of data formats. The appropriate type will be set automatically, if possible, either based on file extension (for the `file` and `http` datasources), or the [HTTP Content-Type][] header, if available. If an unsupported type is detected, gomplate will exit with an error.

These are the supported types:

| Format | MIME Type | Extension(s) | Notes |
|--------|-----------|-------|------|
| CSV | `text/csv` | | Uses the [`data.CSV`][] function to present the file as a 2-dimensional row-first string array |
| JSON | `application/json` | | [JSON][] _objects_ are assumed, and arrays or other values are not parsed with this type. Uses the [`data.JSON`][] function for parsing |
| JSON Array | `application/array+json` | | A special type for parsing datasources containing just JSON arrays. Uses the [`data.JSONArray`][] function for parsing |
| Plain Text | `text/plain` | | Unstructured, and as such only intended for use with the [`include`][] function |
| TOML | `application/toml` | | Parses [TOML][] with the [`data.TOML`][] function |
| YAML | `application/yaml` | | Parses [YAML][] with the [`data.YAML`][] function |

### Overriding MIME Types

On occasion it's necessary to override the detected (via file extension or `Content-Type` header) MIME type. To accomplish this, gomplate supports a `type` query string parameter on datasource URLs. This can contain the same value as a standard [HTTP Content-Type][] header.

For example, to force a file named `data.txt` to be parsed as a JSON document:

```console
$ echo '{"foo": "bar"}' > /tmp/data.txt
$ gomplate -d data=file:///tmp/data.txt?type=application/json -i '{{ (ds "data").foo }}'
bar
```

## Using `aws+smp` datasources

The `aws+smp://` scheme can be used to retrieve data from the [AWS Systems Manager](https://aws.amazon.com/systems-manager/) (née AWS EC2 Simple  Systems Manager) [Parameter Store](https://aws.amazon.com/systems-manager/features/#Parameter_Store). This hierarchically organized key/value store allows you to store text, lists or encrypted secrets for easy retrieval by AWS resources. See [the AWS Systems Manager documentation](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-su-create.html#sysman-paramstore-su-create-about) for details on creating these parameters.

You must grant `gomplate` permission via IAM credentials with the one of the methods supported by the [AWS SDK for Go][] (e.g. environment variables, `~/.aws/*` files, instance profiles) for the [`ssm:GetParameter` action](https://docs.aws.amazon.com/systems-manager/latest/userguide/auth-and-access-control-permissions-reference.html). <!-- List support further requires `ssm.GetParameters` permissions. -->

### URL Considerations

For `aws+smp`, only the _scheme_ and _path_ components are necessary to be defined. Other URL components are ignored.

### Output

The output will be a single `Parameter` object from the
[AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/api/service/ssm/#Parameter):

| name   | description |
|--------|-------|
| `Name` | full Parameter name |
| `Type` | `String`, `StringList` or `SecureString` |
| `Value` | textual value, comma-separated single string if StringList |
| `Version` | incrementing integer version |

If the Parameter key specified is not found (or not allowed to be read due to missing permissions) an error will be generated. There is no default.

### Examples

Given your [AWS account's Parameter Store](https://eu-west-1.console.aws.amazon.com/ec2/v2/home#Parameters:sort=Name) has the following data:

- `/foo/first/others` - `Bill,Ben` (a StringList)
- `/foo/first/password` - `super-secret` (a SecureString)
- `/foo/second/p1` - `aaa`

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

## Using `boltdb` datasources

[BoltDB](https://github.com/boltdb/bolt) is a simple local key/value store used by many Go tools. The `boltdb://` scheme can be used to access values stored in a BoltDB database file. The full path is provided in the URL, and the bucket name can be specified using a URL fragment (e.g. `boltdb:///tmp/database.db#bucket`).

**Note:** Access is implemented through [`libkv`](https://github.com/docker/libkv), and as such, the first 8 bytes of all values are used as an incrementing last modified index value. All values must therefore be at least 9 bytes long, with the first 8 being ignored.

The following environment variables can be set:

| name | usage |
|------|-------|
| `BOLTDB_TIMEOUT` | Timeout (in seconds) to wait for a lock on the database file when opening. |
| `BOLTDB_PERSIST` | If set keep the database open instead of closing after each read. Any value acceptable to [`strconv.ParseBool`](https://golang.org/pkg/strconv/#ParseBool) can be provided. |

### URL Considerations

For `boltdb`, the _scheme_, _path_, and _fragment_ are used.

The _path_ must point to a BoltDB database on the local file system, while the _fragment_ must provide the name of the bucket to use.

### Example

```console
$ gomplate -d config=boltdb:///tmp/config.db#Bucket1 -i '{{(datasource "config" "foo")}}'
bar
```

## Using `consul` datasources

Gomplate supports retrieving data from [HashiCorp Consul][]'s [KV Store](https://www.consul.io/api/kv.html).

### URL Considerations

For `consul`, the _scheme_, _authority_, and _path_ components are used.

- the _scheme_ URL component can be one of three values: `consul`, `consul+http`, and `consul+https`. The first two are equivalent, while the third instructs the client to connect to Consul over an encrypted HTTPS connection. Encryption can alternately be enabled by use of the `$CONSUL_HTTP_SSL` environment variable.
- the _authority_ is used to specify the server to connect to (e.g. `consul://localhost:8500`), but if not specified, the `$CONSUL_HTTP_ADDR` environment variable will be used.
- the _path_ can be provided to select a specific key, or a key prefix

### Consul Environment Variables

The following optional environment variables are understood by the Consul datasource:

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
| `CONSUL_VAULT_MOUNT` | Used to override the mount-point when using Vault's Consul secret back-end for authentication. Defaults to `consul`. |

### Authentication

Instead of using a non-authenticated Consul connection, you can authenticate with these methods:

- provide an [ACL Token](https://www.consul.io/docs/guides/acl.html#acl-tokens) in the `CONSUL_HTTP_TOKEN` environment variable
- use HTTP Basic Auth by setting the `CONSUL_HTTP_AUTH` environment variable
- dynamically generate an ACL token with Vault. This requires Vault to be configured to use the [Consul secret backend](https://www.vaultproject.io/docs/secrets/consul/index.html) and is enabled by passing the name of the role to use in the `CONSUL_VAULT_ROLE` environment variable.

### Examples

```console
$ gomplate -d consul=consul:// -i '{{(datasource "consul" "foo")}}'
value for foo key

$ gomplate -d consul=consul+https://my-consul-server.com:8533/foo -i '{{(datasource "consul" "bar")}}'
value for foo/bar key

$ gomplate -d consul=consul:///foo -i '{{(datasource "consul" "bar/baz")}}'
value for foo/bar/baz key
```

## Using `env` datasources

The `env` datasource type provides access to environment variables. This can be useful for rendering templates that would normally use a different sort of datasource, in test and development scenarios.

No hierarchy or directory semantics are currently supported.

**Note:** Variable names are _case-sensitive!_

### URL Considerations

The _scheme_ and either the _path_ or the _opaque_ part are used, and the _query_ component can be used to [override the MIME type](#overriding-mime-types).

- the _scheme_ must be `env`
- one of the _path_ or _opaque_ component is required, and is interpreted as the environment variable's name. Leading `/` characters are stripped from the _path_.

### Examples

```console
$ gomplate -d user=env:USER -i 'Hello {{ include "user" }}!'
Hello hairyhenderson!

$ gomplate -d homedir=env:///HOME -i '{{ files.IsDir (ds "homedir") }}'
true

$ export foo='{"one":1, "two":2}'
$ gomplate -d foo=env:/foo?type=application/json -i '{{ (ds "foo").two }}'
2
```

## Using `file` datasources

The `file` datasource type provides access to files in any of the [supported formats](#mime-types). [Directory datasource](#directory-datasources) semantics are supported.

### URL Considerations

The _scheme_ and _path_ are used, and the _query_ component can be used to [override the MIME type](#overriding-mime-types).

- the _scheme_ must be `file` for absolute URLs, but may be omitted to allow setting relative paths
- the _path_ component is required, and can be an absolute or relative path, and if the file being referenced is in the current working directory, the file's base name (without extension) is used as the datasource alias in absence of an explicit alias. [Directory](#directory-datasources) semantics are available when the path ends with a `/` character.

### Examples

_`person.json`:_
```json
{
  "name": "Dave"
}
```

_implicit alias:_
```console
$ gomplate -d person.json -i 'Hello {{ (datasource "person").name }}'
Hello Dave
```

_explicit alias:_
```console
$ gomplate -d person=./person.json -i 'Hello {{ (datasource "person").name }}'
Hello Dave

$ gomplate -d person=../path/to/person.json -i 'Hello {{ (datasource "person").name }}'
Hello Dave

$ gomplate -d person=file:///tmp/person.json -i 'Hello {{ (datasource "person").name }}'
Hello Dave
```

## Using `http` datasources

To access datasources from HTTP sites or APIs, simply use a `http` or `https` URL:

```console
$ gomplate -d foo=https://httpbin.org/get -i 'Hello there, {{ (ds "foo").headers.Host }}...'
Hello there, httpbin.org...
$ gomplate -d foo=https://httpbin.org/get -i '{{ $d := ds "foo" }}Hello there, {{ $d.headers.Host }}, you are looking very {{ index $d.headers "User-Agent" }} today...'
Hello there, httpbin.org, you are looking very Go-http-client/1.1 today...
```

### Sending HTTP headers

Additional headers can be provided with the `--datasource-header`/`-H` option:

```console
$ gomplate -d foo=https://httpbin.org/get -H 'foo=Foo: bar' -i '{{(datasource "foo").headers.Foo}}'
bar
```

This can be useful for providing API tokens to authenticated HTTP-based APIs.


## Using `stdin` datasources

Normally _Stdin_ is used as the input for the template, but it can also be used
to stream a datasource. To do this, specify a URL with the `stdin:` scheme.

In order for structured input to be correctly parsed, the URL can be given a "fake" file name with a supported extension, or the [MIME type can be explicitly set](#overriding-mime-types). If the input is unstructured (i.e. if the data is being included verbatim with the [`include`][] function), the scheme alone is enough.

```console
$ echo 'foo: bar' | gomplate -i '{{(ds "data").foo}}' -d data=stdin:///foo.yaml
bar
$ echo 'foo' | gomplate -i '{{ include "data" }}' -d data=stdin:
foo
$ echo '["one", "two"]' | gomplate -i '{{index (ds "data") 1 }}' -d data=stdin:?type=application/array%2Bjson
two
```

## Using `vault` datasources

Gomplate can retrieve secrets and other data from [HashiCorp Vault][].

### URL Considerations

The _scheme_, _authority_, _path_, and _query_ URL components are used by this datasource.

- the _scheme_ must be one of `vault`, `vault+https` (same as `vault`), or `vault+http`. The latter can be used to access [dev mode](https://www.vaultproject.io/docs/concepts/dev-server.html) Vault servers, for test purposes. Otherwise, all connections to Vault are encrypted with TLS.
- the _authority_ component can optionally be used to specify the Vault server's hostname and port. This overrides the value of `$VAULT_ADDR`.
- the _path_ component can optionally be used to specify a full or partial path to a secret. The second argument to the [`datasource`][] function is appended to provide the full secret path. [Directory](#directory-datasources) semantics are available when the path ends with a `/` character.
- the _query_ component is used to provide parameters to dynamic secret back-ends that require these. The values are included in the JSON body of the `PUT` request.

These are all valid `vault` URLs:

- `vault:`, `vault://`, `vault:///` - these all require the [`datasource`][] function to provide the secret path
- `vault://vault.example.com:8200` - connect to `vault.example.com` over HTTPS at port `8200`. The path will be provided by [`datasource`][]
- `vault:///ssh/creds/foo?ip=10.1.2.3&username=user` - create a dynamic secret with the parameters `ip` and `username` provided in the body
- `vault:///secret/configs/` - returns a list of key names with the prefix of `secret/configs/`

### Vault Authentication

This table describes the currently-supported authentication mechanisms and how to use them, in order of precedence:

| auth back-end | configuration |
|-------------:|---------------|
| [`approle`](https://www.vaultproject.io/docs/auth/approle.html) | Environment variables `$VAULT_ROLE_ID` and `$VAULT_SECRET_ID` must be set to the appropriate values.<br/> If the back-end is mounted to a different location, set `$VAULT_AUTH_APPROLE_MOUNT`. |
| [`app-id`](https://www.vaultproject.io/docs/auth/app-id.html) | Environment variables `$VAULT_APP_ID` and `$VAULT_USER_ID` must be set to the appropriate values.<br/> If the back-end is mounted to a different location, set `$VAULT_AUTH_APP_ID_MOUNT`. |
| [`github`](https://www.vaultproject.io/docs/auth/github.html) | Environment variable `$VAULT_AUTH_GITHUB_TOKEN` must be set to an appropriate value.<br/> If the back-end is mounted to a different location, set `$VAULT_AUTH_GITHUB_MOUNT`. |
| [`userpass`](https://www.vaultproject.io/docs/auth/userpass.html) | Environment variables `$VAULT_AUTH_USERNAME` and `$VAULT_AUTH_PASSWORD` must be set to the appropriate values.<br/> If the back-end is mounted to a different location, set `$VAULT_AUTH_USERPASS_MOUNT`. |
| [`token`](https://www.vaultproject.io/docs/auth/token.html) | Determined from either the `$VAULT_TOKEN` environment variable, or read from the file `~/.vault-token` |
| [`aws`](https://www.vaultproject.io/docs/auth/aws.html) | The env var  `$VAULT_AUTH_AWS_ROLE` defines the [role](https://www.vaultproject.io/api/auth/aws/index.html#role-4) to log in with - defaults to the AMI ID of the EC2 instance. Usually a [Client Nonce](https://www.vaultproject.io/docs/auth/aws.html#client-nonce) should be used as well. Set `$VAULT_AUTH_AWS_NONCE` to the nonce value. The nonce can be generated and stored by setting `$VAULT_AUTH_AWS_NONCE_OUTPUT` to a path on the local filesystem.<br/>If the back-end is mounted to a different location, set `$VAULT_AUTH_AWS_MOUNT`.|

_**Note:**_ The secret values listed in the above table can either be set in environment variables or provided in files. This can increase security when using [Docker Swarm Secrets](https://docs.docker.com/engine/swarm/secrets/), for example. To use files, specify the filename by appending `_FILE` to the environment variable, (i.e. `VAULT_USER_ID_FILE`). If the non-file variable is set, this will override any `_FILE` variable and the secret file will be ignored.

### Vault Permissions 

The correct capabilities must be allowed for the [authenticated](#vault-authentication) credentials. See the [Vault documentation](https://www.vaultproject.io/docs/concepts/policies.html#capabilities) for full details.

- regular secret read operations require the `read` capability
- dynamic secret generation requires the `create` and `update` capabilities
- list support requires the `list` capability

### Vault Environment variables

In addition to the variables documented [above](#vault-authentication), a number of environment variables are interpreted by the Vault client, and are documented in the [official Vault documentation](https://www.vaultproject.io/docs/commands/index.html#environment-variables).

### Examples

```console
$ gomplate -d vault=vault:///secret/sneakers -i 'My voice is my passport. {{(datasource "vault").value}}' 
My voice is my passport. Verify me.
```

You can also specify the secret path in the template by omitting the path portion of the URL:

```console
$ gomplate -d vault=vault:/// -i 'My voice is my passport. {{(datasource "vault" "secret/sneakers").value}}'
My voice is my passport. Verify me.
```

And the two can be mixed to scope secrets to a specific namespace:

```console
$ gomplate -d vault=vault:///secret/production -i 'db_password={{(datasource "vault" "db/pass").value}}' 
db_password=prodsecret
```

If you are unable to set the `VAULT_ADDR` environment variable, or need to
specify multiple Vault datasources connecting to different servers, you can set
the address as part of the URL:

```console
$ gomplate -d v=vault://vaultserver.com/secret/foo -i '{{ (ds "v").value }}'
bar
```

To use dynamic secrets:

```console
$ gomplate -d vault=vault:/// -i 'otp={{(ds "vault" "ssh/creds/test?ip=10.1.2.3&username=user").key}}'
otp=604a4bd5-7afd-30a2-d2d8-80c4aebc6183
```

With the AWS auth back-end:

```console
$ export VAULT_AUTH_AWS_NONCE_FILE=/tmp/vault-aws-nonce
$ export VAULT_AUTH_AWS_NONCE_OUTPUT=$VAULT_AUTH_AWS_NONCE_FILE
$ gomplate -d vault=vault:///secret/foo -i '{{ (ds "vault").value }}'
...
```

The file `/tmp/vault-aws-nonce` will be created if it didn't already exist, and further executions of `gomplate` can re-authenticate securely.

[`--datasource`/`-d`]: ../usage/#datasource-d
[`--datasource-header`/`-H`]: ../usage/#datasource-header-h
[`defineDatasource`]: ../functions/data/#definedatasource
[`datasource`]: ../functions/data/#datasource
[`include`]: ../functions/data/#include
[`data.CSV`]: ../functions/data/#data-csv
[`data.JSON`]: ../functions/data/#data-json
[`data.JSONArray`]: ../functions/data/#data-jsonarray
[`data.TOML`]: ../functions/data/#data-toml
[`data.YAML`]: ../functions/data/#data-yaml

[AWS SMP]: https://aws.amazon.com/systems-manager/features#Parameter_Store
[BoltDB]: https://github.com/boltdb/bolt
[HashiCorp Consul]: https://consul.io
[HashiCorp Vault]: https://vaultproject.io
[JSON]: https://json.org
[TOML]: https://github.com/toml-lang/toml
[YAML]: http://yaml.org
[HTTP Content-Type]: https://tools.ietf.org/html/rfc7231#section-3.1.1.1
[URL]: https://tools.ietf.org/html/rfc3986
[AWS SDK for Go]: https://docs.aws.amazon.com/sdk-for-go/api/
