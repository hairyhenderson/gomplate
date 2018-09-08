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

Occasionally, multi-dimensional data such as arrays (lists, slices) and maps (dictionaries) are used in templates, somtimes through the use of
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

Templates rendered by gomplate always have a _default_ context. In future, gomplate's
context may expand (_watch this space!_), but currently, it contains one item: the
system's environment variables, available as [`.Env`](#env).

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
