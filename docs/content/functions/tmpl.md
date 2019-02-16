---
title: template functions
menu:
  main:
    parent: functions
---

Functions for defining or executing templates.

## `tmpl.Exec`

Execute (render) the named template. This is equivalent to using the [`template`](https://golang.org/pkg/text/template/#hdr-Actions) action, except the result is returned as a string.

This allows for post-processing of templates.

### Usage
```go
tmpl.Exec name [context] 
```

```go
context | tmpl.Exec name  
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(required)_ The template's name. |
| `context` | _(optional)_ The context to use. |

### Examples

```console
$ gomplate -i '{{define "T1"}}hello, world!{{end}}{{ tmpl.Exec "T1" | strings.ToUpper }}'
HELLO, WORLD!
```
```console
$ gomplate -i '{{define "T1"}}hello, {{.}}{{end}}{{ tmpl.Exec "T1" "world!" | strings.Title }}'
Hello, World!
```

## `tmpl.Inline`

**Alias:** `tpl`

Render the given string as a template, just like a nested template.

If the template is given a name (see `name` argument below), it can be re-used later with the `template` keyword.

A context can be provided, otherwise the default gomplate context will be used.

### Usage
```go
tmpl.Inline [name] in [context] 
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(optional)_ The template's name. |
| `in` | _(required)_ The template to render, as a string |
| `context` | _(optional)_ The context to use when rendering - this becomes `.` inside the template. |

### Examples

```console
$ gomplate -i '{{ tmpl.Inline "{{print `hello world`}}" }}'
hello world
```
```console
$ gomplate -i '
{{ $tstring := "{{ print .value ` world` }}" }}
{{ $context := dict "value" "hello" }}
{{ tpl "T1" $tstring $context }}
{{ template "T1" (dict "value" "goodbye") }}
'
hello world
goodbye world
```
