---
title: Other functions
menu:
  main:
    parent: functions
---

Miscellaneous functions that aren't part of a specific namespace. Most of these are fairly special-purpose.

## `tpl`

Render the given string as a template, just like a nested template.

If the template is given a name (see `name` argument below), it can be re-used later with the `template` keyword.

A context can be provided, otherwise the default gomplate context will be used.

### Usage
```go
tpl [name] in [context] 
```

### Arguments

| name | description |
|------|-------------|
| `name` | _(optional)_ The template's name. |
| `in` | _(required)_ The template to render, as a string |
| `context` | _(optional)_ The context to use when rendering - this becomes `.` inside the template. |

### Examples

```console
$ gomplate -i '{{ tpl "{{print `hello world`}}" }}'
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
