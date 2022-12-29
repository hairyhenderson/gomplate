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

| name | description |
|------|-------------|
| `offset` | _(optional)_ offset from the start of the string. Must be `4` or greater for ellipses to be added. Defaults to `0` |
| `width` | _(required)_ the desired maximum final width of the string, including ellipses |
| `input` | _(required)_ the input string to abbreviate |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `substr` | _(required)_ the substring to search for |
| `input` | _(required)_ the input to search |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `prefix` | _(required)_ the prefix to search for |
| `input` | _(required)_ the input to search |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `suffix` | _(required)_ the suffix to search for |
| `input` | _(required)_ the input to search |

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

| name | description |
|------|-------------|
| `width` | _(optional)_ number of times to repeat the `indent` string. Default: `1` |
| `indent` | _(optional)_ the string to indent with. Default: `" "` |
| `input` | _(required)_ the string to indent |

### Examples

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
$ gomplate -i '{{ (coll.Slice "foo" "bar" "baz") | strings.Sort }}'
[bar baz foo]
```

## `strings.SkipLines`

Skips the given number of lines (each ending in a `\n`), returning the
remainder.

If `skip` is greater than the number of lines in `in`, an empty string is
returned.

### Usage

```go
strings.SkipLines skip in
```
```go
in | strings.SkipLines skip
```

### Arguments

| name | description |
|------|-------------|
| `skip` | _(required)_ the number of lines to skip - must be a positive number |
| `in` | _(required)_ the input string |

### Examples

```console
$ gomplate -i '{{ "foo\nbar\nbaz" | strings.SkipLines 2 }}'
baz
```
```console
$ gomplate -i '{{ strings.SkipLines 1 "foo\nbar\nbaz" }}'
bar
baz
```

## `strings.Split`

_Not to be confused with [`split`](#split), which is deprecated._

Slices `input` into the substrings separated by `separator`, returning a
slice of the substrings between those separators. If `input` does not
contain `separator` and `separator` is not empty, returns a single-element
slice whose only element is `input`.

If `separator` is empty, it will split after each UTF-8 sequence. If
both inputs are empty (i.e. `strings.Split "" ""`), it will return an
empty slice.

This is equivalent to [`strings.SplitN`](#strings-splitn) with a `count`
of `-1`.

Note that the delimiter is not included in the resulting elements.

### Usage

```go
strings.Split separator input
```
```go
input | strings.Split separator
```

### Arguments

| name | description |
|------|-------------|
| `separator` | _(required)_ the delimiter to split on, can be multiple characters |
| `input` | _(required)_ the input string |

### Examples

```console
$ gomplate -i '{{range ("Bart,Lisa,Maggie" | strings.Split ",") }}Hello, {{.}}
{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```
```console
$ gomplate -i '{{range strings.Split "," "One,Two,Three" }}{{.}}{{"\n"}}{{end}}'
One
Two
Three
```

## `strings.SplitN`

_Not to be confused with [`splitN`](#splitn), which is deprecated._

Slices `input` into the substrings separated by `separator`, returning a
slice of the substrings between those separators. If `input` does not
contain `separator` and `separator` is not empty, returns a single-element
slice whose only element is `input`.

The `count` determines the number of substrings to return:

* `count > 0`: at most `count` substrings; the last substring will be the
  unsplit remainder.
* `count == 0`: the result is nil (zero substrings)
* `count < 0`: all substrings

See [`strings.Split`](#strings-split) for more details.

### Usage

```go
strings.SplitN separator count input
```
```go
input | strings.SplitN separator count
```

### Arguments

| name | description |
|------|-------------|
| `separator` | _(required)_ the delimiter to split on, can be multiple characters |
| `count` | _(required)_ the maximum number of substrings to return |
| `input` | _(required)_ the input string |

### Examples

```console
$ gomplate -i '{{ range ("foo:bar:baz" | strings.SplitN ":" 2) }}{{.}}
{{end}}'
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

### Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the number of times to repeat the input |
| `input` | _(required)_ the input to repeat |

### Examples

```console
$ gomplate -i '{{ "hello " | strings.Repeat 5 }}'
hello hello hello hello hello
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

### Arguments

| name | description |
|------|-------------|
| `old` | _(required)_ the text to replace |
| `new` | _(required)_ the new text to replace with |
| `input` | _(required)_ the input to modify |

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

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input to "slugify" |

### Examples

```console
$ gomplate -i '{{ "Hello, world!" | strings.Slug }}'
hello-world
```
```console
$ echo 'Rock & Roll @ Cafe Wha?' | gomplate -d in=stdin: -i '{{ strings.Slug (include "in") }}'
rock-and-roll-at-cafe-wha
```

## `strings.ShellQuote`

**Alias:** `shellQuote`

Given a string, emits a version of that string that will evaluate to its literal data when expanded by any POSIX-compliant shell.

Given an array or slice, emit a single string which will evaluate to a series of shell words, one per item in that array or slice.

### Usage

```go
strings.ShellQuote in
```
```go
in | strings.ShellQuote
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input to quote |

### Examples

```console
$ gomplate -i "{{ coll.Slice \"one word\" \"foo='bar baz'\" | shellQuote }}"
'one word' 'foo='"'"'bar baz'"'"''
```
```console
$ gomplate -i "{{ strings.ShellQuote \"it's a banana\" }}"
'it'"'"'s a banana'
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

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `cutset` | _(required)_ the set of characters to cut |
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `prefix` | _(required)_ the prefix to trim |
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `suffix` | _(required)_ the suffix to trim |
| `input` | _(required)_ the input |

### Examples

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

### Arguments

| name | description |
|------|-------------|
| `length` | _(required)_ the maximum length of the output |
| `input` | _(required)_ the input |

### Examples

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

## `strings.RuneCount`

Return the number of _runes_ (Unicode code-points) contained within the
input. This is similar to the built-in `len` function, but `len` counts
the length in _bytes_. The length of an input containing multi-byte
code-points should therefore be measured with `strings.RuneCount`.

Inputs will first be converted to strings, and multiple inputs are
concatenated.

This wraps Go's [`utf8.RuneCountInString`](https://golang.org/pkg/unicode/utf8/#RuneCountInString)
function.

### Usage

```go
strings.RuneCount input
```
```go
input | strings.RuneCount
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input(s) to measure |

### Examples

```console
$ gomplate -i '{{ range (coll.Slice "\u03a9" "\u0030" "\u1430") }}{{ printf "%s is %d bytes and %d runes\n" . (len .) (strings.RuneCount .) }}{{ end }}'
Ω is 2 bytes and 1 runes
0 is 1 bytes and 1 runes
ᐰ is 3 bytes and 1 runes
```

## `contains` _(deprecated)_
**Deprecation Notice:** Use [`strings.Contains`](#strings-contains) instead

**See [`strings.Contains`](#strings-contains) for a pipeline-compatible version**

Contains reports whether the second string is contained within the first. Equivalent to
[strings.Contains](https://golang.org/pkg/strings#Contains)

### Usage

```go
contains input substring
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the string to search |
| `substring` | _(required)_ the string to search for |

### Examples

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

## `hasPrefix` _(deprecated)_
**Deprecation Notice:** Use [`strings.HasPrefix`](#strings-hasprefix) instead

**See [`strings.HasPrefix`](#strings-hasprefix) for a pipeline-compatible version**

Tests whether the string begins with a certain substring. Equivalent to
[strings.HasPrefix](https://golang.org/pkg/strings#HasPrefix)

### Usage

```go
hasPrefix input prefix
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the string to search |
| `prefix` | _(required)_ the prefix to search for |

### Examples

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

## `hasSuffix` _(deprecated)_
**Deprecation Notice:** Use [`strings.HasSuffix`](#strings-hassuffix) instead

**See [`strings.HasSuffix`](#strings-hassuffix) for a pipeline-compatible version**

Tests whether the string ends with a certain substring. Equivalent to
[strings.HasSuffix](https://golang.org/pkg/strings#HasSuffix)

### Usage

```go
hasSuffix input suffix
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input to search |
| `suffix` | _(required)_ the suffix to search for |

### Examples

_`input.tmpl`:_
```
{{.Env.URL}}{{if not (hasSuffix .Env.URL ":80")}}:80{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
http://example.com:80
```

## `split` _(deprecated)_
**Deprecation Notice:** Use [`strings.Split`](#strings-split) instead

**See [`strings.Split`](#strings-split) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. Equivalent to
[strings.Split](https://golang.org/pkg/strings#Split)

### Usage

```go
split input separator
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input string |
| `separator` | _(required)_ the string sequence to split |

### Examples

```console
$ gomplate -i '{{range split "Bart,Lisa,Maggie" ","}}Hello, {{.}}
{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `splitN` _(deprecated)_
**Deprecation Notice:** Use [`strings.SplitN`](#strings-splitn) instead

**See [`strings.SplitN`](#strings-splitn) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. The count determines
the number of substrings to return. Equivalent to [strings.SplitN](https://golang.org/pkg/strings#SplitN)

### Usage

```go
splitN input separator count
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input string |
| `separator` | _(required)_ the string sequence to split |
| `count` | _(required)_ the maximum number of substrings to return |

### Examples

```console
$ gomplate -i '{{ range splitN "foo:bar:baz" ":" 2 }}{{.}}
{{end}}'
foo
bar:baz
```

## `trim` _(deprecated)_
**Deprecation Notice:** Use [`strings.Trim`](#strings-trim) instead

**See [`strings.Trim`](#strings-trim) for a pipeline-compatible version**

Trims a string by removing the given characters from the beginning and end of
the string. Equivalent to [strings.Trim](https://golang.org/pkg/strings/#Trim)

### Usage

```go
trim input cutset
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |
| `cutset` | _(required)_ the set of characters to cut |

### Examples

_`input.tmpl`:_
```
Hello, {{trim .Env.FOO " "}}!
```

```console
$ FOO="  world " | gomplate < input.tmpl
Hello, world!
```
