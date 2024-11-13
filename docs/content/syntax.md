---
title: Syntax
weight: 13
menu: main
---

Gomplate uses the syntax understood by the Go language's [`text/template`][]
package. This page documents some of that syntax, but see [the language docs][`text/template`]
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
{{ range coll.Slice "Foo" "bar" "baz" }}
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

You can use [Golang template syntax](https://pkg.go.dev/text/template/#hdr-Text_and_spaces) to fix this. Leading newlines (i.e. newlines that come before the action) can be suppressed by placing a minus sign in front of the first set of delimiters (`{{`). Putting the minus sign behind the trailing set of delimiters (`}}`) will suppress the newline _after_ the action. You can do both to suppress newlines entirely on that line.

Placing the minus sign within the context (i.e. inside of `{{.}}`) has no effect.

Here are a few examples.

### Suppressing leading newlines

```
{{- range coll.Slice "Foo" "bar" "baz" }}
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
{{ range coll.Slice "Foo" "bar" "baz" -}}
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
{{- range coll.Slice "Foo" "bar" "baz" -}}
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

Variables are declared with `:=`, and can be redefined with `=`:

```
{{ $w := "hello" }}
{{ $w = "goodbye" }}
```

### Variable scope

A variable's scope extends to the `end` action of the control structure (`if`,
`with`, or `range`) in which it is declared, or to the end of the template if
there is no such control structure.

In other words, if a variable is initialized inside an `if` or `else` block,
it cannot be referenced outside that block.

This template will error with `undefined variable "$w"` since `$w` is only
declared within `if`/`else` blocks:

```
{{ if 1 }}
{{ $w := "world" }}
{{ else }}
{{ $w := "earth" }}
{{ end }}

Hello, {{ print $w }}!
Goodbye, {{ print $w }}.
```

One way to approach this is to declare the variable first to an empty value:

```
{{ $w := "" }}
{{ if 1 }}
{{ $w = "world" }}
{{ else }}
{{ $w = "earth" }}
{{ end -}}

Hello, {{ print $w }}!
Goodbye, {{ print $w }}.
```

## Indexing arrays and maps

Occasionally, multi-dimensional data such as arrays (lists, slices) and maps
(dictionaries) are used in templates, sometimes through the use of
[data sources][]. Accessing values within these data can be done in a few ways
which bear clarifying.

### Arrays

Arrays are always numerically-indexed, and individual values can be accessed with the `index` built-in function:

```
{{ index $array 0 }}
```

To visit each value, you can loop through an array with `range`:

```
{{ range $array }}
do something with {{ . }}...
{{ end }}
```

If you need to keep track of the index number, you can declare two variables, separated by a comma:

```
{{ range $index, $element := $array }}
do something with {{ $element }}, which is number {{ $index }}
{{ end }}
```

### Maps

For maps, accessing values can be done with the `.` operator. Given a map `$map`
with a key `foo`, you could access it like:

```
{{ $map.foo }}
```

However, this kind of access is limited to keys which are strings and contain
only characters in the set (`a`-`z`,`A`-`Z`,`_`,`1`-`9`), and which do not begin
with a number. If the key doesn't conform to these rules, you can use the `index`
built-in function instead:

```
{{ index $map "foo-bar" }}
```

`index` also supports nested keys and can be combined with other functions as such:

```
{{ index $map "foo" (env.Getenv "BAR") "baz" ... }}
``` 

**Note:** _while `index` can be used to access awkwardly-named values in maps,
it behaves differently than the `.` operator. If the key doesn't exist, `index`
will simply not return a value, while `.` will error._

And, similar to arrays, you can loop through a map with the `range`:

```
{{ range $map }}
The value is {{ . }}
{{ end }}
```

Or if you need keys as well:

```
{{ range $key, $value := $map }}
{{ $key }}'s value is: {{ $value }}
{{ end }}
```

## Functions

Almost all of gomplate's utility is provided as _functions._ These are key
words (like `print` in the previous examples) that perform some action.

See the [functions documentation](/functions/) for more information.

## The Context

Go templates are always executed with a _context_. You can reference the context
with the `.` (period) character, and you can set the context in a block with the
`with` action. Like so:

```
$ gomplate -i '{{ with "foo" }}The context is {{ . }}{{ end }}'
The context is foo
```

Templates rendered by gomplate always have a _default_ context. You can populate
the default context from data sources with the [`--context`/`c`](../usage/#--context-c)
flag. The special context item [`.Env`](#env) is available for referencing the
system's environment variables.

_Note:_ The initial context (`.`) is always available as the variable `$`,
so the initial context is always available, even when shadowed with `range`
or `with` blocks:

```
$ echo '{"bar":"baz"}' | gomplate -c .=stdin:///in.json -i 'context is: {{ . }}
{{ with "foo" }}now context is {{ . }}
but the original context is still {{ $ }}
{{ end }}'
context is: map[bar:baz]
now context is foo
but the original context is still map[bar:baz]
```

## Nested templates

Gomplate supports nested templates, using Go's `template` action. These can be
defined in-line with the `define` action, or external data can be used with the
[`--template`/`-t`](../usage/#--template-t) flag.

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
[`--template`/`-t`](../usage/#--template-t) flag.

_hello.t:_
```
Hello {{ . }}!
```

```
$ gomplate -t hello=hello.t -i '{{ template "hello" "World" }} {{ template "hello" .Env.USER }}'
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

If you want different behaviour, try [`getenv`](../functions/env/#envgetenv).

[`text/template`]: https://pkg.go.dev/text/template/
[`base64.Encode`]: ../functions/base64#base64-encode
[data sources]: ../datasources/
