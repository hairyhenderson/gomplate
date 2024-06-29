---
title: template functions
menu:
  main:
    parent: functions
---

Functions for defining or executing templates.

## `tmpl.Exec`

Execute (render) the named template. This is equivalent to using the [`template`](https://pkg.go.dev/text/template/#hdr-Actions) action, except the result is returned as a string.

This allows for post-processing of templates.

_Added in gomplate [v3.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.3.0)_
### Usage

```
tmpl.Exec name [context]
```
```
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

_Added in gomplate [v3.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.3.0)_
### Usage

```
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

## `tmpl.Path`

Output the path of the current template, if it came from a file. For
inline templates, this will be an empty string.

Note that if this function is called from a nested template, the path
of the main template will be returned instead.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
tmpl.Path
```


### Examples

_`subdir/input.tpl`:_
```
this template is in {{ tmpl.Path }}
```

```console
$ gomplate -f subdir/input.tpl
this template is in subdir/input.tpl
```

## `tmpl.PathDir`

Output the current template's directory. For inline templates, this will
be an empty string.

Note that if this function is called from a nested template, the path
of the main template will be used instead.

_Added in gomplate [v3.11.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.11.0)_
### Usage

```
tmpl.PathDir
```


### Examples

_`subdir/input.tpl`:_
```
this template is in {{ tmpl.Dir }}
```

```console
$ gomplate -f subdir/input.tpl
this template is in subdir
```
