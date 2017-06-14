---
title: regexp functions
menu:
  main:
    parent: functions
---

## `regexp.Replace`

Replaces matches of a regular expression with the replacement string. The syntax
of the regular expressions accepted is [Go's `regexp` syntax](https://golang.org/pkg/regexp/syntax/#hdr-Syntax),
and is the same general syntax used by Perl, Python, and other languages.

### Usage

```go
regexp.Replace expression replacement input
```
```go
input | regexp.Replace expression replacement
```

### Arguments

| name   | description |
|--------|-------|
| `expression` | The regular expression string |
| `replacement` | The replacement string |
| `input` | the input string to operate on |

### Examples

```console
$ gomplate -i '{{ regexp.Replace "(foo)bar" "$1" "foobar"}}'
foo
```

```console
$ gomplate -i '{{ regexp.Replace "(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)" "${last}, ${first}" "Alan Turing"}}'
Turing, Alan
```

## `regexp.Match`

Returns `true` if a given regular expression matches a given input string.

This returns a boolean which can be used in an `if` condition, for example.

### Usage

```go
regexp.Match expression input
```
```go
input | regexp.Match expression
```

### Arguments

| name   | description |
|--------|-------|
| `expression` | the regular expression to match |
| `input` | the input string to test |

### Examples

```console
$ gomplate -i '{{ if (.Env.USER | regexp.Match `^h`) }}username ({{.Env.USER}}) starts with h!{{end}}'
username (hairyhenderson) starts with h!
```
