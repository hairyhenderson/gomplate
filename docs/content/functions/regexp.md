---
title: regexp functions
menu:
  main:
    parent: functions
---

These functions allow user you to search and modify text with regular expressions.

The syntax of the regular expressions accepted is [Go's `regexp` syntax](https://pkg.go.dev/regexp/syntax/#hdr-Syntax),
and is the same general syntax used by Perl, Python, and other languages.

## `regexp.Find`

Returns a string holding the text of the leftmost match in `input`
of the regular expression `expression`.

This function provides the same behaviour as Go's
[`regexp.FindString`](https://pkg.go.dev/regexp/#Regexp.FindString) function.

_Added in gomplate [v3.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.1.0)_
### Usage

```
regexp.Find expression input
```
```
input | regexp.Find expression
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `input` | _(required)_ The input to search |

### Examples

```console
$ gomplate -i '{{ regexp.Find "[a-z]{3}" "foobar"}}'
foo
```
```console
$ gomplate -i 'no {{ "will not match" | regexp.Find "[0-9]" }}numbers'
no numbers
```

## `regexp.FindAll`

Returns a list of all successive matches of the regular expression.

This can be called with 2 or 3 arguments. When called with 2 arguments, the
`n` argument (number of matches) will be set to `-1`, causing all matches
to be returned.

This function provides the same behaviour as Go's
[`regexp.FindAllString`](https://pkg.go.dev/regexp#Regexp.FindAllString) function.

_Added in gomplate [v3.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.1.0)_
### Usage

```
regexp.FindAll expression [n] input
```
```
input | regexp.FindAll expression [n]
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `n` | _(optional)_ The number of matches to return |
| `input` | _(required)_ The input to search |

### Examples

```console
$ gomplate -i '{{ regexp.FindAll "[a-z]{3}" "foobar" | toJSON}}'
["foo", "bar"]
```
```console
$ gomplate -i '{{ "foo bar baz qux" | regexp.FindAll "[a-z]{3}" 3 | toJSON}}'
["foo", "bar", "baz"]
```

## `regexp.Match`

Returns `true` if a given regular expression matches a given input.

This returns a boolean which can be used in an `if` condition, for example.

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
regexp.Match expression input
```
```
input | regexp.Match expression
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `input` | _(required)_ The input to test |

### Examples

```console
$ gomplate -i '{{ if (.Env.USER | regexp.Match `^h`) }}username ({{.Env.USER}}) starts with h!{{end}}'
username (hairyhenderson) starts with h!
```

## `regexp.QuoteMeta`

Escapes all regular expression metacharacters in the input. The returned string is a regular expression matching the literal text.

This function provides the same behaviour as Go's
[`regexp.QuoteMeta`](https://pkg.go.dev/regexp#QuoteMeta) function.

_Added in gomplate [v3.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.7.0)_
### Usage

```
regexp.QuoteMeta input
```
```
input | regexp.QuoteMeta
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input to escape |

### Examples

```console
$ gomplate -i '{{ `{hello}` | regexp.QuoteMeta }}'
\{hello\}
```

## `regexp.Replace`

Replaces matches of a regular expression with the replacement string.

The replacement is substituted after expanding variables beginning with `$`.

This function provides the same behaviour as Go's
[`regexp.ReplaceAllString`](https://pkg.go.dev/regexp/#Regexp.ReplaceAllString) function.

_Added in gomplate [v1.9.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.9.0)_
### Usage

```
regexp.Replace expression replacement input
```
```
input | regexp.Replace expression replacement
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression string |
| `replacement` | _(required)_ The replacement string |
| `input` | _(required)_ The input string to operate on |

### Examples

```console
$ gomplate -i '{{ regexp.Replace "(foo)bar" "$1" "foobar"}}'
foo
```
```console
$ gomplate -i '{{ regexp.Replace "(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)" "${last}, ${first}" "Alan Turing"}}'
Turing, Alan
```

## `regexp.ReplaceLiteral`

Replaces matches of a regular expression with the replacement string.

The replacement is substituted directly, without expanding variables
beginning with `$`.

This function provides the same behaviour as Go's
[`regexp.ReplaceAllLiteralString`](https://pkg.go.dev/regexp/#Regexp.ReplaceAllLiteralString) function.

_Added in gomplate [v3.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.1.0)_
### Usage

```
regexp.ReplaceLiteral expression replacement input
```
```
input | regexp.ReplaceLiteral expression replacement
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression string |
| `replacement` | _(required)_ The replacement string |
| `input` | _(required)_ The input string to operate on |

### Examples

```console
$ gomplate -i '{{ regexp.ReplaceLiteral "(foo)bar" "$1" "foobar"}}'
$1
```
```console
$ gomplate -i '{{ `foo.bar,baz` | regexp.ReplaceLiteral `\W` `$` }}'
foo$bar$baz
```

## `regexp.Split`

Splits `input` into sub-strings, separated by the expression.

This can be called with 2 or 3 arguments. When called with 2 arguments, the
`n` argument (number of matches) will be set to `-1`, causing all sub-strings
to be returned.

This is equivalent to [`strings.SplitN`](../strings/#stringssplitn),
except that regular expressions are supported.

This function provides the same behaviour as Go's
[`regexp.Split`](https://pkg.go.dev/regexp/#Regexp.Split) function.

_Added in gomplate [v3.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.1.0)_
### Usage

```
regexp.Split expression [n] input
```
```
input | regexp.Split expression [n]
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `n` | _(optional)_ The number of matches to return |
| `input` | _(required)_ The input to search |

### Examples

```console
$ gomplate -i '{{ regexp.Split `[\s,.]` "foo bar,baz.qux" | toJSON}}'
["foo","bar","baz","qux"]
```
```console
$ gomplate -i '{{ "foo bar.baz,qux" | regexp.Split `[\s,.]` 3 | toJSON}}'
["foo","bar","baz"]
```
