[![Build Status][circleci-image]][circleci-url]

# gomplate

A simple [Go template](https://golang.org/pkg/text/template/)-based alternative to [`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html).

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables.

## Usage

At the moment, `gomplate` just replaces environment variables. All environment variables are available by referencing `.Env` in the template.

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

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2016 Dave Henderson

[circleci-image]: https://img.shields.io/circleci/project/hairyhenderson/gomplate.svg?style=flat
[circleci-url]: https://circleci.com/gh/hairyhenderson/gomplate
