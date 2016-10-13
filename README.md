[![Build Status][circleci-image]][circleci-url]
[![Go Report Card][reportcard-image]][reportcard-url]
[![Codebeat Status][codebeat-image]][codebeat-url]
[![Coverage][gocover-image]][gocover-url]
[![Total Downloads][gh-downloads-image]][gh-downloads-url]
[![CII Best Practices][cii-bp-image]][cii-bp-url]

# gomplate

A [Go template](https://golang.org/pkg/text/template/)-based alternative to [`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html).

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables. Also there are some useful built-in functions that can be used to make templates even more expressive.

## Installing

### macOS with homebrew

The simplest method for macOS is to use homebrew:

```console
$ brew tap hairyhenderson/tap
$ brew install gomplate
...
```

### manual install

1. Get the latest `gomplate` for your platform from the [releases](https://github.com/hairyhenderson/gomplate/releases) page
2. Store the downloaded binary somewhere in your path as `gomplate` (or `gomplate.exe`
  on Windows)
3. Make sure it's executable (on Linux/macOS)
3. Test it out with `gomplate --help`!

In other words:

```console
$ curl -o /usr/local/bin/gomplate https://github.com/hairyhenderson/gomplate/releases/download/<version>/gomplate_<os>-<arch>
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

## Syntax

#### About `.Env`

You can easily access environment variables with `.Env`, but there's a catch:
if you try to reference an environment variable that doesn't exist, parsing
will fail and `gomplate` will exit with an error condition.

Sometimes, this behaviour is desired; if the output is unusable without certain strings, this is a sure way to know that variables are missing!

If you want different behaviour, try `getenv` (below).

### Built-in functions

In addition to all of the functions and operators that the [Go template](https://golang.org/pkg/text/template/)
language provides (`if`, `else`, `eq`, `and`, `or`, `range`, etc...), there are
some additional functions baked in to `gomplate`:

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

#### `datasource`

Parses a given datasource (provided by the [`--datasource/-d`](#--datasource-d) argument).

Currently, `file://`, `http://` and `https://` URLs are supported.

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

#### Example

```console
$ echo 'This server is in the {{ ec2tag "Account" }} account.' | ./gomplate
foo
$ echo 'I am a {{ ec2tag "classification" "meat popsicle" }}.' | ./gomplate
I am a meat popsicle.
```

### Some more complex examples

##### Variable assignment and `if`/`else`

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
_Note:_ You can use the `{{-` and `-}}` to trim whitespace before and after a multiple line `if`/`else`/`end` statement. See the [Text and spaces](https://golang.org/pkg/text/template/) section of the go template language. Alternatively, you can use a single line `if`/`else`/`end` statement:

```
{{ if eq $u "root" }}You are root!{{else}}You are not root :({{end}}
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

[![Analytics](https://ga-beacon.appspot.com/UA-82637990-1/gomplate/README.md?pixel)](https://github.com/igrigorik/ga-beacon)
