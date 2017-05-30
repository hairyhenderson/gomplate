---
title: Built-in functions
weight: 0
---

In addition to all of the functions and operators that the [Go template](https://golang.org/pkg/text/template/)
language provides (`if`, `else`, `eq`, `and`, `or`, `range`, etc...), there are
some additional functions baked in to `gomplate`:

## `contains`

Contains reports whether the second string is contained within the first. Equivalent to
[strings.Contains](https://golang.org/pkg/strings#Contains)

### Example

_`input.tmpl`:_
```
{{if contains .Env.FOO "f"}}yes{{else}}no{{end}}
```

```console
$ FOO=foo gomplate < input.tmpl
yes
$ FOO=bar gomplate < input.tmpl
no
```

## `getenv`

Exposes the [os.Getenv](https://golang.org/pkg/os/#Getenv) function.

This is a more forgiving alternative to using `.Env`, since missing keys will
return an empty string.

An optional default value can be given as well.

#### Example

```console
$ gomplate -i 'Hello, {{getenv "USER"}}'
Hello, hairyhenderson
$ gomplate -i 'Hey, {{getenv "FIRSTNAME" "you"}}!'
Hey, you!
```

## `hasPrefix`

Tests whether the string begins with a certain substring. Equivalent to
[strings.HasPrefix](https://golang.org/pkg/strings#HasPrefix)

#### Example

_`input.tmpl`:_
```
{{if hasPrefix .Env.URL "https"}}foo{{else}}bar{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
bar
$ URL=https://example.com gomplate < input.tmpl
foo
```

## `hasSuffix`

Tests whether the string ends with a certain substring. Equivalent to
[strings.HasSuffix](https://golang.org/pkg/strings#HasSuffix)

#### Example

_`input.tmpl`:_
```
{{.Env.URL}}{{if not (hasSuffix .Env.URL ":80")}}:80{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
http://example.com:80
```

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

## `split`

Creates a slice by splitting a string on a given delimiter. Equivalent to
[strings.Split](https://golang.org/pkg/strings#Split)

#### Example

```console
$ gomplate -i '{{range split "Bart,Lisa,Maggie" ","}}Hello, {{.}}{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `splitN`

Creates a slice by splitting a string on a given delimiter. The count determines
the number of substrings to return. Equivalent to [strings.SplitN](https://golang.org/pkg/strings#SplitN)

#### Example

```console
$ gomplate -i '{{ range splitN "foo:bar:baz" ":" 2 }}{{.}}{{end}}'
foo
bar:baz
```
## `replaceAll`

Replaces all occurrences of a given string with another.

#### Example

```console
$ gomplate -i '{{ replaceAll "." "-" "172.21.1.42" }}'
172-21-1-42
```

#### Example (with pipeline)

```console
$ gomplate -i '{{ "172.21.1.42" | replaceAll "." "-" }}'
172-21-1-42
```

## `title`

Convert to title-case. Equivalent to [strings.Title](https://golang.org/pkg/strings/#Title)

#### Example

```console
$ echo '{{title "hello, world!"}}' | gomplate
Hello, World!
```

## `toLower`

Convert to lower-case. Equivalent to [strings.ToLower](https://golang.org/pkg/strings/#ToLower)

#### Example

```console
$ echo '{{toLower "HELLO, WORLD!"}}' | gomplate
hello, world!
```

## `toUpper`

Convert to upper-case. Equivalent to [strings.ToUpper](https://golang.org/pkg/strings/#ToUpper)

#### Example

```console
$ echo '{{toUpper "hello, world!"}}' | gomplate
HELLO, WORLD!
```

## `trim`

Trims a string by removing the given characters from the beginning and end of
the string. Equivalent to [strings.Trim](https://golang.org/pkg/strings/#Trim)

#### Example

_`input.tmpl`:_
```
Hello, {{trim .Env.FOO " "}}!
```

```console
$ FOO="  world " | gomplate < input.tmpl
Hello, world!
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

## `indent`

Indents a given string with the given indentation pattern. If the input string has multiple lines, each line will be indented.

#### Example

This function can be especially useful when adding YAML snippets into other YAML documents, where indentation is important:

_`input.tmpl`:_
```
foo:
{{ `{"bar": {"baz": 2}}` | json | toYAML | indent "  " }}
```

```console
$ gomplate -f input.tmpl
foo:
  bar:
    baz: 2
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

Currently-supported formats are JSON, YAML, and CSV.

#### Examples

##### Basic usage

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

##### Usage with HTTP data

```console
$ echo 'Hello there, {{(datasource "foo").headers.Host}}...' | gomplate -d foo=https://httpbin.org/get
Hello there, httpbin.org...
```

Additional headers can be provided with the `--datasource-header`/`-H` option:

```console
$ gomplate -d foo=https://httpbin.org/get -H 'foo=Foo: bar' -i '{{(datasource "foo").headers.Foo}}'
bar
```

##### Usage with Vault data

The special `vault://` URL scheme can be used to retrieve data from [Hashicorp
Vault](https://vaultproject.io). To use this, you must put the Vault server's
URL in the `$VAULT_ADDR` environment variable.

This table describes the currently-supported authentication mechanisms and how to use them, in order of precedence:

| auth backend | configuration |
| ---: |--|
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

## `ec2meta`

Queries AWS [EC2 Instance Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `meta-data` path -- for data in the `dynamic` path use `ec2dynamic`.

This only works when running `gomplate` on an EC2 instance. If the EC2 instance metadata API isn't available, the tool will timeout and fail.

#### Example

```console
$ echo '{{ec2meta "instance-id"}}' | gomplate
i-12345678
```

## `ec2dynamic`

Queries AWS [EC2 Instance Dynamic Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `dynamic` path -- for data in the `meta-data` path use `ec2meta`.

This only works when running `gomplate` on an EC2 instance. If the EC2 instance metadata API isn't available, the tool will timeout and fail.

#### Example

```console
$ echo '{{ (ec2dynamic "instance-identity/document" | json).region }}' | ./gomplate
us-east-1
```

## `ec2region`

Queries AWS to get the region. An optional default can be provided, or returns
`unknown` if it can't be determined for some reason.

#### Example

_In EC2_
```console
$ echo '{{ ec2region }}' | ./gomplate
us-east-1
```
_Not in EC2_
```console
$ echo '{{ ec2region }}' | ./gomplate
unknown
$ echo '{{ ec2region "foo" }}' | ./gomplate
foo
```

## `ec2tag`

Queries the AWS EC2 API to find the value of the given [user-defined tag](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html). An optional default
can be provided.

#### Example

```console
$ echo 'This server is in the {{ ec2tag "Account" }} account.' | ./gomplate
foo
$ echo 'I am a {{ ec2tag "classification" "meat popsicle" }}.' | ./gomplate
I am a meat popsicle.
```
