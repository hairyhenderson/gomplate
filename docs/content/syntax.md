---
title: Syntax
weight: 12
menu: main
---

Gomplate uses the syntax understood by the Go language's [`text/template`][]
package. This page documents some of that syntax, but see [the docs][`text/template`]
for full details.

## The basics

Templates are just regular text, with special actions delimited by `{{` and `}}` markers. Consider the following template:

```
Hello, {{ print "World" }}!
```

If you render this template, it will produce the following output:

```
Hello, World!
```

This is obviously a contrived example, and you would likely never see this in
_real life_, but this conveys the basics, which is that _actions_ are  delimited
by `{{` and `}}`, and are replaced with their output (if any) when the template
is rendered.

## Multi-line templates

By default, every line containing an action will render a newline. For example, the action block below:

```
{{ range slice "Foo" "bar" "baz" }}
Hello, {{ . }}!
{{ end }}
```

will produce the output below:

```

Hello, Foo!

Hello,  bar!

Hello,  baz!

```

This might not be desirable.

You can use [Golang template syntax](https://golang.org/pkg/text/template/#hdr-Text_and_spaces) to fix this. Leading newlines (i.e. newlines that come before the action) can be suppressed by placing a minus sign in front of the first set of delimiters (`{{`). Putting the minus sign behind the trailing set of delimiters (`}}`) will suppress the newline _after_ the action. You can do both to suppress newlines entirely on that line.

Placing the minus sign within the context (i.e. inside of `{{.}}`) has no effect.

Here are a few examples.

### Suppressing leading newlines

```
{{- range slice "Foo" "bar" "baz" }}
Hello, {{ . }}!
{{- end }}
```

will produce this:

```

Hello, Foo!
Hello,  bar!
Hello,  baz!
```

### Suppressing trailling newlines

This code:

```
{{ range slice "Foo" "bar" "baz" -}}
Hello, {{ . }}!
{{ end -}}
```

yields this:

```
Hello, Foo!
Hello,  bar!
Hello,  baz!
```

### Suppressing newlines altogether

This code:

```
{{- range slice "Foo" "bar" "baz" -}}
Hello, {{ . }}!
{{- end -}}
```

Produces:

```
Hello, Foo!Hello,  bar!Hello,  baz!
```


## Variables

The result of an action can be assigned to a _variable_, which is denoted by a
leading `$` character, followed by an alphanumeric string. For example:

```
{{ $w := "world" }}
Hello, {{ print $w }}!
Goodbye, {{ print $w }}.
```

this will render as:

```
Hello, world!
Goodbye, world.
```

## Indexing arrays and maps

Occasionally, multi-dimensional data such as arrays (lists, slices) and maps (dictionaries) are used in templates, sometimes through the use of
[data sources][]. Accessing values within these data can be done in a few
ways which bear clarifying:

Arrays are always numerically-indexed, and so can be accessed with a `[n]` suffix:
```
{{ $array[0] }}
```

You can also loop through an array with `range`:
```
{{ range $array }}
do something with {{ . }}...
{{ end }}
```

For maps, access is done with the `.` operator. Given a map `$map` with a key `foo`, you could access it like:

```
{{ $map.foo }}
```

However, if the key contains a non-alphanumeric character, you can use the `index`
function:

```
{{ index $map "foo-bar" }}
```

## Functions

Almost all of gomplate's utility is provided as _functions._ These are key
words (like `print` in the previous examples) that perform some action.

For example, the [`base64.Encode`][] function will encode some input string as
a base-64 string:

```
The word is {{ base64.Encode "swordfish" }}
```
renders as:
```
The word is c3dvcmRmaXNo
```

The [Go text/template]() language provides a number of built-in functions and 
operators that can be used in templates.

Here is a list of the built-in functions, but see [the documentation](https://golang.org/pkg/text/template/#hdr-Functions)
for full details:

- `and`
- `call`
- `html`
- `index`
- `js`
- `len`
- `not`
- `or`
- `print`
- `printf`
- `println`
- `urlquery`

And the following comparison operators are also supported:

- `eq`
- `ne`
- `lt`
- `le`
- `gt`
- `ge`

See also gomplate's functions, defined to the left.

## The Context

Go templates are always executed with a _context_. You can reference the context
with the `.` (period) character, and you can set the context in a block with the
`with` keyword. Like so:

```
$ gomplate -i '{{ with "foo" }}The context is {{ . }}{{ end }}'
The context is foo
```

Templates rendered by gomplate always have a _default_ context. You can populate
the default context from data sources with the [`--context`/`c`](../usage/#context-c)
flag. The special context item [`.Env`](#env) is available for referencing the
system's environment variables.

## Nested templates

Gomplate supports nested templates, using Go's `template` action. These can be
defined in-line with the `define` action, or external data can be used with the
[`--template`/`-t`](../usage/#template-t) flag.

Note that nested templates do _not_ have access to gomplate's default
[context](#the-context) (though it can be explicitly provided to the `template`
action).

### In-line templates

To define a nested template in-line, you can use the `define` action.

```
{{ define "T1" -}}
Hello {{ . }}!
{{- end -}}

{{ template "T1" "World" }}
{{ template "T1" }}
{{ template "T1" "everybody" }}
```

This renders as:

```
Hello World!
Hello <no value>!
Hello everybody!
```

### External templates

To define a nested template from an external source such as a file, use the
[`--template`/`-t`](../usage/#template-t) flag.

_hello.t:_
```
Hello {{ . }}!
```

```
$ gomplate -t hello=hello.t -i '{{ template "hello" "World" }} {{ template "hello" .Env.USER }}"
Hello World! Hello hairyhenderson!
```

## `.Env`

You can easily access environment variables with `.Env`, but there's a catch:
if you try to reference an environment variable that doesn't exist, parsing
will fail and `gomplate` will exit with an error condition.

For example:

```console
$ gomplate -i 'the user is {{ .Env.USER }}'
the user is hairyhenderson
$ gomplate -i 'this will fail: {{ .Env.BOGUS }}'
this will fail: template: <arg>:1:23: executing "<arg>" at <.Env.BOGUS>: map has no entry for key "BOGUS"
```

Sometimes, this behaviour is desired; if the output is unusable without certain
strings, this is a sure way to know that variables are missing!

If you want different behaviour, try [`getenv`](../functions/env/#env-getenv).

[`text/template`]: https://golang.org/pkg/text/template/
[`base64.Encode`]: ./functions/base64#base64-encode
[data sources]: ./datasources/
