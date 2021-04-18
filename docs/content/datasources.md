---
title: Datasources
weight: 14
menu: main
---

Datasources are an optional, but central concept in gomplate. While the basic flow of template rendering is taking an input template and rendering it into an output, there often is need to include data from one or more sources external to the template itself.

Some common use-cases include injecting sensitive material like passwords (which should not be stored unencrypted in source-control with the templates), or providing simplified configuration formats that can be fed to a template to provide a much more complex output.

Datasources can be defined with the [`--datasource`/`-d`][] command-line flag or the [`defineDatasource`][] function, and referenced via an _alias_ inside the template, using a function such as [`datasource`][] or [`include`][]. Datasources can additionally be loaded into the [context][] with the [`--context`/`-c`][] command-line flag.

Since datasources are defined separately from the template, the same templates can be used with different datasources and even different datasource types. For example, gomplate could be run on a developer machine with a `file` datasource pointing to a JSON file containing test data, where the same template could be used in a production environment using a `consul` datasource with the real production data.

## URL Format

All datasources are defined with a [URL][]. As a refresher, a URL is made up of the following components:

```pre
  foo://userinfo@example.com:8042/over/there?name=ferret#nose
  \_/   \_______________________/\_________/ \_________/ \__/
   |           |                    |            |        |
scheme     authority               path        query   fragment
```

For our purposes, the _scheme_ and the _path_ components are especially important, though the other components are used by certain datasources for particular purposes.

| component | purpose |
|-----------|---------|
| _scheme_ | Identifies which [datasource](#supported-datasources) to access. All datasources require a scheme (except for `file` when using relative paths), and some datasources allow multiple different schemes to clarify access modes, such as `consul+https` |
| _authority_ | Used only by remote datasources, and can be omitted in some of those cases. Consists of _userinfo_ (`user:pass`), _host_, and _port_. |
| _path_ | Can be omitted, but usually used as the basis of the locator for the datasource. If the path ends with a `/` character, [directory](#directory-datasources) semantics are used. |
| _query_ | Used rarely for datasources where information must be provided in order to get a reasonable reply (such as generating dynamic secrets with Vault), or for [overriding MIME types](#overriding-mime-types) |
| _fragment_ | Used rarely for accessing a subset of the given path (such as a bucket name in a BoltDB database) |

### Opaque URIs

For some datasources, such as the [`merge`](#using-merge-datasources), [`aws+sm`](#using-aws-sm-datasources), and [`aws+smp`](#using-aws-smp-datasources) schemes, opaque URIs can be used (rather than a hierarchical URL):

```pre
scheme                   path        query   fragment
   |   _____________________|__   _______|_   _|
  / \ /                        \ /         \ /  \
  urn:example:animal:ferret:nose?name=ferret#nose
```

The semantics of the different URI components are essentially the same as for hierarchical URLs (see above),
but the _path_ component may not start with a `/` character. In gomplate's usage, opaque URIs sometimes contain
characters such as `|`, which require escaping with most shells. You may need to surround the datasource definition
in quotes, or use the `\` escape character.

## Supported datasources

Gomplate supports a number of datasources, each specified with a particular URL scheme. The table below describes these datasources. The names in the _Type_ column link to further documentation for each specific datasource.

| Type | URL Scheme(s) | Description |
|------|---------------|-------------|
| [AWS Systems Manager Parameter Store](#using-aws-smp-datasources) | `aws+smp` | [AWS Systems Manager Parameter Store][AWS SMP] is a hierarchically-organized key/value store which allows storage of text, lists, or encrypted secrets for retrieval by AWS resources |
| [AWS Secrets Manager](#using-aws-sm-datasource) | `aws+sm` | [AWS Secrets Manager][] helps you protect secrets needed to access your applications, services, and IT resources. |
| [Amazon S3](#using-s3-datasources) | `s3` | [Amazon S3][] is a popular object storage service. |
| [BoltDB](#using-boltdb-datasources) | `boltdb` | [BoltDB][] is a simple local key/value store used by many Go tools |
| [Consul](#using-consul-datasources) | `consul`, `consul+http`, `consul+https` | [HashiCorp Consul][] provides (among many other features) a key/value store |
| [Environment](#using-env-datasources) | `env` | Environment variables can be used as datasources - useful for testing |
| [File](#using-file-datasources) | `file` | Files can be read in any of the [supported formats](#mime-types), including by piping through standard input (`Stdin`). [Directories](#directory-datasources) are also supported. |
| [Git](#using-git-datasources) | `git`, `git+file`, `git+http`, `git+https`, `git+ssh` | Files can be read from a local or remote git repository, at specific branches or tags. [Directory semantics](#directory-datasources) are also supported. |
| [Google Cloud Storage](#using-google-cloud-storage-gs-datasources) | `gs` | [Google Cloud Storage][] is the object storage service available on GCP, comparable to AWS S3. |
| [HTTP](#using-http-datasources) | `http`, `https` | Data can be sourced from HTTP/HTTPS sites in many different formats. Arbitrary HTTP headers can be set with the [`--datasource-header`/`-H`][] flag |
| [Merged Datasources](#using-merge-datasources) | `merge` | Merge two or more datasources together to produce the final value - useful for resolving defaults. Uses [`coll.Merge`][] for merging. |
| [Stdin](#using-stdin-datasources) | `stdin` | A special case of the `file` datasource; allows piping through standard input (`Stdin`) |
| [Vault](#using-vault-datasources) | `vault`, `vault+http`, `vault+https` | [HashiCorp Vault][] is an industry-leading open-source secret management tool. [List support](#directory-datasources) is also available. |

## Directory Datasources

When the _path_ component of the URL ends with a `/` character, the datasource is read with _directory_ semantics. Not all datasource types support this, and for those that don't support the notion of a directory, the behaviour is currently undefined. See each documentation section for details.
 
Currently the following datasources support directory semantics:

- [File](#using-file-datasources)
- [Vault](#using-vault-datasources) - translates to Vault's [LIST](https://www.vaultproject.io/api/index.html#reading-writing-and-listing-secrets) method
- [Consul](#using-consul-datasources)
When accessing a directory datasource, an array of key names is returned, and can be iterated through to access each individual value contained within.
- [AWS S3](#using-s3-datasources)
- [Google Cloud Storage](#using-google-cloud-storage-gs-datasources)
- [Git](#using-git-datasources) 
- [AWS Systems Manager Parameter Store](#using-aws-smp-datasources)

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

Gomplate will read and parse a number of data formats. The appropriate type will be set automatically, if possible, either based on file extension (for the `file`, `http`, `gs`, and `s3` datasources), or the [HTTP Content-Type][] header, if available. If an unsupported type is detected, gomplate will exit with an error.

These are the supported types:

| Format | MIME Type | Extension(s) | Notes |
|--------|-----------|-------|------|
| CSV | `text/csv` | `.csv` | Uses the [`data.CSV`][] function to present the file as a 2-dimensional row-first string array |
| JSON | `application/json` | `.json` | [JSON][] _objects_ are assumed, but will support arrays as well. Other values are not parsed with this type. Uses the [`data.JSON`][] function for parsing. [EJSON][] (encrypted JSON) is supported and will be decrypted. |
| JSON Array | `application/array+json` | | A special type for parsing datasources containing just JSON arrays. Uses the [`data.JSONArray`][] function for parsing |
| Plain Text | `text/plain` | | Unstructured, and as such only intended for use with the [`include`][] function |
| TOML | `application/toml` | `.toml` | Parses [TOML][] with the [`data.TOML`][] function |
| YAML | `application/yaml` | `.yml`, `.yaml` | Parses [YAML][] with the [`data.YAML`][] function |
| [.env](#the-env-file-format) | `application/x-env` | `.env` | Basically just a file of `key=value` pairs separated by newlines, usually intended for sourcing into a shell. Common in [Docker Compose](https://docs.docker.com/compose/env-file/), [Ruby](https://github.com/bkeepers/dotenv), and [Node.js](https://github.com/motdotla/dotenv) applications. See [below](#the-env-file-format) for more information. |

### Overriding MIME Types

On occasion it's necessary to override the detected (via file extension or `Content-Type` header) MIME type. To accomplish this, gomplate supports a `type` query string parameter on datasource URLs. This can contain the same value as a standard [HTTP Content-Type][] header.

For example, to force a file named `data.txt` to be parsed as a JSON document:

```console
$ echo '{"foo": "bar"}' > /tmp/data.txt
$ gomplate -d data=file:///tmp/data.txt?type=application/json -i '{{ (ds "data").foo }}'
bar
```

### The `.env` file format

Many applications and frameworks support the use of a ".env" file for providing environment variables. It can also be considerd a simple key/value file format, and as such can be used as a datasource in gomplate.

To [override](#overriding-mime-types), use the unregistered `application/x-env` MIME type.

Here's a sample explaining the syntax:

```bash
FOO=a regular unquoted value
export BAR=another value, exports are ignored

# comments are totally ignored, as are blank lines
FOO.BAR = "values can be double-quoted, and\tshell\nescapes are supported"

BAZ="variable expansion: ${FOO}"
QUX='single quotes ignore $variables and newlines'
```

The [`github.com/joho/godotenv`](https://github.com/joho/godotenv) package is used for parsing - see the full details there.


## Using `aws+smp` datasources

The `aws+smp://` scheme can be used to retrieve data from the [AWS Systems Manager](https://aws.amazon.com/systems-manager/) (née AWS EC2 Simple Systems Manager) [Parameter Store](https://aws.amazon.com/systems-manager/features/#Parameter_Store). This hierarchically organized key/value store allows you to store text, lists or encrypted secrets for easy retrieval by AWS resources. See [the AWS Systems Manager documentation](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-su-create.html#sysman-paramstore-su-create-about) for details on creating these parameters.

You must grant `gomplate` permission via IAM credentials for the [`ssm:GetParameter` action](https://docs.aws.amazon.com/systems-manager/latest/userguide/auth-and-access-control-permissions-reference.html). <!-- List support further requires `ssm.GetParameters` permissions. -->

See details on how to configure gomplate's AWS support in [_Configuring AWS_](../functions/aws/#configuring-aws).

### URL Considerations

The _scheme_ and _path_ URL components are used by this datasource. This may be an _opaque_ URI instead of an URL, when the key does not begin with a `/` character (e.g. `aws+smp:myparam`).

- the _scheme_ must be `aws+smp`
- the _path_ component is used to specify the path to the parameter (this may be a hierarchical path beginning with `/`, or an opaque path). [Directory](#directory-datasources) semantics are available when the path ends with a `/` character.

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
- `myparameter` - `bar`

```console
$ echo '{{ ds "foo" }}' | gomplate -d foo=aws+smp:///foo/first/password
map[Name:/foo/first/password Type:SecureString Value:super-secret Version:1]

$ echo '{{ (ds "foo").Value }}' | gomplate -d foo=aws+smp:///foo/first/password
super-secret

$ echo '{{ (ds "foo" "/foo/first/others").Value }}' | gomplate -d foo=aws+smp:
Bill,Ben

$ echo '{{ (ds "foo" "/second/p1").Value }}' | gomplate -d foo=aws+smp:///foo/
aaa

$ gomplate -d foo=aws+smp:///foo/first/ -i '{{ range (ds "foo") }}
{{ . }}: {{ (ds "foo" .).Value }}
{{- end }}'
others: Bill,Ben
password: super-secret

$ gomplate -d foo=aws+smp:myparameter -i '{{ (ds "foo").Value }}
bar
```

## Using `aws+sm` datasource

### URL Considerations

For `aws+sm`, only the _scheme_ and _path_ components are necessary to be defined. This may be an _opaque_ URI instead of an URL, when the key does not begin with a `/` character (e.g. `aws+sm:myparam`).

- the _scheme_ must be `aws+sm`
- the _path_ component is used to specify the path to the secret (this may be a hierarchical path beginning with `/`, or an opaque path)

### Output

The output will be the SecretString from the `GetSecretValueOutput` object from the [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/api/service/secretsmanager/#GetSecretValueOutput)

### Examples

Given your [AWS account's Secret Manager](https://eu-central-1.console.aws.amazon.com/secretsmanager/home?region=eu-central-1#/listSecrets) has the following data:

- `/foo/bar/password` - `super-secret`
- `mysecret` - `bar`

```console
$ echo '{{ (ds "foo") }}' | gomplate -d foo=aws+sm:///foo/bar/password
super-secret

$ echo '{{ (ds "foo" "/foo/bar/password") }}' | gomplate -d foo=aws+sm:
super-secret

$ echo '{{ (ds "foo" "/bar/password") }}' | gomplate -d foo=aws+sm:///foo/
super-secret

$ echo '{{ (ds "foo") }}' | gomplate -d foo=aws+sm:mysecret
bar
```

## Using `s3` datasources

### URL Considerations

The _scheme_, _authority_, _path_, and _query_ URL components are used by this datasource.

- the _scheme_ must be `s3`
- the _authority_ component is used to specify the s3 bucket name
- the _path_ component is used to specify the path to the object. [Directory](#directory-datasources) semantics are available when the path ends with a `/` character.
- the _query_ component can be used to provide parameters to configure the connection:
  - `region`: The AWS region for requests. Defaults to the value from the `AWS_REGION` or `AWS_DEFAULT_REGION` environment variables, or the EC2 region if run in AWS EC2.
  - `endpoint`: The endpoint (`hostname`, `hostname:port`, or fully qualified URI). Useful for using a different S3-compatible object storage server. You can also set the `AWS_S3_ENDPOINT` environment variable.
  - `s3ForcePathStyle`: A value of `true` forces use of the deprecated "path-style" access. This is necessary for some S3-compatible object storage servers.
  - `disableSSL`: A value of `true` disables SSL when sending requests. Use only for test scenarios!
  - `type`: can be used to [override the MIME type](#overriding-mime-types)

#### URL Examples

Here are a few examples to help explain `s3` URLs:

- `s3://mybucket/config/file.json`
  - the bucket region will be inferred, and the blob `config/file.json` in the `mybucket` bucket will be located in Amazon S3.
- `s3://mybucket/`
  - The contents of the bucket `mybucket` will be listed into an array. Note that only the last portion of the path (the file name) will be listed.
- `s3://mybucket/config/file?region=eu-west-1`
  - same as the first example, except the bucket's region is overridden to `eu-west-1`
  - the lack of file extension means that file will be parsed according to the file's `Content-Type` metadata
- `s3://mybucket/config/file.json?endpoint=localhost:5432&disableSSL=true&s3ForcePathStyle=true`
  - this example is typical of a scenario where an S3-compatible server (such as [Minio][], [Zenko CloudServer][], or testing-focused servers such as [gofakes3][])
  - the endpoint is overridden to be a server running on localhost
  - encryption is disabled since the endpoint is local
  - "path-style" access is used - this is typical for local servers, or scenarios where modifying DNS is impossible or impractical

### Output

The output will be the object contents, parsed based on the discovered [MIME type](#mime-types).

### Examples

Given the S3 bucket named `my-bucket` has the following objects:

- `foo/bar.json` - `{"hello": "world"}`
- `foo/baz.txt` - `hello world`

```console
$ gomplate -c foo=s3://my-bucket/foo/bar.json -i 'Hello {{ .foo.hello }}'
Hello world

$ gomplate -c foo=s3://my-bucket/foo/ -i 'my-bucket/foo contains:{{ range .foo }}{{ print "\n" . }}{{ end }}'
my-bucket/foo contains:
bar.json
baz.txt

$ gomplate -c foo=s3://my-bucket/foo/bar.json?region=eu-west-1 -i 'Hello {{ .foo.hello }}'
Hello world

$ gomplate -c foo=s3://my-bucket/foo/bar.json?region=eu-west-1&
endpoint=my-test-site& -i 'Hello {{ .foo.hello }}'
Hello world

$ gomplate -d bucket=s3://my-bucket/?region=eu-west-1&endpoint=my-test-site& -i 'Hello {{ (ds "bucket" "/foo/bar.json").hello }}'
Hello world
```

## Using `boltdb` datasources

[BoltDB][] is a simple local key/value store used by many Go tools. The `boltdb://` scheme can be used to access values stored in a BoltDB database file. The full path is provided in the URL, and the bucket name can be specified using a URL fragment (e.g. `boltdb:///tmp/database.db#bucket`).

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

$ gomplate -d homedir=env:///HOME -i '{{ file.IsDir (ds "homedir") }}'
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

## Using `git` datasources

The `git` datasource type provides access to files in any of the [supported formats](#mime-types) hosted in local or remote git repositories. [Directory datasource](#directory-datasources) semantics are supported.

Remote repositories can be accessed by SSH, HTTP(S), and Git protocols.

Note that this datasource accesses the git state, and so for local filesystem repositories, any files not committed to a branch (i.e. "dirty" or modified files) will not be visible.

### URL Considerations

The _scheme_, _authority_ (with _userinfo_), _path_, and _fragment_ are used, and the _query_ component can be used to [override the MIME type](#overriding-mime-types).

- the _scheme_ may be one of these values:
  - `git`: uses the [classic Git protocol](https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols#_the_git_protocol) (as served by `git daemon`)
  - `git+file`: uses the local filesystem (repo can be bare or not)
  - `git+http`, `git+https`: uses the [Smart HTTP protocol](https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols#_the_http_protocols)
  - `git+ssh`: uses the [SSH protocol](https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols#_the_ssh_protocol)
- the _authority_ component points to the remote git server hostname (and optional port, if applicable). The _userinfo_ subcomponent can be used for authenticated datasources like `git+https` and `git+ssh`.
- the _path_ component is a composite of the path to the repository, and the path to the file or directory being referenced within. The `//` sequence (double forward-slash) is used to separate the repository from the path. If no `//` is present in the URL, the datasource will point to the root directory of the repository.
- the _fragment_ component can be used to specify which branch or tag to reference. By default, the repository's default branch will be chosen.
  - branches can be referenced by short name or by the long form. Valid fragments are `#master`, `#develop`, `#refs/heads/mybranch`, etc...
  - tags must use the long form prefixed by `refs/tags/`, i.e. `#refs/tags/v1` for the `v1` tag

### Authentication

The `git` and `git+file` schemes are always unauthenticated, `git+http`/`git+https` can _optionally_ be authenticated, and `git+ssh` _must_ be authenticated.

Authenticating with both HTTP and SSH requires the user to be set (like `git+ssh://user@example.com`), but the credentials vary otherwise.

#### HTTP(S) Authentication

Note that because HTTP connections are unencypted, and HTTP authentication is performed with headers, it is strongly recommended to _only_ use HTTPS (`git+https`) connections when accessing authenticated repositories.

##### Basic Auth

The most common form. The password can be specified as part of the URL, or provided through the `GIT_HTTP_PASSWORD` environment variable, or in a file referenced by the `GIT_HTTP_PASSWORD_FILE` environment variable.

For authenticating with GitHub, Bitbucket, GitLab and other popular git hosts, use this method with a _personal access token_, and the user set to `git`.

##### Token Auth

Some servers require the use of a bearer token. To use this method, a user is _not_ required, and the token must be set in the `GIT_HTTP_TOKEN` environment variable, or in a file referenced by the `GIT_HTTP_TOKEN_FILE` environment variable.

#### SSH Authentication

Only public key based authentication is supported for `git+ssh` connections. The key can be provided directly, or via the SSH Agent (or Pageant on Windows).

To provide a key directly, set the `GIT_SSH_KEY` to the contents of the key, or point `GIT_SSH_KEY_FILE` to a file containing the key. Because the file may contain newline characters that may be difficult to provide in an environment variable, it can also be Base64-encoded.

If neither `GIT_SSH_KEY` nor `GIT_SSH_KEY_FILE` are set, gomplate will attempt to use the SSH Agent.

**Note:** password-protected SSH keys are currently not supported. If you have a password-protected key, use the SSH Agent.

### Examples

Accessing a file in a publicly-readable GitHub repository:
```console
$ gomplate -c doc=git+https://github.com/hairyhenderson/gomplate//docs-src/content/functions/env.yml -i 'namespace is: {{ .doc.ns }}'
namespace is: env
```

Accessing a file from a local repo (using arguments):
```console
$ gomplate -d which=git+file:///repos/go-which -i 'GOPATH on Windows is {{ (datasource "which" "//appveyor.yml").environment.GOPATH }}'
GOPATH on Windows is c:\gopath
```

Accessing a directory at a specific tag:
```console
$ gomplate -d 'cmd=git+https://github.com/hairyhenderson/go-which//cmd/which#refs/tags/v0.1.0' -i '{{ ds "cmd" }}'
[main.go]
```

Authenticating with the SSH Agent
```console
$ gomplate -d 'which=git+ssh://git@github.com/hairyhenderson/go-which' -i '{{ len (ds "which") }}'
18
```

Using arguments to specify different repos
```console
$ gomplate -d 'hairyhenderson=git+https://github.com/hairyhenderson' -i '{{ (ds "hairyhenderson" "/gomplate//docs-src/content/functions/env.yml").ns }}'
env
```

## Using Google Cloud Storage (`gs`) datasources

### URL Considerations

The _scheme_, _authority_, _path_, and _query_ URL components are used by this datasource.

- the _scheme_ must be `gs`
- the _authority_ component is used to specify the bucket name
- the _path_ component is used to specify the path to the object. [Directory](#directory-datasources) semantics are available when the path ends with a `/` character.
- the _query_ component can be used to provide parameters to configure the connection:
  - `access_id`: (optional) Usually unnecessary. Sets the GoogleAccessID (see https://godoc.org/cloud.google.com/go/storage#SignedURLOptions)
  - `private_key_path`: (optional) Usually unnecessary. Sets the path to the Google service account private key (see https://godoc.org/cloud.google.com/go/storage#SignedURLOptions)
  - `type`: can be used to [override the MIME type](#overriding-mime-types)

### Authentication

All `gs` datasources need credentials, provided by the `GOOGLE_APPLICATION_CREDENTIALS` environment variable. This should point to an authentication configuration JSON file.

See Google Cloud's [Getting Started with Authentication](https://cloud.google.com/docs/authentication/getting-started) documentation for details.

### Output

The output will be the object contents, parsed based on the discovered [MIME type](#mime-types).

### Examples

Given the bucket named `my-bucket` has the following objects:

- `foo/bar.json` - `{"hello": "world"}`
- `foo/baz.txt` - `hello world`

```console
$ gomplate -c foo=gs://my-bucket/foo/bar.json -i 'Hello {{ .foo.hello }}'
Hello world

$ gomplate -c foo=gs://my-bucket/foo/ -i 'my-bucket/foo contains:{{ range .foo }}{{ print "\n" . }}{{ end }}'
my-bucket/foo contains:
bar.json
baz.txt
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

## Using `merge` datasources

The `merge` scheme can be used to merge two or more other datasources together.

`merge:` uses an [_opaque_ URI](#opaque-uris) format, where the _path_ component
is a list of datasource aliases or URLs, separated by the `|` character. The
datasources are read and merged together from right to left (i.e. the left-most
datasource values _override_ those to the right).

Multiple different formats can be mixed, as long as they produce maps with string
keys as their data type.

The [`coll.Merge`][] function is used to perform the merge operation.

### Merging separately-defined datasources

Consider this example:

```console
$ gomplate -d "foo=merge:foo|bar|baz" -d foo=... -d bar=... -d baz=... ...
```

This will read the `foo`, `bar`, and `baz` datasources (which must be otherwise
defined), and then overlay `bar`'s values on top of `baz`'s, then `foo`'s values
on top of those.

The disadvantage with this option is verbosity, but the advantage is that the
individual datasources can still be referenced.

### Merging datasources defined in-line

Here's an example using URLs instead of aliases:

```console
$ gomplate -d "foo=merge:./config/main.yaml|http://example.com/defaults.json" ...
```

This has the advantage of being slightly less verbose. Note that relative URLs
in a subdirectory are supported in this context, as well as any other supported
datasource URL.

A caveat to defining datasources in-line is that the _query_ and _fragment_
components of the URI are interpreted as part of the `merge:` URI. To merge
datasources with query strings or fragments, define separate sources first and
use the aliases. Similarly, extra HTTP headers can only be defined for separately-
defined datasources.

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
[`--context`/`-c`]: ../usage/#context-c
[context]: ../syntax/#the-context
[`--datasource-header`/`-H`]: ../usage/#datasource-header-h
[`defineDatasource`]: ../functions/data/#definedatasource
[`datasource`]: ../functions/data/#datasource
[`include`]: ../functions/data/#include
[`data.CSV`]: ../functions/data/#data-csv
[`data.JSON`]: ../functions/data/#data-json
[EJSON]: ../functions/data/#encrypted-json-support-ejson
[`data.JSONArray`]: ../functions/data/#data-jsonarray
[`data.TOML`]: ../functions/data/#data-toml
[`data.YAML`]: ../functions/data/#data-yaml
[`coll.Merge`]: ../functions/coll/#coll-merge

[AWS SMP]: https://aws.amazon.com/systems-manager/features#Parameter_Store
[AWS Secrets Manager]: https://aws.amazon.com/secrets-manager
[BoltDB]: https://pkg.go.dev/go.etcd.io/bbolt
[HashiCorp Consul]: https://consul.io
[HashiCorp Vault]: https://vaultproject.io
[JSON]: https://json.org
[TOML]: https://github.com/toml-lang/toml
[YAML]: http://yaml.org
[HTTP Content-Type]: https://tools.ietf.org/html/rfc7231#section-3.1.1.1
[URL]: https://tools.ietf.org/html/rfc3986
[AWS SDK for Go]: https://docs.aws.amazon.com/sdk-for-go/api/
[Amazon S3]: https://aws.amazon.com/s3/
[Google Cloud Storage]: https://cloud.google.com/storage/

[Minio]: https://min.io
[Zenko CloudServer]: https://www.zenko.io/cloudserver/
[gofakes3]: https://github.com/johannesboyne/gofakes3
