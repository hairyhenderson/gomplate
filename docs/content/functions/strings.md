---
title: strings functions
menu:
  main:
    parent: functions
---

## `strings.Abbrev`

Abbreviates a string using `...` (ellipses). Takes an optional offset from the beginning of the string, and a maximum final width (including added ellipses).

_Also see [`strings.Trunc`](#strings-trunc)._

### Usage
```go
strings.Abbrev [offset] width input
```
```go
input | strings.Abbrev [offset] width
```

### Arguments

| name   | description |
|--------|-------|
| `offset` | _(optional)_ offset from the start of the string. Must be `4` or greater for ellipses to be added. Defaults to `0` |
| `width` | _(required)_ the desired maximum final width of the string, including ellipses |
| `input` | _(required)_ the input string to abbreviate |

### Example

```console
$ gomplate -i '{{ "foobarbazquxquux" | strings.Abbrev 9 }}'
foobar...
$ gomplate -i '{{ "foobarbazquxquux" | strings.Abbrev 6 9 }}'
...baz...
```

## `strings.Contains`

Reports whether a substring is contained within a string.

### Usage
```go
strings.Contains substr input
```
```go
input | strings.Contains substr
```

### Example

_`input.tmpl`:_
```
{{ if (.Env.FOO | strings.Contains "f") }}yes{{else}}no{{end}}
```

```console
$ FOO=foo gomplate < input.tmpl
yes
$ FOO=bar gomplate < input.tmpl
no
```

## `strings.HasPrefix`

Tests whether a string begins with a certain prefix.

### Usage
```go
strings.HasPrefix prefix input
```
```go
input | strings.HasPrefix prefix
```

#### Example

```console
$ URL=http://example.com gomplate -i '{{if .Env.URL | strings.HasPrefix "https"}}foo{{else}}bar{{end}}'
bar
$ URL=https://example.com gomplate -i '{{if .Env.URL | strings.HasPrefix "https"}}foo{{else}}bar{{end}}'
foo
```

## `strings.HasSuffix`

Tests whether a string ends with a certain suffix.

### Usage
```go
strings.HasSuffix suffix input
```
```go
input | strings.HasSuffix suffix
```

### Examples

_`input.tmpl`:_
```
{{.Env.URL}}{{if not (.Env.URL | strings.HasSuffix ":80")}}:80{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
http://example.com:80
```

## `strings.Indent`

**Alias:** `indent`

Indents a string. If the input string has multiple lines, each line will be indented.

### Usage
```go
strings.Indent [width] [indent] input
```
```go
input | strings.Indent [width] [indent]
```

### Arguments

| name   | description |
|--------|-------|
| `width` | _(optional)_ number of times to repeat the `indent` string. Default: `1` |
| `indent` | _(optional)_ the string to indent with. Default: `" "` |
| `input` | the string to indent |

### Example

This function can be especially useful when adding YAML snippets into other YAML documents, where indentation is important:

_`input.tmpl`:_
```
foo:
{{ `{"bar": {"baz": 2}}` | json | toYAML | strings.Indent "  " }}
{{- `{"qux": true}` | json | toYAML | strings.Indent 2 }}
  quux:
{{ `{"quuz": 42}` | json | toYAML | strings.Indent 2 "  " -}}
```

```console
$ gomplate -f input.tmpl
foo:
  bar:
    baz: 2
  qux: true

  quux: 
    quuz: 42
```

## `strings.Sort` _(deprecated)_
**Deprecation Notice:** Use [`coll.Sort`](../coll/#coll-sort) instead

Returns an alphanumerically-sorted copy of a given string list.

### Usage
```go
strings.Sort list 
```

```go
list | strings.Sort  
```

### Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ The list to sort |

### Examples

```console
$ gomplate -i '{{ (slice "foo" "bar" "baz") | strings.Sort }}'
[bar baz foo]
```

## `strings.Split`

Creates a slice by splitting a string on a given delimiter.

### Usage
```go
strings.Split separator input
```
```go
input | strings.Split separator
```

### Examples

```console
$ gomplate -i '{{range ("Bart,Lisa,Maggie" | strings.Split ",") }}Hello, {{.}}{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `strings.SplitN`

Creates a slice by splitting a string on a given delimiter. The count determines
the number of substrings to return.

### Usage
```go
strings.SplitN separator count input
```
```go
input | strings.SplitN separator count
```

#### Example

```console
$ gomplate -i '{{ range ("foo:bar:baz" | strings.SplitN ":" 2) }}{{.}}{{end}}'
foo
bar:baz
```

## `strings.Quote`

**Alias:** `quote`

Surrounds an input string with double-quote characters (`"`). If the input is not a string, converts first.

`"` characters in the input are first escaped with a `\` character.

This is a convenience function which is equivalent to:

```
{{ print "%q" "input string" }}
```

### Usage
```go
strings.Quote in 
```

```go
in | strings.Quote  
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input to quote |

### Examples

```console
$ gomplate -i '{{ "in" | quote }}'
"in"
$ gomplate -i '{{ strings.Quote 500 }}'
"500"
```

## `strings.Repeat`

Returns a new string consisting of `count` copies of the input string.

It errors if `count` is negative or if the length of `input` multiplied by `count` overflows.

This wraps Go's [`strings.Repeat`](https://golang.org/pkg/strings/#Repeat).

### Usage
```go
strings.Repeat count input
```
```go
input | strings.Repeat count
```

#### Example

```console
$ gomplate -i '{{ "hello, world" | strings.Repeat "world" }}jello'
hello, jello
```

## `strings.ReplaceAll`

**Alias:** `replaceAll`

Replaces all occurrences of a given string with another.

### Usage
```go
strings.ReplaceAll old new input
```
```go
input | strings.ReplaceAll old new
```

### Examples

```console
$ gomplate -i '{{ strings.ReplaceAll "." "-" "172.21.1.42" }}'
172-21-1-42
$ gomplate -i '{{ "172.21.1.42" | strings.ReplaceAll "." "-" }}'
172-21-1-42
```

## `strings.Slug`

Creates a a "slug" from a given string - supports Unicode correctly. This wraps the [github.com/gosimple/slug](https://github.com/gosimple/slug) package. See [the github.com/gosimple/slug docs](https://godoc.org/github.com/gosimple/slug) for more information.

### Usage
```go
strings.Slug input
```
```go
input | strings.Slug
```

### Examples
```console
$ gomplate -i '{{ "Hello, world!" | strings.Slug }}'
hello-world

$ echo 'Rock & Roll @ Cafe Wha?' | gomplate -d in=stdin: -i '{{ strings.Slug (include "in") }}'
rock-and-roll-at-cafe-wha
```

## `strings.Squote`

**Alias:** `squote`

Surrounds an input string with a single-quote (apostrophe) character (`'`). If the input is not a string, converts first.

`'` characters in the input are first escaped in the YAML-style (by repetition: `''`).

### Usage
```go
strings.Squote in 
```

```go
in | strings.Squote  
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input to quote |

### Examples

```console
$ gomplate -i '{{ "in" | squote }}'
'in'
```
```console
$ gomplate -i "{{ strings.Squote \"it's a banana\" }}"
'it''s a banana'
```

## `strings.Title`

**Alias:** `title`

Convert to title-case.

### Usage
```go
strings.Title input
```
```go
input | strings.Title
```

### Example

```console
$ gomplate -i '{{strings.Title "hello, world!"}}'
Hello, World!
```

## `strings.ToLower`

**Alias:** `toLower`

Convert to lower-case.

### Usage
```go
strings.ToLower input
```
```go
input | strings.ToLower
```

#### Example

```console
$ echo '{{strings.ToLower "HELLO, WORLD!"}}' | gomplate
hello, world!
```

## `strings.ToUpper`

**Alias:** `toUpper`

Convert to upper-case.

### Usage
```go
strings.ToUpper input
```
```go
input | strings.ToUpper
```

#### Example

```console
$ gomplate -i '{{strings.ToUpper "hello, world!"}}'
HELLO, WORLD!
```

## `strings.Trim`

Trims a string by removing the given characters from the beginning and end of
the string.

### Usage
```go
strings.Trim cutset input
```
```go
input | strings.Trim cutset
```

#### Example

```console
$ gomplate -i '{{ "_-foo-_" | strings.Trim "_-" }}
foo
```

## `strings.TrimPrefix`

Returns a string without the provided leading prefix string, if the prefix is present.

This wraps Go's [`strings.TrimPrefix`](https://golang.org/pkg/strings/#TrimPrefix).

### Usage
```go
strings.TrimPrefix prefix input
```
```go
input | strings.TrimPrefix prefix
```

#### Example

```console
$ gomplate -i '{{ "hello, world" | strings.TrimPrefix "hello, " }}'
world
```

## `strings.TrimSpace`

**Alias:** `trimSpace`

Trims a string by removing whitespace from the beginning and end of
the string.

### Usage
```go
strings.TrimSpace input
```
```go
input | strings.TrimSpace
```

#### Example

```console
$ gomplate -i '{{ "  \n\t foo" | strings.TrimSpace }}'
foo
```


## `strings.TrimSuffix`

Returns a string without the provided trailing suffix string, if the suffix is present.

This wraps Go's [`strings.TrimSuffix`](https://golang.org/pkg/strings/#TrimSuffix).

### Usage
```go
strings.TrimSuffix suffix input
```
```go
input | strings.TrimSuffix suffix
```

#### Example

```console
$ gomplate -i '{{ "hello, world" | strings.TrimSuffix "world" }}jello'
hello, jello
```

## `strings.Trunc`

Returns a string truncated to the given length.

_Also see [`strings.Abbrev`](#strings-abbrev)._

### Usage
```go
strings.Trunc length input
```
```go
input | strings.Trunc length
```

#### Example

```console
$ gomplate -i '{{ "hello, world" | strings.Trunc 5 }}'
hello
```

## `strings.CamelCase`

Converts a sentence to CamelCase, i.e. `The quick brown fox` becomes `TheQuickBrownFox`.

All non-alphanumeric characters are stripped, and the beginnings of words are upper-cased. If the input begins with a lower-case letter, the result will also begin with a lower-case letter.

See [CamelCase on Wikipedia](https://en.wikipedia.org/wiki/Camel_case) for more details.

### Usage
```go
strings.CamelCase in 
```

```go
in | strings.CamelCase  
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input |

### Examples

```console
$ gomplate -i '{{ "Hello, World!" | strings.CamelCase }}'
HelloWorld
```
```console
$ gomplate -i '{{ "hello jello" | strings.CamelCase }}'
helloJello
```

## `strings.SnakeCase`

Converts a sentence to snake_case, i.e. `The quick brown fox` becomes `The_quick_brown_fox`.

All non-alphanumeric characters are stripped, and spaces are replaced with an underscore (`_`). If the input begins with a lower-case letter, the result will also begin with a lower-case letter.

See [Snake Case on Wikipedia](https://en.wikipedia.org/wiki/Snake_case) for more details.

### Usage
```go
strings.SnakeCase in 
```

```go
in | strings.SnakeCase  
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input |

### Examples

```console
$ gomplate -i '{{ "Hello, World!" | strings.SnakeCase }}'
Hello_world
```
```console
$ gomplate -i '{{ "hello jello" | strings.SnakeCase }}'
hello_jello
```

## `strings.KebabCase`

Converts a sentence to kebab-case, i.e. `The quick brown fox` becomes `The-quick-brown-fox`.

All non-alphanumeric characters are stripped, and spaces are replaced with a hyphen (`-`). If the input begins with a lower-case letter, the result will also begin with a lower-case letter.

See [Kebab Case on Wikipedia](https://en.wikipedia.org/wiki/Kebab_case) for more details.

### Usage
```go
strings.KebabCase in 
```

```go
in | strings.KebabCase  
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input |

### Examples

```console
$ gomplate -i '{{ "Hello, World!" | strings.KebabCase }}'
Hello-world
```
```console
$ gomplate -i '{{ "hello jello" | strings.KebabCase }}'
hello-jello
```

## `strings.WordWrap`

Inserts new line breaks into the input string so it ends up with lines that are at most `width` characters wide.

The line-breaking algorithm is _naïve_ and _greedy_: lines are only broken between words (i.e. on whitespace characters), and no effort is made to "smooth" the line endings.

When words that are longer than the desired width are encountered (e.g. long URLs), they are not broken up. Correctness is valued above line length.

The line-break sequence defaults to `\n` (i.e. the LF/Line Feed character), regardless of OS. 

### Usage
```go
strings.WordWrap [width] [lbseq] in 
```

```go
in | strings.WordWrap [width] [lbseq]  
```

### Arguments

| name | description |
|------|-------------|
| `width` | _(optional)_ The desired maximum line length (number of characters - defaults to `80`) |
| `lbseq` | _(optional)_ The line-break sequence to use (defaults to `\n`) |
| `in` | _(required)_ The input |

### Examples

```console
$ gomplate -i '{{ "Hello, World!" | strings.WordWrap 7 }}'
Hello,
World!
```
```console
$ gomplate -i '{{ strings.WordWrap 20 "\\\n" "a string with a long url http://example.com/a/very/long/url which should not be broken" }}'
a string with a long
url
http://example.com/a/very/long/url
which should not be
broken
```

## `contains`

**See [`strings.Contains`](#strings-contains) for a pipeline-compatible version**

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

## `hasPrefix`

**See [`strings.HasPrefix`](#strings-hasprefix) for a pipeline-compatible version**

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

**See [`strings.HasSuffix`](#strings-hassuffix) for a pipeline-compatible version**

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

## `split`

**See [`strings.Split`](#strings-split) for a pipeline-compatible version**

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

**See [`strings.SplitN`](#strings-splitn) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. The count determines
the number of substrings to return. Equivalent to [strings.SplitN](https://golang.org/pkg/strings#SplitN)

#### Example

```console
$ gomplate -i '{{ range splitN "foo:bar:baz" ":" 2 }}{{.}}{{end}}'
foo
bar:baz
```

## `trim`

**See [`strings.Trim`](#strings-trim) for a pipeline-compatible version**

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
