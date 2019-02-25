---
title: regexp functions
menu:
  main:
    parent: functions
---

These functions allow user you to search and modify text with regular expressions.

The syntax of the regular expressions accepted is [Go's `regexp` syntax](https://golang.org/pkg/regexp/syntax/#hdr-Syntax),
and is the same general syntax used by Perl, Python, and other languages.

## `regexp.Find`

Returns a string holding the text of the leftmost match in `input`
of the regular expression `expression`.

This function provides the same behaviour as Go's
[`regexp.FindString`](https://golang.org/pkg/regexp/#Regexp.FindString) function.

### Usage

```go
regexp.Find expression input
```
```go
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
[`regexp.FindAllString`](https://golang.org/pkg/regexp/#Regexp.FindAllString) function.

### Usage

```go
regexp.FindAll expression [false] input
```
```go
input | regexp.FindAll expression [false]
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `false` | _(optional)_ The number of matches to return |
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

### Usage

```go
regexp.Match expression input
```
```go
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

## `regexp.Replace`

Replaces matches of a regular expression with the replacement string.

The replacement is substituted after expanding variables beginning with `$`.

This function provides the same behaviour as Go's
[`regexp.ReplaceAllString`](https://golang.org/pkg/regexp/#Regexp.ReplaceAllString) function.

### Usage

```go
regexp.Replace expression replacement input
```
```go
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
[`regexp.ReplaceAllLiteralString`](https://golang.org/pkg/regexp/#Regexp.ReplaceAllLiteralString) function.

### Usage

```go
regexp.ReplaceLiteral expression replacement input
```
```go
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

This is equivalent to [`strings.SplitN`](../strings/#strings-splitn),
except that regular expressions are supported.

This function provides the same behaviour as Go's
[`regexp.Split`](https://golang.org/pkg/regexp/#Regexp.Split) function.

### Usage

```go
regexp.Split expression [false] input
```
```go
input | regexp.Split expression [false]
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `false` | _(optional)_ The number of matches to return |
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
