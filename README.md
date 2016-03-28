[![Build Status][circleci-image]][circleci-url]

# gomplate

A [Go template](https://golang.org/pkg/text/template/)-based alternative to [`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html).

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables. Also there are some useful built-in functions that can be used to make templates even more expressive.

## Usage

The usual and most basic usage of `gomplate` is to just replace environment variables. All environment variables are available by referencing `.Env` (or `getenv`) in the template.

The template is read from standard in, and written to standard out.

Use it like this:

```console
$ echo "Hello, {{.Env.USER}}" | gomplate
Hello, hairyhenderson
```

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

Queries AWS to get the region. Returns `unknown` if it can't be determined for some reason.

##### Example

```console
$ echo '{{ ec2region }}' | ./gomplate
us-east-1
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
{{ if eq $u "root" }}You are root!{{else}}You are not root :({{end}}
```

```console
$ gomplate < input.tmpl
You are not root :(
$ sudo gomplate < input.tmpl
You are root!
```

_Note:_ it's important for the `if`/`else`/`end` keywords to appear on the same line, or else `gomplate` will not be able to parse the pipeline properly

## Releasing

Right now the release process is semi-automatic.

1. Create a release tag: `git tag -a v0.0.9 -m "Releasing v0.9.9"`
2. Build binaries: `make build-x`
3. _(optional)_ compress the binaries with `upx` (may not work for all architectures)
4. Create a release in [github](https://github.com/hairyhenderson/gomplate/releases)!

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2016 Dave Henderson

[circleci-image]: https://img.shields.io/circleci/project/hairyhenderson/gomplate.svg?style=flat
[circleci-url]: https://circleci.com/gh/hairyhenderson/gomplate
