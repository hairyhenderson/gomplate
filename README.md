[![Build Status][circleci-image]][circleci-url]
[![Go Report Card][reportcard-image]][reportcard-url]
[![Codebeat Status][codebeat-image]][codebeat-url]
[![Coverage][gocover-image]][gocover-url]
[![Total Downloads][gh-downloads-image]][gh-downloads-url]
[![CII Best Practices][cii-bp-image]][cii-bp-url]

[![hairyhenderson/gomplate on DockerHub][dockerhub-image]][dockerhub-url]
[![DockerHub Stars][dockerhub-stars-image]][dockerhub-url]
[![DockerHub Pulls][dockerhub-pulls-image]][dockerhub-url]
[![DockerHub Image Layers][microbadger-layers-image]][microbadger-url]
[![DockerHub Latest Version ][microbadger-version-image]][microbadger-url]
[![DockerHub Latest Commit][microbadger-commit-image]][microbadger-url]

# gomplate

A [Go template](https://golang.org/pkg/text/template/)-based CLI tool. `gomplate` can be used as an alternative to
[`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html) but also supports
additional template datasources such as: JSON, YAML, AWS EC2 metadata, and
[Hashicorp Vault](https://https://www.vaultproject.io/) secrets.

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables. Also there are some useful built-in functions that can be used to make templates even more expressive.

<!-- TOC depthFrom:1 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [gomplate](#gomplate)
	- [Installing](#installing)
		- [macOS with homebrew](#macos-with-homebrew)
		- [Alpine Linux](#alpine-linux)
		- [use with Docker](#use-with-docker)
		- [manual install](#manual-install)
	- [Usage](#usage)
		- [Commandline Arguments](#commandline-arguments)
			- [`--datasource`/`-d`](#-datasource-d)
			- [Overriding the template delimiters](#overriding-the-template-delimiters)
	- [Syntax](#syntax)
		- [About `.Env`](#about-env)
		- [Built-in functions](#built-in-functions)
			- [`contains`](#contains)
				- [Example](#example)
			- [`getenv`](#getenv)
				- [Example](#example)
			- [`hasPrefix`](#hasprefix)
				- [Example](#example)
			- [`hasSuffix`](#hassuffix)
				- [Example](#example)
			- [`bool`](#bool)
				- [Example](#example)
			- [`slice`](#slice)
				- [Example](#example)
			- [`split`](#split)
				- [Example](#example)
			- [`title`](#title)
				- [Example](#example)
			- [`toLower`](#tolower)
				- [Example](#example)
			- [`toUpper`](#toupper)
				- [Example](#example)
			- [`trim`](#trim)
				- [Example](#example)
			- [`has`](#has)
				- [Example](#example)
			- [`json`](#json)
				- [Example](#example)
			- [`jsonArray`](#jsonarray)
				- [Example](#example)
			- [`yaml`](#yaml)
				- [Example](#example)
			- [`yamlArray`](#yamlarray)
				- [Example](#example)
			- [`toJSON`](#tojson)
				- [Example](#example)
			- [`toYAML`](#toyaml)
				- [Example](#example)
			- [`datasource`](#datasource)
				- [Examples](#examples)
					- [Basic usage](#basic-usage)
					- [Usage with HTTP data](#usage-with-http-data)
					- [Usage with Vault data](#usage-with-vault-data)
			- [`datasourceExists`](#datasourceexists)
			- [`ec2meta`](#ec2meta)
				- [Example](#example)
			- [`ec2dynamic`](#ec2dynamic)
				- [Example](#example)
			- [`ec2region`](#ec2region)
				- [Example](#example)
			- [`ec2tag`](#ec2tag)
				- [Example](#example)
		- [Some more complex examples](#some-more-complex-examples)
			- [Variable assignment and `if`/`else`](#variable-assignment-and-ifelse)
	- [Releasing](#releasing)
	- [License](#license)

<!-- /TOC -->

## Installing

### macOS with homebrew

The simplest method for macOS is to use homebrew:

```console
$ brew tap hairyhenderson/tap
$ brew install gomplate
...
```

### Alpine Linux

Currently, `gomplate` is available in the `testing` repository.

```console
$ echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing/" >> /etc/apk/repositories
$ apk update
$ apk add gomplate
...
```

_Note: the Alpine version of gomplate may lag behind the latest release of gomplate._

### use with Docker

A simple way to get started is with the Docker image.

```console
$ docker run hairyhenderson/gomplate --version
```

Of course, there are some drawbacks - any files to be used for [datasources][]
must be mounted and any environment variables to be used must be passed through:

```console
$ echo 'My voice is my {{.Env.THING}}. {{(datasource "vault").value}}' \
  | docker run -e THING=passport -v /home/me/.vault-token:/root/.vault-token hairyhenderson/gomplate -d vault=vault:///secret/sneakers
My voice is my passport. Verify me.
```

It can be pretty awkward to always type `docker run hairyhenderson/gomplate`,
so this can be made simpler with a shell alias:

```console
$ alias gomplate=docker run hairyhenderson/gomplate
$ gomplate --version
gomplate version 1.2.3
```

### manual install

1. Get the latest `gomplate` for your platform from the [releases](https://github.com/hairyhenderson/gomplate/releases) page
2. Store the downloaded binary somewhere in your path as `gomplate` (or `gomplate.exe`
  on Windows)
3. Make sure it's executable (on Linux/macOS)
3. Test it out with `gomplate --help`!

In other words:

```console
$ curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/<version>/gomplate_<os>-<arch>
$ chmod 755 /usr/local/bin/gomplate
$ gomplate --help
...
```

_Please report any bugs found in the [issue tracker](https://github.com/hairyhenderson/gomplate/issues/)._

## Usage

The usual and most basic usage of `gomplate` is to just replace environment variables. All environment variables are available by referencing `.Env` (or `getenv`) in the template.

The template is read from standard in, and written to standard out.

Use it like this:

```console
$ echo "Hello, {{.Env.USER}}" | gomplate
Hello, hairyhenderson
```

### Commandline Arguments

#### `--datasource`/`-d`

Add a data source in `name=URL` form. Specify multiple times to add multiple sources. The data can then be used by the [`datasource`](#datasource) function.

A few different forms are valid:
- `mydata=file:///tmp/my/file.json`
  - Create a data source named `mydata` which is read from `/tmp/my/file.json`. This form is valid for any file in any path.
- `mydata=file.json`
  - Create a data source named `mydata` which is read from `file.json` (in the current working directory). This form is only valid for files in the current directory.
- `mydata.json`
  - This form infers the name from the file name (without extension). Only valid for files in the current directory.

#### Overriding the template delimiters

Sometimes it's necessary to override the default template delimiters (`{{`/`}}`).
Use `--left-delim`/`--right-delim` or set `$GOMPLATE_LEFT_DELIM`/`$GOMPLATE_RIGHT_DELIM`.

## Syntax

### About `.Env`

You can easily access environment variables with `.Env`, but there's a catch:
if you try to reference an environment variable that doesn't exist, parsing
will fail and `gomplate` will exit with an error condition.

Sometimes, this behaviour is desired; if the output is unusable without certain strings, this is a sure way to know that variables are missing!

If you want different behaviour, try `getenv` (below).

### Built-in functions

In addition to all of the functions and operators that the [Go template](https://golang.org/pkg/text/template/)
language provides (`if`, `else`, `eq`, `and`, `or`, `range`, etc...), there are
some additional functions baked in to `gomplate`:

#### `contains`

Contains reports whether the second string is contained within the first. Equivalent to
[strings.Contains](https://golang.org/pkg/strings#Contains)

##### Example

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

#### `getenv`

Exposes the [os.Getenv](https://golang.org/pkg/os/#Getenv) function.

This is a more forgiving alternative to using `.Env`, since missing keys will
return an empty string.

An optional default value can be given as well.

##### Example

```console
$ echo 'Hello, {{getenv "USER"}}' | gomplate
Hello, hairyhenderson
$ echo 'Hey, {{getenv "FIRSTNAME" "you"}}!' | gomplate
Hey, you!
```

#### `hasPrefix`

Tests whether the string begins with a certain substring. Equivalent to
[strings.HasPrefix](https://golang.org/pkg/strings#HasPrefix)

##### Example

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

#### `hasSuffix`

Tests whether the string ends with a certain substring. Equivalent to
[strings.HasSuffix](https://golang.org/pkg/strings#HasSuffix)

##### Example

_`input.tmpl`:_
```
{{.Env.URL}}{{if not (hasSuffix .Env.URL ":80")}}:80{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
http://example.com:80
```

#### `bool`

Converts a true-ish string to a boolean. Can be used to simplify conditional statements based on environment variables or other text input.

##### Example

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

#### `slice`

Creates a slice. Useful when needing to `range` over a bunch of variables.

##### Example

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

#### `split`

Creates a slice by splitting a string on a given delimiter. Equivalent to
[strings.Split](https://golang.org/pkg/strings#Split)

##### Example

_`input.tmpl`:_
```
{{range split "Bart,Lisa,Maggie"}}
Hello, {{.}}
{{- end}}
```

```console
$ gomplate < input.tmpl
Hello, Bart
Hello, Lisa
Hello, Maggie
```

#### `title`

Convert to title-case. Equivalent to [strings.Title](https://golang.org/pkg/strings/#Title)

##### Example

```console
$ echo '{{title "hello, world!"}}' | gomplate
Hello, World!
```

#### `toLower`

Convert to lower-case. Equivalent to [strings.ToLower](https://golang.org/pkg/strings/#ToLower)

##### Example

```console
$ echo '{{toLower "HELLO, WORLD!"}}' | gomplate
hello, world!
```

#### `toUpper`

Convert to upper-case. Equivalent to [strings.ToUpper](https://golang.org/pkg/strings/#ToUpper)

##### Example

```console
$ echo '{{toUpper "hello, world!"}}' | gomplate
HELLO, WORLD!
```

#### `trim`

Trims a string by removing the given characters from the beginning and end of
the string. Equivalent to [strings.Trim](https://golang.org/pkg/strings/#Trim)

##### Example

_`input.tmpl`:_
```
Hello, {{trim .Env.FOO " "}}!
```

```console
$ FOO="  world " | gomplate < input.tmpl
Hello, world!
```

#### `has`

Has reports whether or not a given object has a property with the given key. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

##### Example

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

#### `json`

Converts a JSON string into an object. Only works for JSON Objects (not Arrays or other valid JSON types). This can be used to access properties of JSON objects.

##### Example

_`input.tmpl`:_
```
Hello {{ (getenv "FOO" | json).hello }}
```

```console
$ export FOO='{"hello":"world"}'
$ gomplate < input.tmpl
Hello world
```

#### `jsonArray`

Converts a JSON string into a slice. Only works for JSON Arrays.

##### Example

_`input.tmpl`:_
```
Hello {{ index (getenv "FOO" | jsonArray) 1 }}
```

```console
$ export FOO='[ "you", "world" ]'
$ gomplate < input.tmpl
Hello world
```

#### `yaml`

Converts a YAML string into an object. Only works for YAML Objects (not Arrays or other valid YAML types). This can be used to access properties of YAML objects.

##### Example

_`input.tmpl`:_
```
Hello {{ (getenv "FOO" | yaml).hello }}
```

```console
$ export FOO='hello: world'
$ gomplate < input.tmpl
Hello world
```

#### `yamlArray`

Converts a YAML string into a slice. Only works for YAML Arrays.

##### Example

_`input.tmpl`:_
```
Hello {{ index (getenv "FOO" | yamlArray) 1 }}
```

```console
$ export FOO='[ "you", "world" ]'
$ gomplate < input.tmpl
Hello world
```

#### `toJSON`

Converts an object to a JSON document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.

##### Example

_This is obviously contrived - `json` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | json).foo | toJSON }}
```

```console
$ gomplate < input.tmpl
{"hello":"world"}
```

#### `toYAML`

Converts an object to a YAML document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.

##### Example

_This is obviously contrived - `json` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | json).foo | toYAML }}
```

```console
$ gomplate < input.tmpl
hello: world

```

#### `datasource`

Parses a given datasource (provided by the [`--datasource/-d`](#--datasource-d) argument).

Currently, `file://`, `http://`, `https://`, and `vault://` URLs are supported.

Currently-supported formats are JSON and YAML.

##### Examples

###### Basic usage

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

###### Usage with HTTP data

```console
$ echo 'Hello there, {{(datasource "foo").headers.Host}}...' | gomplate -d foo=https://httpbin.org/get
Hello there, httpbin.org...
```

###### Usage with Vault data

The special `vault://` URL scheme can be used to retrieve data from [Hashicorp
Vault](https://vaultproject.io). To use this, you must put the Vault server's
URL in the `$VAULT_ADDR` environment variable.

Currently, the [`app-id`](https://www.vaultproject.io/docs/auth/app-id.html)
auth backend is supported, as well as Vault tokens obtained through external
means.

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

#### `datasourceExists`

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

#### `ec2meta`

Queries AWS [EC2 Instance Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `meta-data` path -- for data in the `dynamic` path use `ec2dynamic`.

This only works when running `gomplate` on an EC2 instance. If the EC2 instance metadata API isn't available, the tool will timeout and fail.

##### Example

```console
$ echo '{{ec2meta "instance-id"}}' | gomplate
i-12345678
```

#### `ec2dynamic`

Queries AWS [EC2 Instance Dynamic Metadata](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) for information. This only retrieves data in the `dynamic` path -- for data in the `meta-data` path use `ec2meta`.

This only works when running `gomplate` on an EC2 instance. If the EC2 instance metadata API isn't available, the tool will timeout and fail.

##### Example

```console
$ echo '{{ (ec2dynamic "instance-identity/document" | json).region }}' | ./gomplate
us-east-1
```

#### `ec2region`

Queries AWS to get the region. An optional default can be provided, or returns
`unknown` if it can't be determined for some reason.

##### Example

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

#### `ec2tag`

Queries the AWS EC2 API to find the value of the given [user-defined tag](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html). An optional default
can be provided.

##### Example

```console
$ echo 'This server is in the {{ ec2tag "Account" }} account.' | ./gomplate
foo
$ echo 'I am a {{ ec2tag "classification" "meat popsicle" }}.' | ./gomplate
I am a meat popsicle.
```

### Some more complex examples

#### Variable assignment and `if`/`else`

_`input.tmpl`:_
```
{{ $u := getenv "USER" }}
{{ if eq $u "root" -}}
You are root!
{{- else -}}
You are not root :(
{{- end}}
```

```console
$ gomplate < input.tmpl
You are not root :(
$ sudo gomplate < input.tmpl
You are root!
```

## Releasing

Right now the release process is semi-automatic.

1. Create a release tag: `git tag -a v0.0.9 -m "Releasing v0.9.9" && git push --tags`
2. Build binaries & compress most of them: `make build-release`
3. Create a release in [github](https://github.com/hairyhenderson/gomplate/releases)!

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2016 Dave Henderson

[circleci-image]: https://img.shields.io/circleci/project/hairyhenderson/gomplate.svg?style=flat
[circleci-url]: https://circleci.com/gh/hairyhenderson/gomplate
[reportcard-image]: https://goreportcard.com/badge/github.com/hairyhenderson/gomplate
[reportcard-url]: https://goreportcard.com/report/github.com/hairyhenderson/gomplate
[codebeat-image]: https://codebeat.co/badges/39ed2148-4b86-4d1e-8526-25f60e159ba1
[codebeat-url]: https://codebeat.co/projects/github-com-hairyhenderson-gomplate
[gocover-image]: https://gocover.io/_badge/github.com/hairyhenderson/gomplate
[gocover-url]: https://gocover.io/github.com/hairyhenderson/gomplate
[gh-downloads-image]: https://img.shields.io/github/downloads/hairyhenderson/gomplate/total.svg
[gh-downloads-url]: https://github.com/hairyhenderson/gomplate/releases

[cii-bp-image]: https://bestpractices.coreinfrastructure.org/projects/337/badge
[cii-bp-url]: https://bestpractices.coreinfrastructure.org/projects/337

[dockerhub-image]: https://img.shields.io/badge/docker-ready-blue.svg
[dockerhub-url]: https://hub.docker.com/r/hairyhenderson/gomplate
[dockerhub-stars-image]: https://img.shields.io/docker/stars/hairyhenderson/gomplate.svg
[dockerhub-pulls-image]: https://img.shields.io/docker/pulls/hairyhenderson/gomplate.svg

[microbadger-version-image]: https://images.microbadger.com/badges/version/hairyhenderson/gomplate.svg
[microbadger-layers-image]: https://images.microbadger.com/badges/image/hairyhenderson/gomplate.svg
[microbadger-commit-image]: https://images.microbadger.com/badges/commit/hairyhenderson/gomplate.svg
[microbadger-url]: https://microbadger.com/image/hairyhenderson/gomplate

[![Analytics](https://ga-beacon.appspot.com/UA-82637990-1/gomplate/README.md?pixel)](https://github.com/igrigorik/ga-beacon)
